NAME := eqspellfix
BUILD_VERSION ?= 0.1

SHELL := /bin/bash

##@ Build
.PHONY: build
build: ## build quail for local OS and windows
	@echo "build: building to bin/quail..."
	go build main.go
	-mv main bin/quail

build-all: build-darwin build-windows build-linux build-windows-addon ## build all supported os's


build-darwin: ## build darwin
	@echo "build-darwin: ${BUILD_VERSION}"
	@CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -buildmode=pie -ldflags="-X main.Version=${BUILD_VERSION} -s -w" -o bin/${NAME}-darwin main.go

build-linux: ## build linux
	@echo "build-linux: ${BUILD_VERSION}"
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-X main.Version=${BUILD_VERSION} -s -w" -o bin/${NAME}-linux main.go

build-windows: ## build windows
	@echo "build-windows: ${BUILD_VERSION}"
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -buildmode=pie -ldflags="-X main.Version=${BUILD_VERSION} -s -w" -o bin/${NAME}.exe main.go

build-windows-addon: ## build windows blender addon
	@echo "build-windows-addon: ${BUILD_VERSION}"
	@CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -buildmode=pie -ldflags="-X main.Version=${BUILD_VERSION} -X main.ShowVersion=1 -s -w" -o bin/${NAME}-addon.exe main.go

build-copy: build-darwin ## used by xackery, build darwin copy and move to blender path
	@echo "copying to quail-addons..."
	cp bin/quail-darwin "/Users/xackery/Library/Application Support/Blender/3.4/scripts/addons/quail-addon/quail-darwin"

# CICD triggers this
set-version-%:
	@echo "VERSION=${BUILD_VERSION}.$*" >> $$GITHUB_ENV
