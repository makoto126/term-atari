default: run

bindata:
	go-bindata roms/

build:
	go build

run:
	go run *.go