BINARY_NAME = t42
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS = -s -w \
	-X 'github.com/naokiiida/t42-cli/cmd.version=$(VERSION)' \
	-X 'github.com/naokiiida/t42-cli/cmd.commit=$(COMMIT)' \
	-X 'github.com/naokiiida/t42-cli/cmd.date=$(DATE)'

# Append OAuth credentials if available (from env vars or secret/.env)
ifneq ($(FT_UID),)
LDFLAGS += -X 'github.com/naokiiida/t42-cli/cmd.embeddedClientID=$(FT_UID)'
endif
ifneq ($(FT_SECRET),)
LDFLAGS += -X 'github.com/naokiiida/t42-cli/cmd.embeddedClientSecret=$(FT_SECRET)'
endif

.PHONY: all build install tidy lint test clean

all: build

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) .

install:
	go install -ldflags "$(LDFLAGS)" .

tidy:
	go mod tidy

lint:
	golangci-lint run

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)
