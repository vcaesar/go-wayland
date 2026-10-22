package client

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// pair returns a Context and the compositor end of a connected socket pair.
func pair(t *testing.T) (*Context, *net.UnixConn) {
	t.Helper()
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	if err != nil {
		t.Fatal(err)
	}
	conn := func(fd int) *net.UnixConn {
		f := os.NewFile(uintptr(fd), "wl")
		defer f.Close()
		c, err := net.FileConn(f)
		if err != nil {
			t.Fatal(err)
		}
		return c.(*net.UnixConn)
	}
	ctx := newContext(conn(fds[0]))
	srv := conn(fds[1])
	t.Cleanup(func() {
		ctx.Close()
		srv.Close()
	})
	return ctx, srv
}

func u32(v uint32) []byte {
	b := make([]byte, 4)
	PutUint32(b, v)
	return b
}

func str(s string) []byte {
	b := make([]byte, 4+PaddedLen(len(s)+1))
	PutString(b, s)
	return b
}

// event encodes one wire message.
func event(id, opcode uint32, args ...[]byte) []byte {
	var body []byte
	for _, a := range args {
		body = append(body, a...)
	}
	return append(append(u32(id), u32(uint32(8+len(body))<<16|opcode)...), body...)
}

func send(t *testing.T, srv *net.UnixConn, b []byte, fds ...int) {
	t.Helper()
	var oob []byte
	if len(fds) > 0 {
		oob = unix.UnixRights(fds...)
	}
	if _, _, err := srv.WriteMsgUnix(b, oob, nil); err != nil {
		t.Fatal(err)
	}
}

func dispatch(t *testing.T, ctx *Context) {
	t.Helper()
	if err := ctx.Dispatch(); err != nil {
		t.Fatal(err)
	}
}

// pipeFd returns the read end of a pipe holding s; the caller closes it.
func pipeFd(t *testing.T, s string) (*os.File, *os.File) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.WriteString(s); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { r.Close(); w.Close() })
	return r, w
}

// An fd sent along with an earlier event must reach the event that carries
// it (wl_keyboard.keymap), not the event it was read with.
func TestFdGoesToOwningEvent(t *testing.T) {
	ctx, srv := pair(t)
	ptr, kb := &Pointer{}, &Keyboard{}
	ctx.Register(ptr)
	ctx.Register(kb)

	frames := 0
	ptr.SetFrameHandler(func(PointerFrameEvent) { frames++ })
	var km KeyboardKeymapEvent
	kb.SetKeymapHandler(func(e KeyboardKeymapEvent) { km = e })

	r, _ := pipeFd(t, "keymap")
	msgs := append(event(ptr.ID(), 5), event(kb.ID(), 0, u32(1), u32(6))...)
	send(t, srv, msgs, int(r.Fd()))

	dispatch(t, ctx)
	dispatch(t, ctx)
	if frames != 1 || km.Format != 1 || km.Size != 6 || km.Fd < 0 {
		t.Fatalf("frames=%d keymap=%+v", frames, km)
	}
	f := os.NewFile(uintptr(km.Fd), "keymap")
	defer f.Close()
	buf := make([]byte, 6)
	if _, err := f.Read(buf); err != nil || string(buf) != "keymap" {
		t.Fatalf("keymap fd read %q, %v", buf, err)
	}
}

// A message split across several socket reads is reassembled.
func TestShortRead(t *testing.T) {
	ctx, srv := pair(t)
	reg := &Registry{}
	ctx.Register(reg)
	var got RegistryGlobalEvent
	reg.SetGlobalHandler(func(e RegistryGlobalEvent) { got = e })

	msg := event(reg.ID(), 0, u32(7), str("zwlr_virtual_pointer_manager_v1"), u32(2))
	go func() {
		for _, part := range [][]byte{msg[:5], msg[5:12], msg[12:]} {
			if _, err := srv.Write(part); err != nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	dispatch(t, ctx)
	if got.Name != 7 || got.Interface != "zwlr_virtual_pointer_manager_v1" || got.Version != 2 {
		t.Fatalf("global = %+v", got)
	}
}

// Events racing a destroy are dropped (closing their fds) until delete_id,
// which then frees the id for reuse. Unknown ids are still errors.
func TestZombieAndIDReuse(t *testing.T) {
	ctx, srv := pair(t)
	display := NewDisplay(ctx)
	kb := &Keyboard{}
	ctx.Register(kb)
	id := kb.ID()
	if err := kb.Release(); err != nil {
		t.Fatal(err)
	}

	r, w := pipeFd(t, "")
	send(t, srv, event(id, 0, u32(1), u32(0)), int(r.Fd()))
	r.Close()
	dispatch(t, ctx)
	if len(ctx.fds) != 0 {
		t.Fatalf("queued fds = %v", ctx.fds)
	}
	if _, err := w.WriteString("x"); !errors.Is(err, unix.EPIPE) {
		t.Fatalf("zombie fd left open: write err = %v", err)
	}

	cb := &Callback{}
	ctx.Register(cb)
	if cb.ID() == id {
		t.Fatal("id reused before delete_id")
	}

	deleted := uint32(0)
	display.SetDeleteIdHandler(func(e DisplayDeleteIdEvent) { deleted = e.Id })
	send(t, srv, event(display.ID(), 1, u32(id)))
	dispatch(t, ctx)
	if deleted != id {
		t.Fatalf("delete_id handler got %d, want %d", deleted, id)
	}
	cb2 := &Callback{}
	ctx.Register(cb2)
	if cb2.ID() != id {
		t.Fatalf("id = %d, want reused %d", cb2.ID(), id)
	}

	send(t, srv, event(99, 0))
	if err := ctx.Dispatch(); !errors.Is(err, ErrDispatchSenderNotFound) {
		t.Fatalf("unknown sender: err = %v", err)
	}
}

// delete_id may arrive before the proxy is destroyed (wl_callback).
func TestDeleteIDBeforeDestroy(t *testing.T) {
	ctx, srv := pair(t)
	display := NewDisplay(ctx)
	cb := &Callback{}
	ctx.Register(cb)
	send(t, srv, event(display.ID(), 1, u32(cb.ID())))
	dispatch(t, ctx)
	if err := cb.Destroy(); err != nil {
		t.Fatal(err)
	}
	next := &Callback{}
	ctx.Register(next)
	if next.ID() != cb.ID() {
		t.Fatalf("id = %d, want reused %d", next.ID(), cb.ID())
	}
}

// new_id event arguments register the server-created object; null object
// arguments decode to nil instead of panicking.
func TestServerCreatedObjects(t *testing.T) {
	ctx, srv := pair(t)
	dev := &DataDevice{}
	ctx.Register(dev)
	const offerID = serverIDStart + 1

	// No handler: the object must still be registered.
	send(t, srv, event(dev.ID(), 0, u32(offerID)))
	dispatch(t, ctx)
	offer, ok := ctx.GetProxy(offerID).(*DataOffer)
	if !ok || offer.ID() != offerID || offer.Context() != ctx {
		t.Fatalf("data_offer not registered: %#v", ctx.GetProxy(offerID))
	}

	var mime string
	offer.SetOfferHandler(func(e DataOfferOfferEvent) { mime = e.MimeType })
	send(t, srv, event(offerID, 0, str("text/plain")))
	dispatch(t, ctx)
	if mime != "text/plain" {
		t.Fatalf("mime = %q", mime)
	}

	sel := &DataOffer{}
	dev.SetSelectionHandler(func(e DataDeviceSelectionEvent) { sel = e.Id })
	send(t, srv, event(dev.ID(), 5, u32(0)))
	dispatch(t, ctx)
	if sel != nil {
		t.Fatalf("null selection = %#v, want nil", sel)
	}
}

func TestConnectAbsoluteDisplay(t *testing.T) {
	dir, err := os.MkdirTemp("", "wl")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	sock := filepath.Join(dir, "wayland-9")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	t.Setenv("XDG_RUNTIME_DIR", "")
	t.Setenv("WAYLAND_DISPLAY", "wayland-9")
	if _, err := Connect(""); err == nil {
		t.Fatal("relative WAYLAND_DISPLAY without XDG_RUNTIME_DIR: want error")
	}

	t.Setenv("WAYLAND_DISPLAY", sock)
	d, err := Connect("")
	if err != nil {
		t.Fatal(err)
	}
	d.Context().Close()

	t.Setenv("XDG_RUNTIME_DIR", dir)
	t.Setenv("WAYLAND_DISPLAY", "wayland-9")
	d, err = Connect("")
	if err != nil {
		t.Fatal(err)
	}
	d.Context().Close()
}
