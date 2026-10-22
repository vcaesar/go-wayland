package color_representation

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg color_representation -prefix wp -suffix v1 -o color_representation.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/staging/color-representation/color-representation-v1.xml?ref_type=tags
