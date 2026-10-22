package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"unsafe"

	"golang.org/x/sys/unix"
)

// maxFds is the most fds one sendmsg may carry (libwayland MAX_FDS_OUT).
const maxFds = 28

var oobSpace = unix.CmsgSpace(maxFds * 4)

// ReadMsg reads one message. fd is the first queued fd if any fd arrived
// while reading it, else -1. [Context.GetDispatch] instead hands fds to the
// events that carry them.
func (ctx *Context) ReadMsg() (senderID uint32, opcode uint32, fd int, msg []byte, err error) {
	senderID, opcode, nfds, msg, err := ctx.readMsg()
	fd = -1
	if nfds > 0 {
		fd = ctx.popFd()
	}
	return senderID, opcode, fd, msg, err
}

// readMsg reads one whole message, queueing every fd received meanwhile, and
// returns how many fds arrived with it.
func (ctx *Context) readMsg() (senderID, opcode uint32, nfds int, msg []byte, err error) {
	nfds, err = ctx.readFull(ctx.header[:])
	if err != nil {
		return 0, 0, nfds, nil, fmt.Errorf("ctx.ReadMsg: header: %w", err)
	}

	senderID = Uint32(ctx.header[:4])
	opcodeAndSize := Uint32(ctx.header[4:8])
	opcode = opcodeAndSize & 0xffff
	size := int(opcodeAndSize >> 16)
	if size < 8 {
		return senderID, opcode, nfds, nil, fmt.Errorf("ctx.ReadMsg: invalid message size %d", size)
	}
	if size == 8 {
		return senderID, opcode, nfds, nil, nil
	}

	// Not reused: decoded strings alias the message bytes.
	msg = make([]byte, size-8)
	n, err := ctx.readFull(msg)
	nfds += n
	if err != nil {
		return senderID, opcode, nfds, msg, fmt.Errorf("ctx.ReadMsg: body: %w", err)
	}
	return senderID, opcode, nfds, msg, nil
}

// readFull fills b from the socket, appending any received fds to the queue.
func (ctx *Context) readFull(b []byte) (nfds int, err error) {
	for len(b) > 0 {
		n, oobn, flags, _, err := ctx.conn.ReadMsgUnix(b, ctx.oob)
		if oobn > 0 {
			fds, perr := getFdsFromOob(ctx.oob, oobn, "msg")
			if perr != nil {
				return nfds, perr
			}
			ctx.mu.Lock()
			ctx.fds = append(ctx.fds, fds...)
			ctx.mu.Unlock()
			nfds += len(fds)
		}
		if err != nil {
			return nfds, err
		}
		if flags&unix.MSG_CTRUNC != 0 {
			return nfds, errors.New("file descriptors truncated")
		}
		if n == 0 {
			return nfds, io.ErrUnexpectedEOF
		}
		b = b[n:]
	}
	return nfds, nil
}

func getFdsFromOob(oob []byte, oobn int, source string) ([]int, error) {
	if oobn > len(oob) {
		return nil, fmt.Errorf("getFdsFromOob: incorrect number of bytes read from %s for oob (oobn=%d)", source, oobn)
	}
	scms, err := unix.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return nil, fmt.Errorf("getFdsFromOob: unable to parse control message from %s: %w", source, err)
	}

	var fdsRet []int
	for _, scm := range scms {
		fds, err := unix.ParseUnixRights(&scm)
		if err != nil {
			return nil, fmt.Errorf("getFdsFromOob: unable to parse unix rights from %s: %w", source, err)
		}

		fdsRet = append(fdsRet, fds...)
	}

	return fdsRet, nil
}

func Uint32(src []byte) uint32 {
	_ = src[3]
	return *(*uint32)(unsafe.Pointer(&src[0]))
}

func String(src []byte) string {
	if idx := bytes.IndexByte(src, 0); idx >= 0 {
		src = src[:idx:idx]
	}
	return *(*string)(unsafe.Pointer(&src))
}

func Fixed(src []byte) float64 {
	_ = src[3]
	fx := *(*int32)(unsafe.Pointer(&src[0]))
	return fixedToFloat64(fx)
}
