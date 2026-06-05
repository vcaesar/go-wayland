module github.com/vcaesar/go-wayland/examples/imageviewer

go 1.19

require (
	github.com/vcaesar/go-wayland/wayland v0.1.0
	github.com/nfnt/resize v0.0.0-20180221191011-83c6a9932646
	golang.org/x/image v0.3.0
	golang.org/x/sys v0.29.0
)

replace github.com/vcaesar/go-wayland/wayland v0.1.0 => ../../wayland
