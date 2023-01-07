#!/usr/bin/make -f

BUILDDIR=build

build: go.sum
ifeq ($(OS),Windows_NT)
	exit 1
else
	go build -mod=readonly $(BUILD_FLAGS) -o build/mantlemint ./sync.go
endif


build-static:
	mkdir -p $(BUILDDIR)
	docker buildx build --tag terramoney/mantlemint ./
	docker create --name temp terramoney/mantlemint:latest
	docker cp temp:/usr/local/bin/mantlemint $(BUILDDIR)/
	docker rm temp

install: go.sum
	go install -mod=readonly $(BUILD_FLAGS) ./

## format: Install and run goimports and gofumpt
format:
	@echo Formatting...
	@go run mvdan.cc/gofumpt -w .
	@go run golang.org/x/tools/cmd/goimports -w -local github.com/terra-money/feather-toolkit .

## lint: Run Golang CI Lint.
lint:
	@echo Running gocilint...
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint run --out-format=tab --issues-exit-code=0