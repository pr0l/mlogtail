# Makefile for mlogtail

BINARY_NAME=mlogtail
VERSION=1.2.0
BUILD_FLAGS=-ldflags="-s -w"

.PHONY: all build test clean deb install

# Default target
all: build

# Build the binary
build:
	CGO_ENABLED=0 go build -ldflags="-s -w" -o mlogtail .

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf debian/mlogtail/
	rm -f ../*.deb

# Build DEB package
deb: clean
	dpkg-buildpackage -us -uc -b

# Install locally (for development)
install: build
	sudo cp $(BINARY_NAME) /usr/local/bin/
	sudo cp mlogtail.service /etc/systemd/system/
	sudo cp mlogtail-reset.service /etc/systemd/system/
	sudo cp mlogtail-reset.timer /etc/systemd/system/
	sudo systemctl daemon-reload

# Uninstall
uninstall:
	sudo systemctl stop mlogtail.service || true
	sudo systemctl disable mlogtail.service || true
	sudo systemctl stop mlogtail-reset.timer || true
	sudo systemctl disable mlogtail-reset.timer || true
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	sudo rm -f /etc/systemd/system/mlogtail.service
	sudo rm -f /etc/systemd/system/mlogtail-reset.service
	sudo rm -f /etc/systemd/system/mlogtail-reset.timer
	sudo systemctl daemon-reload

# Show help
help:
	@echo "Available targets:"
	@echo "  build     - Build the binary"
	@echo "  test      - Run tests"
	@echo "  clean     - Clean build artifacts"
	@echo "  deb       - Build DEB package"
	@echo "  install   - Install locally for development"
	@echo "  uninstall - Remove local installation"
	@echo "  help      - Show this help"