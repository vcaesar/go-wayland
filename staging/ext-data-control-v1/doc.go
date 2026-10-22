package ext_data_control

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg ext_data_control -prefix ext -suffix v1 -o ext_data_control.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/staging/ext-data-control/ext-data-control-v1.xml?ref_type=tags
