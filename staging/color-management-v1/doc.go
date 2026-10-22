package color_management

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg color_management -prefix wp -suffix v1 -o color_management.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/staging/color-management/color-management-v1.xml?ref_type=tags
