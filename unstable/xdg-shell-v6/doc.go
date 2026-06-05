package xdg_shell

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg xdg_shell -prefix zxdg -suffix v6 -o xdg_shell.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.48/unstable/xdg-shell/xdg-shell-unstable-v6.xml?ref_type=tags
