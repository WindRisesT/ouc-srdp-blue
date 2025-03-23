module loglog

go 1.22.1

replace gopkg.in/fsnotify.v1 => github.com/fsnotify/fsnotify v1.5.1

require github.com/hpcloud/tail v1.0.0

require (
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	gopkg.in/fsnotify.v1 v1.0.0-00010101000000-000000000000 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
)
