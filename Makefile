.PHONY: test test-race vet lint cover check clean

MODULES := core sim game

test:
	@for mod in $(MODULES); do echo "=== test $$mod ===" && (cd $$mod && go test ./...) || exit 1; done

test-race:
	@for mod in $(MODULES); do echo "=== test-race $$mod ===" && (cd $$mod && go test -race ./...) || exit 1; done

vet:
	@for mod in $(MODULES); do echo "=== vet $$mod ===" && (cd $$mod && go vet ./...) || exit 1; done

lint:
	@for mod in $(MODULES); do echo "=== lint $$mod ===" && (cd $$mod && golangci-lint run) || exit 1; done

cover:
	@for mod in $(MODULES); do echo "=== cover $$mod ===" && (cd $$mod && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html) || exit 1; done

check: vet lint test-race

clean:
	@for mod in $(MODULES); do rm -f $$mod/coverage.out $$mod/coverage.html; done
