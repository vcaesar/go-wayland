package ext_image_capture_source

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg ext_image_capture_source -prefix ext -suffix v1 -o ext_image_capture_source.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/staging/ext-image-capture-source/ext-image-capture-source-v1.xml?ref_type=tags
