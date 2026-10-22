package tablet

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg tablet -prefix zwp -suffix v2 -o tablet.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/stable/tablet/tablet-v2.xml?ref_type=tags
