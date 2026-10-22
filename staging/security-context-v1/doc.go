package security_context

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg security_context -prefix wp -suffix v1 -o security_context.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/staging/security-context/security-context-v1.xml?ref_type=tags
