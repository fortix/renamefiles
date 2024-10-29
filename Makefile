# Set the name of your project
PROJECT_NAME := renamefiles

# Set the platforms to build for
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64

# Set the flags to use for building
BUILD_FLAGS := -ldflags="-s -w" -tags=netgo -installsuffix netgo -trimpath

# Set the output directory
OUTPUT_DIR := bin

# Get the VERSION from go run ./scripts/getversion
VERSION := $(shell go run ./scripts/getversion)

default: all

.PHONY: all
## Build the binary for all platforms
all: clean $(PLATFORMS)

.PHONY: $(PLATFORMS)
$(PLATFORMS):
	GOOS=$(word 1,$(subst /, ,$@)) GOARCH=$(word 2,$(subst /, ,$@)) go build $(BUILD_FLAGS) -o $(OUTPUT_DIR)/$(PROJECT_NAME)_$(VERSION)_$(word 1,$(subst /, ,$@))_$(word 2,$(subst /, ,$@))$(if $(filter windows,$(word 1,$(subst /, ,$@))),.exe,) .

.PHONY: clean
## Remove the previous build
clean:
	rm -rf $(OUTPUT_DIR)/*

.PHONY: build
## Build the binary for the current platform
build:
	go build $(BUILD_FLAGS) -o $(OUTPUT_DIR)/$(PROJECT_NAME)$(if $(filter windows,$(word 1,$(subst /, ,$@))),.exe,) .

.PHONY: release
## Tag, build, and create a GitHub release
release: clean tag all container create-release

.PHONY: tag
## Tag the current code
tag:
	git tag -a v$(VERSION) -m "Release $(VERSION)"
	git push origin v$(VERSION)

.PHONY: create-release
## Create a GitHub release
create-release:
	gh release create v$(VERSION) $(OUTPUT_DIR)/proxydns_$(VERSION)* -t "Release $(VERSION)" -n "ProxyDNS $(VERSION)"

.PHONY: help
## This help screen
help:
	@printf "Available targets:\n\n"
	@awk '/^[a-zA-Z\-_0-9%:\\]+/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = $$1; \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			gsub("\\\\", "", helpCommand); \
			gsub(":+$$", "", helpCommand); \
			printf "  \x1b[32;01m%-20s\x1b[0m %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST) | sort -u
	@printf "\n"
