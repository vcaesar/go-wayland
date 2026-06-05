package fifo

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg fifo -prefix wp -suffix v1 -o fifo.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.48/staging/fifo/fifo-v1.xml?ref_type=tags
