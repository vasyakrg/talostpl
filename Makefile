help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  build-linux-amd64 - Build for Linux"
	@echo "  build-linux-arm64 - Build for Linux"
	@echo "  build-macos-amd64 - Build for macOS"
	@echo "  build-macos-arm64 - Build for macOS"
	@echo "  build-windows - Build for Windows"
	@echo "  build-all - Build for all platforms"
	@echo "  build-all-linux - Build for Linux only"

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o distrib/talostpl-linux-amd64 main.go
	chmod +x distrib/talostpl-linux-amd64

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o distrib/talostpl-linux-arm64 main.go
	chmod +x distrib/talostpl-linux-arm64

build-macos-amd64:
	GOOS=darwin GOARCH=amd64 go build -o distrib/talostpl-macos-amd64 main.go
	chmod +x distrib/talostpl-macos-amd64

build-macos-arm64:
	GOOS=darwin GOARCH=arm64 go build -o distrib/talostpl-macos-arm64 main.go
	chmod +x distrib/talostpl-macos-arm64

build-windows:
	GOOS=windows GOARCH=amd64 go build -o distrib/talostpl-windows.exe main.go

build-all-amd64:
	make build-linux-amd64
	make build-linux-arm64
	make build-macos-amd64
	make build-macos-arm64
	make build-windows

build-all-arm64:
	make build-linux-arm64
	make build-macos-arm64
	make build-windows

build-all:
	make build-all-amd64
	make build-all-arm64

default:
	make build-all-amd64
