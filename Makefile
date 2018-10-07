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

build-docker-demo: build-linux
	@cp ./bin/vault-gcp-cloud-kms-pki ./demos/docker
	@docker build -t vault-gcp-cloud-kms-pki ./demos/docker

run-docker-demo: build-docker-demo
	@docker run --rm --name vault \
		-v "$(PWD)/tmp-vault-data:/vault-file" \
		vault-gcp-cloud-kms-pki

.PHONY: all test build build-linux release-snapshot