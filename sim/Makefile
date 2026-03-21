.PHONY: build test vet lint clean

build:
	go build -o chaosseed-sim ./cmd/chaosseed-sim

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run

clean:
	rm -f chaosseed-sim coverage.out coverage.html
