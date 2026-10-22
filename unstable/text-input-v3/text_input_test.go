package text_input

import (
	"bytes"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/vcaesar/go-wayland/client"
)

// TestArrayRequestWire checks array request encoding: length word, payload
// and padding to 32-bit boundary.
func TestArrayRequestWire(t *testing.T) {
	dir, err := os.MkdirTemp("", "wl")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	ln, err := net.Listen("unix", filepath.Join(dir, "wayland-0"))
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()

	d, err := client.Connect(filepath.Join(dir, "wayland-0"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Context().Close()
	srv, err := ln.Accept()
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	ti := NewTextInput(d.Context())
	if err := ti.SetAvailableActions([]byte{1, 2, 3, 4, 5}); err != nil {
		t.Fatal(err)
	}

	got := make([]byte, 20)
	if _, err := io.ReadFull(srv, got); err != nil {
		t.Fatal(err)
	}
	want := make([]byte, 20)
	client.PutUint32(want[0:], ti.ID())
	client.PutUint32(want[4:], 20<<16|8)
	client.PutUint32(want[8:], 5)
	copy(want[12:], []byte{1, 2, 3, 4, 5})
	if !bytes.Equal(got, want) {
		t.Fatalf("wire: got %v, want %v", got, want)
	}
}
