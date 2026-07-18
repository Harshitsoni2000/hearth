BINARY := hearth

.PHONY: build run test clean

build:
	go build -o $(BINARY) ./cmd/hearth

run: build
	./$(BINARY)

test:
	go test ./...

clean:
	rm -f $(BINARY)
