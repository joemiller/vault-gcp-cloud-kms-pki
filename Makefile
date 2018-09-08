APP := vault-cloud-kms-pki

test:
	@go test -v ./...

build:
	@CGO_ENABLED=0 go build -o bin/$(APP)

build-linux:
	@GOOS=linux GOOARCH=amd64 CGO_ENABLED=0 go build -o bin/$(APP)

run-dev: build

.PHONY: all test build build-linux
