.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	staticcheck ./...

.PHONY: lint.tools.install
lint.tools.install:
	go install honnef.co/go/tools/cmd/staticcheck@2023.1.2
