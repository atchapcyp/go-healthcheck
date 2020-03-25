PKG = github.com/atchapcyp/go-healthcheck

WINDOWS_BUILD_VARS = CC=x86_64-w64-mingw32-gcc CGO_ENABLED=1 GOOS=windows GOARCH=amd64 BIN_EXTENSION=.exe
WIN_FLAGS = -tags "window"

.PHONY: build

build: 
	go build -o health -v $(PKG)/cmd
