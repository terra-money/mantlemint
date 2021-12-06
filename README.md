# terra-money/mantlemint


## What is Mantlemint?

Mantlemint is a fast core optimized for serving massive user queries.

Native query performance on RPC is very slow and is not suitable for massive query handling, due to the inefficiencies introduced by IAVL tree. Mantlemint is running on `fauxMerkleTree` mode, basically removing the IAVL inefficiencies while using the same core to compute the same module outputs. 

If you are looking to serve any kind of public node accepting varying degrees of end-user queries, it is recommended that you run a mantlemint instance alongside of your RPC. While mantlemint is indeed faster at resolving queries, due to the absence of IAVL tree and native tendermint, it cannot join p2p network by itself. Rather, you would have to relay finalized blocks to mantlemint, using RPC's websocket.

## Currently supported terra-money/core versions
- columbus-5 | terra-money/core@0.5.x | tendermint v0.34.x


## Features

- Superior LCD performance
    - With the exception of Tendermint RPC/Transactions.
- Super reliable and effective LCD response cache to prevent unnecessary computation for query resolving
- Fully archival; historical states are available with `?height` query parameter.
- [Useful default indexes](#default-indexes)


## Installation

This specific directory contains mantlemint implementation for [@terra-money/core@0.5.x](https://github.com/terra-money/core) (compatible with [tendermint@0.34.x](https://github.com/tendermint/tendermint)).

Go v1.17+ is recommended for this project.

```sh
$ export MANTLEMINT_VERSION=v0.34.x
$ git clone https://github.com/terra-money/mantlemint.git
$ cd bin/$MANTLEMINT_VERSION
$ go build sync.go
```

## Usage

### Things you will need

#### 1. Access to at least 1 running RPC node

Since mantlemint cannot join p2p network by itself, it depends on RPC to receive recently proposed blocks.

Any [Terra node](https://github.com/terra-money/core) with port 26657 enabled can be used for this.

#### 2. `config/app.toml`, a genesis file

Mantlemint internally runs the same Terra Core, therefore you need to provide the same configuration files as if you would run an RPC. Bare minimum you would at least need `app.toml` and `genesis.json`.

It is __required__ to run mantlemint in a separate `$HOME` directory than RPC; while mantlemint maintains its own database, some of the data may be overwritten by either mantlemint or RPC and may cause trouble.

### Running

Mantlemint depends on 2 configs:
- `$HOME/config/app.toml`; you can reuse `app.toml` you're using with core
- Environment variables; mantlemint specific runtime variables to configure various properties of mantlemint. Examples as follows

```sh
# Location of genesis file
GENESIS_PATH=config/genesis.json \

# Home directory for mantlemint.
# Mantlemint will use this to:
# - read $HOME/config/app.toml
# - create and maintain $HOME/mantlemint.db directory
# - create and maintain $HOME/data/* for wasm blobs; (unsafe to share with RPC!)
# - create and maintain $HOME/$(INDEXER_DB).db for mantle indexers
HOME=mantlemint \

# Chain ID 
CHAIN_ID=columbus-5 \

# RPC Endpoint; used to sync previous blocks when mantlemint is catching up
RPC_ENDPOINTS=http://rpc1:26657,http://rpc2:26657 \

# WS Endpoint; used to sync live block as soon as they are available through RPC websocket  
WS_ENDPOINTS=ws://rpc1:26657/websocket,ws://rpc2:26657/websocket \

# Name of indexer db
INDEXER_DB=indexer \

# Flag to enable/disable mantlemint sync, mainly for debugging
DISABLE_SYNC=true \

# Run sync binary
sync
```


## Default Indexes

- `/index/tx/by_height/{height}`: List all transactions and their responses in a block. Equivalent to `tendermint/block?height=xxx`, with tx responses base64-decoded for better usability.
- `/index/tx/by_hash/{txHash}`: Get transaction and its response by hash. Equivalent to `lcd/txs/{hash}`, but without hitting RPC.

# LICENSE

Apache 2.0
