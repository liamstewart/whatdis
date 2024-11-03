.PHONY: all server test clean

all: server

server:
	go build -o server cmd/server/main.go

test:
	go test ./...

clean:
	rm -f server
