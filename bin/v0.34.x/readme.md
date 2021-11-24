
# mantlemint-v0.34.x

## Features

- Superior LCD performance due to removal of [IAVL tree](https://github.com/cosmos/iavl)
- Is compatible with almost all native [LCD resources](https://lcd.terra.dev/swagger/);
  - With the exception of Tendermint RPC/Transactions.
- Super reliable and effective LCD response cache to prevent unnecessary computation for query resolving
- Fully archival; historical states are available with `?height` query parameter.
- [Useful default indexes](#default-indexes)


## Installation

This specific directory contains mantlemint implementation for [@terra-money/core@0.5.x](https://github.com/terra-money/core) (compatible with [tendermint@0.34.x](https://github.com/tendermint/tendermint)).

```sh
$ git clone https://github.com/terra-money/mantlemint.git
$ cd bin/v0.34.x
$ go build sync.go
```

## Usage

Mantlemint depends on 2 configs:
- `$HOME/config/app.toml`; you can reuse `app.toml` you're using with core
- Environment variables; mantlemint specific runtime variables to configure various properties of mantlemint. Examples as follows

You also need at least 1 running RPC node, since mantlemint cannot join p2p network (as it is NOT running the p2p module) and it depends on RPC to receive blocks. 

```sh
# Location of genesis file
GENESIS_PATH=$(pwd)/test.no-commit.db/config/genesis.json \

# 
HOME=$(pwd)/test.no-commit.db \

# Chain ID in
CHAIN_ID=localterra \

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

## Accessing historical states

Mantlemint 


## Default Indexes

- `/index/tx/by_height/{height}`: List all transactions and their responses in a block. Equivalent to `tendermint/block?height=xxx`, with tx responses base64-decoded for better usability.
- `/index/tx/by_hash/{txHash}`: Get transaction and its response by hash. Equivalent to `lcd/txs/{hash}`, but without hitting RPC.

## Caveats



