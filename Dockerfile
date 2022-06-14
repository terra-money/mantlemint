# docker build . -t cosmwasm/wasmd:latest
# docker run --rm -it cosmwasm/wasmd:latest /bin/sh
FROM golang:1.18-alpine3.14 AS go-builder

# this comes from standard alpine nightly file
#  https://github.com/rust-lang/docker-rust-nightly/blob/master/alpine3.12/Dockerfile
# with some changes to support our toolchain, etc
RUN set -eux; apk add --no-cache ca-certificates build-base;

RUN apk add git build-base cmake linux-headers
# NOTE: add these to run with LEDGER_ENABLED=true
# RUN apk add libusb-dev linux-headers

WORKDIR /code
COPY . /code/

# See https://github.com/CosmWasm/wasmvm/releases
ADD https://github.com/CosmWasm/wasmvm/releases/download/v1.0.0/libwasmvm_muslc.x86_64.a /lib/libwasmvm_muslc.a

# use mimalloc for musl
RUN git clone --depth 1 https://github.com/microsoft/mimalloc; cd mimalloc; mkdir build; cd build; cmake ..; make -j$(nproc); make install

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
RUN LEDGER_ENABLED=false go build -work -tags muslc,linux -mod=readonly -ldflags="-extldflags '-L/code/mimalloc/build -lmimalloc -static'" -o build/mantlemint ./sync.go

FROM alpine:3.12

WORKDIR /root

COPY --from=go-builder /code/build/mantlemint /usr/local/bin/mantlemint

# rest server
EXPOSE 1317
# grpc
EXPOSE 9090

CMD ["/usr/local/bin/mantlemint"]