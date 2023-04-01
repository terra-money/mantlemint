# docker build . -t cosmwasm/wasmd:latest
# docker run --rm -it cosmwasm/wasmd:latest /bin/sh
FROM golang:1.18-alpine3.14 AS go-builder

COPY . src/mantlemint

# See https://github.com/CosmWasm/wasmvm/releases
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.a

# NOTE: add libusb-dev to run with LEDGER_ENABLED=true
RUN set -eux &&\
    apk add --no-cache \
    linux-headers \
    ca-certificates \
    build-base \
    cmake \
    git

# use mimalloc for musl
RUN set -eux &&\
    git clone --depth 1 --branch v2.0.9 https://github.com/microsoft/mimalloc src/mimalloc &&\
    mkdir -p src/mimalloc/build &&\
    cd src/mimalloc/build &&\
    cmake .. &&\
    make -j$(nproc) &&\
    make install

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
RUN set -eux &&\
    cd src/mantlemint &&\
    LEDGER_ENABLED=false \
    go build -work \
    -tags muslc,linux \
    -mod=readonly \
    -ldflags="-extldflags '-L/go/src/mimalloc/build -lmimalloc -static'" \
    -o /go/bin/mantlemint \
    ./sync.go

###############################################################################
FROM alpine:3.14

WORKDIR /root

COPY --from=go-builder /go/bin/mantlemint /usr/local/bin/mantlemint

RUN apk add bash

# lcd & grpc ports
EXPOSE 1317
EXPOSE 9090

CMD ["/usr/local/bin/mantlemint"]