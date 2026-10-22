package wlr_virtual_keyboard

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg wlr_virtual_keyboard -prefix zwp -suffix v1 -o wlr_virtual_keyboard.go -i https://gitlab.freedesktop.org/wlroots/wlroots/-/raw/0.20.0/protocol/virtual-keyboard-unstable-v1.xml?ref_type=tags
