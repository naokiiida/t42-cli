BINARY_NAME = t42

.PHONY: all build install tidy lint test clean

all: build

build:
	go build -o $(BINARY_NAME) .

install:
	go install

tidy:
	go mod tidy

lint:
	golangci-lint run

test:
	go test ./...

clean:
	rm -f $(BINARY_NAME)