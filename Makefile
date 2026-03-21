.PHONY: test test-race vet lint cover check clean viz

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

lint:
	golangci-lint run

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

check: vet lint test-race

viz:
	go run ./cmd/caveviz

clean:
	rm -f coverage.out coverage.html
