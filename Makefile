help:
	@echo "Usage: make <target>"
	@echo "Targets:"
	@echo "  build-linux-amd64 - Build for Linux"
	@echo "  build-linux-arm64 - Build for Linux"
	@echo "  build-darwin-amd64 - Build for macOS"
	@echo "  build-darwin-arm64 - Build for macOS"
	@echo "  build-all - Build for all platforms"
	@echo "  build-all-linux - Build for Linux only"

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o distrib/talostpl-linux-amd64 main.go
	chmod +x distrib/talostpl-linux-amd64

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o distrib/talostpl-linux-arm64 main.go
	chmod +x distrib/talostpl-linux-arm64

build-darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -o distrib/talostpl-darwin-amd64 main.go
	chmod +x distrib/talostpl-darwin-amd64

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -o distrib/talostpl-darwin-arm64 main.go
	chmod +x distrib/talostpl-darwin-arm64

build-all-linux:
	make build-linux-amd64
	make build-linux-arm64

build-all-darwin:
	make build-darwin-amd64
	make build-darwin-arm64

build-all:
	make build-all-linux
	make build-all-darwin
