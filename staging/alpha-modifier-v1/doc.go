package alpha_modifier

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg alpha_modifier -prefix wp -suffix v1 -o alpha_modifier.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/staging/alpha-modifier/alpha-modifier-v1.xml?ref_type=tags
