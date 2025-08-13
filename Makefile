# Default target
all: build

test:
	go test -v ./...

test-watch:
	find . -name "*.go" | entr -c go test -v ./...

test-wrapper:
	@echo "Testing bash wrapper..."
	@bash shell/test-wrapper.sh

build:
	go build -o denv ./cmd/denv

clean:
	rm -f denv
	go clean -testcache

install: build
	cp denv /usr/local/bin/denv-core
	@echo "Installed denv-core to /usr/local/bin/"

install-wrapper: build
	@echo "Installing with bash wrapper support..."
	@bash shell/install-wrapper.sh

# Run both Go and wrapper tests
test-all: test test-wrapper

.PHONY: all test test-watch test-wrapper build clean install install-wrapper test-all