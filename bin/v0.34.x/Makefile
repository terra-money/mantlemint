#!/usr/bin/make -f

BUILDDIR=build

build: go.sum
ifeq ($(OS),Windows_NT)
	exit 1
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/mantlemint ./sync
endif


build-linux:
	mkdir -p $(BUILDDIR)
	docker buildx build --platform=linux/amd64 --tag terramoney/mantlemint ./
	docker create --platform=linux/amd64 --name temp terramoney/mantlemint:latest
	docker cp temp:/usr/local/bin/mantlemint $(BUILDDIR)/
	docker rm temp

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./