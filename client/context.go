package client

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/sys/unix"
)

// serverIDStart is the first object id allocated by the server
// (WL_SERVER_ID_START); ids below it are allocated by the client.
const serverIDStart = 0xff000000

// FdEventer is implemented by generated proxies whose events carry a file
// descriptor. Bit n of FdEvents is set when the event with opcode n has an fd
// argument; the Context uses it to hand each received fd to the event that
// owns it, as libwayland does, instead of the event read with it.
type FdEventer interface {
	FdEvents() uint64
}

// Context is a Wayland connection and its object table. Registering,
// looking up and writing are safe for concurrent use; reading
// ([Context.GetDispatch], [Context.ReadMsg]) must happen on one goroutine at a
// time. Requests that create objects must be serialized by caller: the id is
// allocated and sent under separate locks, and the server rejects ids that
// arrive out of order.
type Context struct {
	conn *net.UnixConn

	mu        sync.Mutex
	objects   map[uint32]Proxy
	zombies   map[uint32]uint64 // destroyed ids awaiting delete_id -> fd event mask
	deleted   map[uint32]bool   // ids the server deleted while still registered
	free      []uint32          // client ids released by delete_id, reused LIFO
	currentID uint32
	fds       []int // received fds not yet taken by an event

	wmu sync.Mutex // serializes WriteMsg

	header [8]byte
	oob    []byte
}

func newContext(conn *net.UnixConn) *Context {
	return &Context{
		conn:    conn,
		objects: map[uint32]Proxy{},
		zombies: map[uint32]uint64{},
		deleted: map[uint32]bool{},
		oob:     make([]byte, oobSpace),
	}
}

// Register allocates a client object id for p, reusing ids the server has
// released with wl_display.delete_id.
func (ctx *Context) Register(p Proxy) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	var id uint32
	if n := len(ctx.free); n > 0 {
		id = ctx.free[n-1]
		ctx.free = ctx.free[:n-1]
	} else {
		ctx.currentID++
		id = ctx.currentID
	}
	p.SetID(id)
	p.SetContext(ctx)
	ctx.objects[id] = p
}

// SetProxy registers p under an id chosen by the server (a new_id event
// argument).
func (ctx *Context) SetProxy(id uint32, p Proxy) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	p.SetID(id)
	p.SetContext(ctx)
	delete(ctx.zombies, id)
	ctx.objects[id] = p
}

// Unregister removes p from the object table. Until the server confirms the
// deletion, events still addressed to p are dropped (closing their fds), not
// reported as errors.
func (ctx *Context) Unregister(p Proxy) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	id := p.ID()
	if ctx.objects[id] != p {
		return
	}
	delete(ctx.objects, id)
	if ctx.deleted[id] {
		delete(ctx.deleted, id)
		ctx.release(id)
		return
	}
	var mask uint64
	if fe, ok := p.(FdEventer); ok {
		mask = fe.FdEvents()
	}
	ctx.zombies[id] = mask
}

// deleteID handles wl_display.delete_id: the server no longer references id.
func (ctx *Context) deleteID(id uint32) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if _, ok := ctx.zombies[id]; ok {
		delete(ctx.zombies, id)
		ctx.release(id)
	} else if _, ok := ctx.objects[id]; ok {
		ctx.deleted[id] = true
	}
}

// release makes a client id available again; ctx.mu must be held.
func (ctx *Context) release(id uint32) {
	if id < serverIDStart {
		ctx.free = append(ctx.free, id)
	}
}

func (ctx *Context) GetProxy(id uint32) Proxy {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	return ctx.objects[id]
}

// Close closes the connection and any received fds no event has taken.
func (ctx *Context) Close() error {
	ctx.mu.Lock()
	for _, fd := range ctx.fds {
		unix.Close(fd)
	}
	ctx.fds = nil
	ctx.mu.Unlock()
	return ctx.conn.Close()
}

// Dispatch reads and processes incoming messages and calls [client.Dispatcher.Dispatch] on the
// respective wayland protocol.
// Dispatch must be called on same goroutine as other interactions with Context.
// If multi goroutine approach is desired, use [Context.GetDispatch] instead.
// Dispatch blocks if there are no incoming messages.
// A Dispatch loop is usually used to handle incoming messages.
func (ctx *Context) Dispatch() error {
	return ctx.GetDispatch()()
}

var ErrDispatchSenderNotFound = errors.New("dispatch: unable to find sender")
var ErrDispatchSenderUnsupported = errors.New("dispatch: sender does not implement Dispatch method")
var ErrDispatchUnableToReadMsg = errors.New("dispatch: unable to read msg")

// GetDispatch reads incoming messages and returns dispatch function which calls
// [client.Dispatcher.Dispatch] on respective wayland protocol.
// While GetDispatch is usually called in loop in separate goroutine, dispatch function it
// returns must be called in same goroutine as other interactions with Context.
// GetDispatch blocks if there are no incoming messages.
func (ctx *Context) GetDispatch() func() error {
	senderID, opcode, nfds, data, err := ctx.readMsg() // Blocks if there are no incoming messages
	if err != nil {
		return func() error {
			return fmt.Errorf("%w: %w", ErrDispatchUnableToReadMsg, err)
		}
	}

	return func() error {
		return ctx.dispatch(senderID, opcode, nfds, data)
	}
}

// dispatch delivers one event. nfds is the number of fds read along with it,
// used only for proxies that do not implement FdEventer.
func (ctx *Context) dispatch(senderID, opcode uint32, nfds int, data []byte) error {
	ctx.mu.Lock()
	sender, ok := ctx.objects[senderID]
	mask, zombie := ctx.zombies[senderID]
	ctx.mu.Unlock()

	if senderID == 1 && opcode == 1 && len(data) >= 4 { // wl_display.delete_id
		ctx.deleteID(Uint32(data))
	}
	if !ok {
		if zombie {
			if mask&(1<<opcode) != 0 {
				if fd := ctx.popFd(); fd != -1 {
					unix.Close(fd)
				}
			}
			return nil
		}
		return fmt.Errorf("%w (senderID=%d)", ErrDispatchSenderNotFound, senderID)
	}
	d, ok := sender.(Dispatcher)
	if !ok {
		return fmt.Errorf("%w (senderID=%d)", ErrDispatchSenderUnsupported, senderID)
	}

	fd := -1
	if fe, ok := sender.(FdEventer); ok {
		if fe.FdEvents()&(1<<opcode) != 0 {
			fd = ctx.popFd()
		}
	} else if nfds > 0 {
		fd = ctx.popFd()
	}
	d.Dispatch(opcode, fd, data)
	return nil
}

// popFd returns the oldest received fd, or -1 if none is queued.
func (ctx *Context) popFd() int {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if len(ctx.fds) == 0 {
		return -1
	}
	fd := ctx.fds[0]
	ctx.fds = ctx.fds[1:]
	return fd
}

// Connect connects to the Wayland compositor. An empty addr uses
// $WAYLAND_DISPLAY (default "wayland-0"), relative to $XDG_RUNTIME_DIR unless
// it is an absolute path; a non-empty addr is used as the socket path.
func Connect(addr string) (*Display, error) {
	if addr == "" {
		addr = os.Getenv("WAYLAND_DISPLAY")
		if addr == "" {
			addr = "wayland-0"
		}
		if !filepath.IsAbs(addr) {
			runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
			if runtimeDir == "" {
				return nil, errors.New("env XDG_RUNTIME_DIR not set")
			}
			addr = filepath.Join(runtimeDir, addr)
		}
	}

	conn, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: addr, Net: "unix"})
	if err != nil {
		return nil, err
	}

	return NewDisplay(newContext(conn)), nil
}
