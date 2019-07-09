.PHONY: lint test dev-requirements requirements
PKGS := $(shell go list ./... | grep -v /vendor/)

lint:
	for pkg in $(PKGS) ; do \
		golint $$pkg ; \
	done
	go vet $(PKGS)

test: lint
	go test -cover -v ./...

dev-requirements:
	go mod download
	go install golang.org/x/lint/golint

