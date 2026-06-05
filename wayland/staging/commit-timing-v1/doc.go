package commit_timing

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg commit_timing -prefix wp -suffix v1 -o commit_timing.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.48/staging/commit-timing/commit-timing-v1.xml?ref_type=tags
