.PHONY: build run test lint clean

build:
	go build -o bb ./cmd/bb

run: build
	./bb

test:
	go test ./... -v

lint:
	golangci-lint run ./...

clean:
	rm -f bb
