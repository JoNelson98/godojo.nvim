.PHONY: build test clean

build:
	@mkdir -p bin
	go build -o bin/godojo ./cmd/godojo

test:
	go test -v ./...

clean:
	rm -rf bin
