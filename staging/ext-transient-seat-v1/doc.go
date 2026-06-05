package ext_transient_seat

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg ext_transient_seat -prefix ext -suffix v1 -o ext_transient_seat.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.48/staging/ext-transient-seat/ext-transient-seat-v1.xml?ref_type=tags
