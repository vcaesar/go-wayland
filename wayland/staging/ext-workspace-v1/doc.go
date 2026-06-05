package ext_workspace

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg ext_workspace -prefix ext -suffix v1 -o ext_workspace.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.48/staging/ext-workspace/ext-workspace-v1.xml?ref_type=tags
