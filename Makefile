SRC="main"

build: 
	go build -ldflags "-s" -o init main/main.go
debug: 
	go build -o init main/main.go
