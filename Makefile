IMAGE_NAME ?= vennekilde/gw2verify:latest

.build(%): export CGO_ENABLED=0
.build(%):
	go build -installsuffix 'static' -o ./bin/$(BIN_NAME) ./cmd/gw2verify

build: .build(native)
build_all: build_linux build_windows

build_windows: export GOOS=windows
build_windows: export GOARCH=amd64
build_windows: BIN_NAME := ${BIN_NAME}.exe
build_windows: .build(windows)

build_linux: export GOOS=linux
build_linux: export GOARCH=amd64
build_linux: .build(linux)

package: build_linux image-build

image-build:
	docker build . -t ${IMAGE_NAME}