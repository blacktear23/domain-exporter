CURDIR     := $(shell pwd)
BUILD_PATH := $(CURDIR)/build
LDFLAGS    := -s -w
BUILD_ARGS := -trimpath -ldflags '$(LDFLAGS)'

.PHONY: all build-linux

all: build-linux build-darwin

prepare-path:
	@mkdir -p $(BUILD_PATH)/linux
	@mkdir -p $(BUILD_PATH)/darwin

build-linux: prepare-path
	GOOS=linux go build $(BUILD_ARGS) -o $(BUILD_PATH)/linux/domain-exporter
	cp $(CURDIR)/config.yaml $(BUILD_PATH)/linux/config.yaml

build-darwin: prepare-path
	GOOS=darwin go build $(BUILD_ARGS) -o $(BUILD_PATH)/darwin/domain-exporter
	cp $(CURDIR)/config.yaml $(BUILD_PATH)/darwin/config.yaml

build-darwin-only: prepare-path
	GOOS=darwin go build $(BUILD_ARGS) -o $(BUILD_PATH)/darwin/domain-exporter
