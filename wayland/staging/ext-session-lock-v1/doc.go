package ext_session_lock

//go:generate go run github.com/vcaesar/go-wayland/cmd/go-wayland-scanner -pkg ext_session_lock -suffix v1 -o ext_session_lock.go -i https://gitlab.freedesktop.org/wayland/wayland-protocols/-/raw/1.48/staging/ext-session-lock/ext-session-lock-v1.xml?ref_type=tags
