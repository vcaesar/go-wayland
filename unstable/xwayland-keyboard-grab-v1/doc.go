package xwayland_keyboard_grab

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg xwayland_keyboard_grab -prefix zwp -suffix v1 -o xwayland_keyboard_grab.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/unstable/xwayland-keyboard-grab/xwayland-keyboard-grab-unstable-v1.xml?ref_type=tags
