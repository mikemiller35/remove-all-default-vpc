BIN_NAME=remove-all-default-vpc

.PHONY: build clean test mockgen

build: clean
	CGO_ENABLED=0 go build -ldflags='-s -w' -o bin/${BIN_NAME} .

clean:
	rm -rf bin/${BIN_NAME}

test:
	go test -covermode=count -coverprofile=coverage.out ./...
	go tool gocover-cobertura < coverage.out > coverage.xml

mockgen: ## Generate mocks for interfaces using mockgen.
	@echo "Generating mocks for interfaces..."
	go generate ./...