APP_NAME = hclvet
EPOCH_TIME = $(shell date +%s)
GIT_COMMIT = $(shell git rev-parse --short HEAD)
GO_LDFLAGS = '-X "github.com/clintjedwards/${APP_NAME}/internal/cli.appVersion=$(VERSION)"'
SHELL = /bin/bash
VERSION = ${SEMVER}_${GIT_COMMIT}_${EPOCH_TIME}

build-protos:
	protoc --go_out=. --go_opt=paths=source_relative \
	 --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	 internal/plugin/proto/*.proto

build: export CGO_ENABLED=0
build: check-path-included check-semver-included build-protos
	go build -ldflags $(GO_LDFLAGS) -o $(path)

check-path-included:
ifndef path
	$(error path is undefined; ex. path=/tmp/${APP_NAME})
endif

check-semver-included:
ifndef SEMVER
	$(error SEMVER is undefined; ex. SEMVER=1.1.0)
endif
