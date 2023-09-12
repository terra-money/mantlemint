#!/usr/bin/make -f

BUILDDIR ?= $(CURDIR)/build

build: go.sum
ifeq ($(OS),Windows_NT)
	exit 1
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/mantlemint ./sync.go
endif

lint:
	golangci-lint run --out-format=tab

lint-fix:
	golangci-lint run --fix --out-format=tab --issues-exit-code=0

lint-strict:
	find . -path './_build' -prune -o -type f -name '*.go' -exec gofumpt -w -l {} +

build-static:
	mkdir -p $(BUILDDIR)
	docker buildx build --tag terramoney/mantlemint ./
	docker create --name temp terramoney/mantlemint:latest
	docker cp temp:/usr/local/bin/mantlemint $(BUILDDIR)/
	docker rm temp

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

clean:
	rm -rf $(BUILDDIR)/
