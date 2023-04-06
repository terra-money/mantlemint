ARG GO_VERSION="1.18"
ARG ALPINE_VERSION="3.16"
ARG BUILDPLATFORM=linux/amd64
ARG BASE_IMAGE="golang:${GO_VERSION}-alpine${ALPINE_VERSION}"
FROM --platform=${BUILDPLATFORM} ${BASE_IMAGE} as base

###############################################################################
# Builder
###############################################################################
FROM base as builder-stage-1

ARG BUILDPLATFORM

# NOTE: add libusb-dev to run with LEDGER_ENABLED=true
RUN set -eux &&\
    apk add --no-cache \
    linux-headers \
    ca-certificates \
    build-base \
    cmake \
    git

WORKDIR /go/src/mimalloc

# use mimalloc for musl
RUN set -eux &&\
    git clone --depth 1 --branch v2.0.9 \
        https://github.com/microsoft/mimalloc . &&\
    mkdir -p build &&\
    cd build &&\
    cmake .. &&\
    make -j$(nproc) &&\
    make install

WORKDIR /go/src/mantlemint
COPY . .

# Cosmwasm - Download correct libwasmvm version
# See https://github.com/CosmWasm/wasmvm/releases
RUN set -eux &&\
    WASMVM_VERSION=$(go list -m github.com/CosmWasm/wasmvm | cut -d ' ' -f 2) && \
    WASMVM_DOWNLOADS="https://github.com/CosmWasm/wasmvm/releases/download/${WASMVM_VERSION}"; \
    wget ${WASMVM_DOWNLOADS}/checksums.txt -O /tmp/checksums.txt; \
    if [ ${BUILDPLATFORM} = "linux/amd64" ]; then \
        WASMVM_URL="${WASMVM_DOWNLOADS}/libwasmvm_muslc.x86_64.a"; \
    elif [ ${BUILDPLATFORM} = "linux/arm64" ]; then \
        WASMVM_URL="${WASMVM_DOWNLOADS}/libwasmvm_muslc.aarch64.a"; \
    else \
        echo "Unsupported Build Platfrom ${BUILDPLATFORM}"; \
        exit 1; \
    fi; \
    wget ${WASMVM_URL} -O /lib/libwasmvm_muslc.a; \
    CHECKSUM=`sha256sum /lib/libwasmvm_muslc.a | cut -d" " -f1`; \
    grep ${CHECKSUM} /tmp/checksums.txt; \
    rm /tmp/checksums.txt 

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
RUN set -eux &&\
    LEDGER_ENABLED=false \
    go build -work \
    -tags muslc,linux \
    -mod=readonly \
    -ldflags="-extldflags '-L/go/src/mimalloc/build -lmimalloc -static'" \
    -o /go/bin/mantlemint \
    ./sync.go

###############################################################################
FROM alpine:${ALPINE_VERSION} as terra-core

WORKDIR /root

COPY --from=builder-stage-1 /go/bin/mantlemint /usr/local/bin/mantlemint

ENV CHAIN_ID="localterra" \
    MANTLEMINT_HOME="/app" \
    ## db paths relative to MANTLEMINT_HOME
    INDEXER_DB="/data/indexer" \ 
    MANTLEMINT_DB="/data/mantlemint" \
    GENESIS_PATH="/app/config/genesis.json" \
    DISABLE_SYNC="false" \
    RUST_BACKTRACE="full" \
    ENABLE_EXPORT_MODULE="false" \
    RICHLIST_LENGTH="100" \
    RICHLIST_THRESHOLD="0uluna" \
    ACCOUNT_ADDRESS_PREFIX="terra" \
    BOND_DENOM="uluna" \
    LCD_ENDPOINTS="http://localhost:1317" \
    RPC_ENDPOINTS="http://localhost:26657" \
    WS_ENDPOINTS="ws://localhost:26657/websocket" 

# lcd & grpc ports
EXPOSE 1317
EXPOSE 9090

CMD ["/usr/local/bin/mantlemint"]
