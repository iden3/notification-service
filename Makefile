.PHONY: test
test:
	go test -v -race -short -timeout=60s ./...

.PHONY: lint
lint:
	 golangci-lint --config .golangci.yml run
