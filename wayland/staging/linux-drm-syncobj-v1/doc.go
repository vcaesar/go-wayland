package linux_drm_syncobj

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg linux_drm_syncobj -prefix wp -suffix v1 -o linux_drm_syncobj.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.48/staging/linux-drm-syncobj/linux-drm-syncobj-v1.xml?ref_type=tags
