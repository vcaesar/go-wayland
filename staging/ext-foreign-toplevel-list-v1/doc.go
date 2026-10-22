package ext_foreign_toplevel_list

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg ext_foreign_toplevel_list -prefix ext -suffix v1 -o ext_foreign_toplevel_list.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.49/staging/ext-foreign-toplevel-list/ext-foreign-toplevel-list-v1.xml?ref_type=tags
