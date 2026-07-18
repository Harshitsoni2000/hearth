BINARY := hearth

.PHONY: build run test fmt clean

build:
	go build -o $(BINARY) ./cmd/hearth

run:
	go run ./cmd/hearth -dir ./testmedia

test:
	go test ./...

fmt:
	go fmt ./...

clean:
	rm -f $(BINARY)
