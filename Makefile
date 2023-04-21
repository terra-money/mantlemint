#!/usr/bin/make -f

BUILDDIR ?= $(CURDIR)/build
DOCKER := $(shell which docker)
SHA256_CMD = sha256sum
GO_VERSION ?= "1.20"

ifeq (,$(VERSION))
  VERSION := $(shell git describe --tags)
  # if VERSION is empty, then populate it with branch's name and raw commit hash
  ifeq (,$(VERSION))
    VERSION := $(BRANCH)-$(COMMIT)
  endif
endif

build: go.sum
ifeq ($(OS),Windows_NT)
	exit 1
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/mantlemint ./sync.go
endif


build-static:
	mkdir -p $(BUILDDIR)
	$(DOCKER) buildx build --tag terramoney/mantlemint ./
	$(DOCKER) create --name temp terramoney/mantlemint:latest
	$(DOCKER) cp temp:/usr/local/bin/mantlemint $(BUILDDIR)/
	$(DOCKER) rm temp

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./


build-release-amd64: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name mantlemint-builder || true
	$(DOCKER) buildx use mantlemint-builder
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
    --build-arg BUILDPLATFORM=linux/amd64 \
    --build-arg GOOS=linux \
    --build-arg GOARCH=amd64 \
		-t mantlemint:local-amd64 \
		--load \
		-f Dockerfile .
	$(DOCKER) rm -f mantlemint-builder || true
	$(DOCKER) create -ti --name mantlemint-builder mantlemint:local-amd64
	$(DOCKER) cp mantlemint-builder:/usr/local/bin/mantlemint $(BUILDDIR)/release/mantlemint
	tar -czvf $(BUILDDIR)/release/mantlemint_$(VERSION)_Linux_x86_64.tar.gz -C $(BUILDDIR)/release/ mantlemint
	rm $(BUILDDIR)/release/mantlemint
	$(DOCKER) rm -f mantlemint-builder

build-release-arm64: go.sum $(BUILDDIR)/
	$(DOCKER) buildx create --name mantlemint-builder  || true
	$(DOCKER) buildx use mantlemint-builder 
	$(DOCKER) buildx build \
		--build-arg GO_VERSION=$(GO_VERSION) \
		--build-arg GIT_VERSION=$(VERSION) \
		--build-arg GIT_COMMIT=$(COMMIT) \
    --build-arg BUILDPLATFORM=linux/arm64 \
    --build-arg GOOS=linux \
    --build-arg GOARCH=arm64 \
		-t mantlemint:local-arm64 \
		--load \
		-f Dockerfile .
	$(DOCKER) rm -f mantlemint-builder || true
	$(DOCKER) create -ti --name mantlemint-builder mantlemint:local-arm64
	$(DOCKER) cp mantlemint-builder:/usr/local/bin/mantlemint $(BUILDDIR)/release/mantlemint 
	tar -czvf $(BUILDDIR)/release/mantlemint_$(VERSION)_Linux_aarch64.tar.gz -C $(BUILDDIR)/release/ mantlemint 
	rm $(BUILDDIR)/release/mantlemint
	$(DOCKER) rm -f mantlemint-builder