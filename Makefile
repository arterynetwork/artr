PACKAGES=$(shell go list ./... | grep -v '/simulation')
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')
DOCKER := $(shell which docker)

STATIK = "$(GOPATH)/bin/statik"

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=Artery \
	-X github.com/cosmos/cosmos-sdk/version.ServerName=artrd \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) 

BUILD_FLAGS := -ldflags '$(ldflags)'

build-all: proto
		"$(DOCKER)" run --rm -v "$(CURDIR):/art-node" -w /art-node golang:1.15-alpine sh ./scripts/build-all.sh

build: go.sum
		go build $(BUILD_FLAGS) ./cmd/artrd

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		go mod verify
		go mod tidy

test: proto
		@go test -v -mod=readonly -tags=testing $(PACKAGES)

# look into .golangci.yml for enabling / disabling linters
lint:
	@echo "--> Running linter"
	@golangci-lint run
	@go mod verify

android:
	ANDROID_HOME=~/Android/Sdk gomobile bind -target=android/amd64 -v ./app

###############################################################################
###                                Protobuf                                 ###
###############################################################################

proto: proto-clean proto-gen proto-swagger update-swagger-docs

proto-clean:
	find x -type f -iname *.pb.go -delete
	find x -type f -iname *.pb.gw.go -delete

proto-gen:
	@echo "Generating Protobuf files"
	"$(DOCKER)" run --rm -v "$(CURDIR):/workspace" -w /workspace tendermintdev/sdk-proto-gen \
		sh ./scripts/protocgen.sh

proto-swagger:
	@echo "Generating Protobuf Swagger"
	"$(DOCKER)" run --rm -v "$(CURDIR):/workspace" -w /workspace tendermintdev/sdk-proto-gen \
		sh ./scripts/protoc-swagger-gen.sh

update-swagger-docs: statik
	$(STATIK) -src=client/docs/swagger-ui -dest=client/docs -f -m -ns=swagger
	@if [ -n "$(git status --porcelain)" ]; then \
        echo "\033[91mSwagger docs are out of sync!!!\033[0m";\
        exit 1;\
    else \
        echo "\033[92mSwagger docs are in sync\033[0m";\
    fi

statik:
	@echo "Installing statik..."
	@(cd /tmp && export GO111MODULE=on && go get github.com/rakyll/statik@v0.1.7)
