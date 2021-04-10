PACKAGES=$(shell go list ./... | grep -v '/simulation')

VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
COMMIT := $(shell git log -1 --format='%H')

ldflags = -X github.com/cosmos/cosmos-sdk/version.Name=Artery \
	-X github.com/cosmos/cosmos-sdk/version.ServerName=artrd \
	-X github.com/cosmos/cosmos-sdk/version.ClientName=artrcli \
	-X github.com/cosmos/cosmos-sdk/version.Version=$(VERSION) \
	-X github.com/cosmos/cosmos-sdk/version.Commit=$(COMMIT) 

BUILD_FLAGS := -ldflags '$(ldflags)'

all: install

install: go.sum
		go install $(BUILD_FLAGS) ./cmd/artrd
		go install $(BUILD_FLAGS) ./cmd/artrcli

dev:
		go install $(BUILD_FLAGS) ./cmd/artrd
		go install $(BUILD_FLAGS) ./cmd/artrcli

go.sum: go.mod
		@echo "--> Ensure dependencies have not been modified"
		GO111MODULE=on go mod verify

test:
		@go test -v -mod=readonly -tags=testing $(PACKAGES)

# look into .golangci.yml for enabling / disabling linters
lint:
	@echo "--> Running linter"
	@golangci-lint run
	@go mod verify

android:
	ANDROID_HOME=~/Android/Sdk gomobile bind -target=android/amd64 -v ./app
