.PHONY: lint test dev-requirements requirements

lint:
	golint ./...
	go vet ./...

test: lint
	go test -cover -v ./...

dev-requirements:
	go mod download
	go install golang.org/x/lint/golint

