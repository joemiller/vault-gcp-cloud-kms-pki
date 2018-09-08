APP := vault-gcp-cloud-kms-pki

test:
	@go test -v ./...

build:
	@CGO_ENABLED=0 go build -o bin/$(APP)

build-debug:
	@CGO_ENABLED=0 go build -gcflags='-N -l' -a -o bin/$(APP)

build-linux:
	@GOOS=linux GOOARCH=amd64 CGO_ENABLED=0 go build -o bin/$(APP)

run-dev: build
	sh ./run-dev.sh

.PHONY: all test build build-linux
