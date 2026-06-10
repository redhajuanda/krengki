BINARY=krengki
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_FLAGS=-buildvcs=false -ldflags="-s -w -X github.com/redhajuanda/krengki/cmd.Version=$(VERSION)"

.PHONY: build install clean

build:
	go build $(BUILD_FLAGS) -o $(BINARY) .

install: build
	mv $(BINARY) /usr/local/bin/$(BINARY)

clean:
	rm -f $(BINARY)
