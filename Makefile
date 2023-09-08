.PHONY: test
test:
	go test ./...

.PHONY: lint
lint:
	staticcheck ./...

.PHONY: lint.tools.install
lint.tools.install:
	go install honnef.co/go/tools/cmd/staticcheck@2023.1.5
