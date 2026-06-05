package xdg_system_bell

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg xdg_system_bell -prefix xdg -suffix v1 -o xdg_system_bell.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.48/staging/xdg-system-bell/xdg-system-bell-v1.xml?ref_type=tags
