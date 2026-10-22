# Wayland implementation in Go

[![Go Reference](https://pkg.go.dev/badge/github.com/vcaesar/go-wayland.svg)](https://pkg.go.dev/github.com/vcaesar/go-wayland)

This module contains pure Go implementation of the Wayland protocol.
Currently only wayland-client functionality is supported.

Go code is generated from protocol XML files using
[`go-wayland-scanner`](cmd/go-wayland-scanner/scanner.go).

Protocols: core `wayland` 1.25 ([`client`](client)), `wayland-protocols` 1.49
([`stable`](stable), [`staging`](staging), [`unstable`](unstable)) and the
wlroots protocols ([`wlr`](wlr): layer-shell, screencopy, foreign-toplevel,
data-control, output-management, virtual-pointer, virtual-keyboard, ...).

Like libwayland, the [`client.Context`](client/context.go) queues received file
descriptors and hands each one to the event that carries it, registers objects
created by the compositor (`new_id` events), drops events racing a destroy and
reuses object ids after `wl_display.delete_id`. Registering and sending
requests are safe for concurrent use, except that requests creating objects
must be serialized (the compositor rejects new ids sent out of order).

To load cursor, minimal port of `wayland-cursor` & `xcursor` in pure Go
is located at [`wayland/cursor`](wayland/cursor) & [`wayland/cursor/xcursor`](wayland/cursor/xcursor)
respectively.

To demonstrate the functionality of this module
[`examples/imageviewer`](examples/imageviewer) contains a simple image
viewer. It demos displaying a top-level window, resizing of window,
cursor themes, pointer and keyboard. Because it's in pure Go, it can be
compiled without CGO. You can try it using the following commands:

```sh
CGO_ENABLED=0 go install github.com/vcaesar/go-wayland/examples/imageviewer@latest

imageviewer file.jpg
```

## Fork

This repository is forked from [github.com/rajveermalviya/go-wayland](https://github.com/rajveermalviya/go-wayland) and includes several fixes and improvements.
Forked at `0ad78d1`, see [the diff](https://github.com/vcaesar/go-wayland/compare/0ad78d1...main).
