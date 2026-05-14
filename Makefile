HOSTNAME=registry.terraform.io
NAMESPACE=terraform-providers
NAME=airtelcloud
BINARY=terraform-provider-${NAME}
VERSION=1.0.0
OS_ARCH=linux_amd64

default: install

build:
	go build -o ${BINARY}

release:
	GOOS=darwin GOARCH=amd64 go build -o bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -o bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o bin/${BINARY}_${VERSION}_windows_386.exe
	GOOS=windows GOARCH=amd64 go build -o bin/${BINARY}_${VERSION}_windows_amd64.exe
	GOOS=darwin GOARCH=arm64 go build -o bin/${BINARY}_${VERSION}_darwin_arm64


	zip bin/${BINARY}_${VERSION}_darwin_amd64.zip bin/${BINARY}_${VERSION}_darwin_amd64
	zip bin/${BINARY}_${VERSION}_darwin_arm64.zip bin/${BINARY}_${VERSION}_darwin_arm64
	zip bin/${BINARY}_${VERSION}_freebsd_386.zip bin/${BINARY}_${VERSION}_freebsd_386
	zip bin/${BINARY}_${VERSION}_freebsd_amd64.zip bin/${BINARY}_${VERSION}_freebsd_amd64
	zip bin/${BINARY}_${VERSION}_freebsd_arm.zip bin/${BINARY}_${VERSION}_freebsd_arm
	zip bin/${BINARY}_${VERSION}_linux_386.zip bin/${BINARY}_${VERSION}_linux_386
	zip bin/${BINARY}_${VERSION}_linux_amd64.zip bin/${BINARY}_${VERSION}_linux_amd64
	zip bin/${BINARY}_${VERSION}_linux_arm.zip bin/${BINARY}_${VERSION}_linux_arm
	zip bin/${BINARY}_${VERSION}_openbsd_386.zip bin/${BINARY}_${VERSION}_openbsd_386
	zip bin/${BINARY}_${VERSION}_openbsd_amd64.zip bin/${BINARY}_${VERSION}_openbsd_amd64
	zip bin/${BINARY}_${VERSION}_solaris_amd64.zip bin/${BINARY}_${VERSION}_solaris_amd64
	zip bin/${BINARY}_${VERSION}_windows_386zip bin/${BINARY}_${VERSION}_windows_386.exe
	zip bin/${BINARY}_${VERSION}_windows_amd64.zip bin/${BINARY}_${VERSION}_windows_amd64.exe


install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	cp ${BINARY} ~/go/bin/

test:
	go test -i $(shell go list ./...) || exit 1
	echo $(shell go list ./...) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc:
	TF_ACC=1 go test $(shell go list ./...) -v $(TESTARGS) -timeout 120m

fmt:
	gofmt -w $(shell find . -name '*.go' | grep -v vendor)

lint:
	golangci-lint run

docs-generate:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

.PHONY: build release install test testacc fmt lint docs-generate
