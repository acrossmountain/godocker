SOURCE=cmd/main.go
TARGET=godocker

compile: build

build:
	go build -o ${TARGET} ${SOURCE}
