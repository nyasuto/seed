.PHONY: build test vet lint clean

build:
	go build -o chaosforge .

test:
	go test ./...

vet:
	go vet ./...

lint:
	golangci-lint run

clean:
	rm -f chaosforge coverage.out coverage.html
