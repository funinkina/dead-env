.PHONY: build test lint fuzz integration clean

build:
	go build -o bin/deadenv ./...

test:
	go test -race ./...

lint:
	golangci-lint run ./...

fuzz:
	go test -fuzz=FuzzParseEnvContent -fuzztime=60s ./internal/parser/

integration:
	go test -tags=integration -v ./...

clean:
	rm -rf bin/
