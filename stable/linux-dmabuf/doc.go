package linux_dmabuf

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg linux_dmabuf -prefix zwp -suffix v1 -o linux_dmabuf.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/stable/linux-dmabuf/linux-dmabuf-v1.xml?ref_type=tags
