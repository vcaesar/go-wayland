package xdg_dialog

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg xdg_dialog -prefix xdg -suffix v1 -o xdg_dialog.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/staging/xdg-dialog/xdg-dialog-v1.xml?ref_type=tags
