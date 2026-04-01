build:
	CGO_ENABLED=0 go build -o outpost ./cmd/outpost/

run: build
	./outpost

test:
	go test ./...

clean:
	rm -f outpost

.PHONY: build run test clean
