package cursor_shape

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg cursor_shape -prefix wp -suffix v1 -o cursor_shape.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.48/staging/cursor-shape/cursor-shape-v1.xml?ref_type=tags
