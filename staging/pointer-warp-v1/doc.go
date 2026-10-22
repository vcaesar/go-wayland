package pointer_warp

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg pointer_warp -prefix wp -suffix v1 -o pointer_warp.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/staging/pointer-warp/pointer-warp-v1.xml?ref_type=tags
