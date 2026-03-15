BINARY_NAME := sshm
MODULE := github.com/sadh/sshm
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
ARCHIVE_VERSION := $(patsubst v%,%,$(VERSION))
LDFLAGS := -ldflags "-X $(MODULE)/cmd.Version=$(VERSION) -s -w"
INSTALL_DIR := /usr/local/bin

.PHONY: build install uninstall release clean tidy test

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@install -m 755 $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Done. Run '$(BINARY_NAME) --help' to get started."

uninstall:
	@echo "Removing $(INSTALL_DIR)/$(BINARY_NAME)..."
	@rm -f $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Done."

tidy:
	go mod tidy

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

release: clean
	mkdir -p dist/pkg/darwin_amd64 dist/pkg/darwin_arm64 dist/pkg/linux_amd64 dist/pkg/linux_arm64 dist/pkg/windows_amd64
	GOOS=linux   GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64    .
	GOOS=linux   GOARCH=arm64  go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64    .
	GOOS=darwin  GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64   .
	GOOS=darwin  GOARCH=arm64  go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64   .
	GOOS=windows GOARCH=amd64  go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe .
	cp dist/$(BINARY_NAME)-darwin-amd64 dist/pkg/darwin_amd64/$(BINARY_NAME)
	cp dist/$(BINARY_NAME)-darwin-arm64 dist/pkg/darwin_arm64/$(BINARY_NAME)
	cp dist/$(BINARY_NAME)-linux-amd64 dist/pkg/linux_amd64/$(BINARY_NAME)
	cp dist/$(BINARY_NAME)-linux-arm64 dist/pkg/linux_arm64/$(BINARY_NAME)
	cp dist/$(BINARY_NAME)-windows-amd64.exe dist/pkg/windows_amd64/$(BINARY_NAME).exe
	tar -C dist/pkg/darwin_amd64 -czf dist/$(BINARY_NAME)_$(ARCHIVE_VERSION)_darwin_amd64.tar.gz $(BINARY_NAME)
	tar -C dist/pkg/darwin_arm64 -czf dist/$(BINARY_NAME)_$(ARCHIVE_VERSION)_darwin_arm64.tar.gz $(BINARY_NAME)
	tar -C dist/pkg/linux_amd64 -czf dist/$(BINARY_NAME)_$(ARCHIVE_VERSION)_linux_amd64.tar.gz $(BINARY_NAME)
	tar -C dist/pkg/linux_arm64 -czf dist/$(BINARY_NAME)_$(ARCHIVE_VERSION)_linux_arm64.tar.gz $(BINARY_NAME)
	cd dist/pkg/windows_amd64 && zip -q ../../$(BINARY_NAME)_$(ARCHIVE_VERSION)_windows_amd64.zip $(BINARY_NAME).exe
	@echo "\nBuilt binaries:"
	@ls -lh dist/
