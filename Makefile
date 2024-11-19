.PHONY: all server test clean lint publish

all: server

server:
	go build -o server cmd/server/main.go

test:
	go test ./...

clean:
	rm -f server

lint:
	golangci-lint run ./...

publish:
	ko build --bare ./cmd/server
