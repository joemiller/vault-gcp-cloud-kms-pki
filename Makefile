APP := vault-gcp-cloud-kms-pki

test:
	@go test -v ./...

build:
	@CGO_ENABLED=0 go build -o bin/$(APP)

build-debug:
	@go build -a -gcflags='-N -l' -o bin/$(APP)

build-linux:
	@GOOS=linux GOOARCH=amd64 CGO_ENABLED=0 go build -o bin/$(APP)

release-snapshot:
	@rm -rf ./dist
	@goreleaser --snapshot

run-dev: build
	sh ./run-dev.sh

.PHONY: all test build build-linux release-snapshot