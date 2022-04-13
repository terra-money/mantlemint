# terra-money/mantlemint


## What is Mantlemint?

Mantlemint is a fast core optimized for serving massive user queries.

Native query performance on RPC is slow and is not suitable for massive query handling, due to the inefficiencies introduced by IAVL tree. Mantlemint is running on `fauxMerkleTree` mode, basically removing the IAVL inefficiencies while using the same core to compute the same module outputs. 

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

#### 1. As a statically-linked application
```sh
$ make build-static # results in build/mantlemint
```

#### 2. As a dynamically-linked application
```sh
$ make build # results in build/mantlemint
$ make install # results in $GOPATH/bin/mantlemint
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

> Make sure you separate `MANTLEMINT_HOME` from other mantlemint instances, or core. Doing so may result in an undefined behaviour.

```sh
# Location of genesis file
GENESIS_PATH=config/genesis.json \

# Home directory for mantlemint.
# Mantlemint will use this to:
# - read $HOME/config/app.toml
# - create and maintain $HOME/mantlemint.db directory
# - create and maintain $HOME/data/* for wasm blobs; (unsafe to share with RPC!)
# - create and maintain $HOME/$(INDEXER_DB).db for mantle indexers
MANTLEMINT_HOME=mantlemint \

# Chain ID 
CHAIN_ID=columbus-5 \

# RPC Endpoint; used to sync previous blocks when mantlemint is catching up
RPC_ENDPOINTS=http://rpc1:26657,http://rpc2:26657 \

# WS Endpoint; used to sync live block as soon as they are available through RPC websocket  
WS_ENDPOINTS=ws://rpc1:26657/websocket,ws://rpc2:26657/websocket \

# Name of mantlemint.db, akin to application.db for core
MANTLEMINT_DB=mantlemint \

# Name of indexer db
INDEXER_DB=indexer \

# Flag to enable/disable mantlemint sync, mainly for debugging
DISABLE_SYNC=false \

# Run sync binary
sync

# Optional: crisis module's invariant check is known to take hours.
# You can skip it by providing --x-crisis-skip-assert-invariants flag
sync --x-crisis-skip-assert-invariants 
```

### Adjusting smart contract memory cache size

The `wasm` section in `config.toml` may play a critical role in how mantlemint performs under heavy load. We recommend adjusting `contract-memory-cache-size` if you are planning to run mantlemint publicly, as loading contract instances from disk is an expensive operation.

```toml
[wasm]
# The maximum gas amount can be spent for contract query.
# The contract query will invoke contract execution vm,
# so we need to restrict the max usage to prevent DoS attack
contract-query-gas-limit = "3000000"

# The flag to specify whether print contract logs or not
contract-debug-mode = "false"

# The WASM VM memory cache size in MiB not bytes.
# Adjust this if you need to hold more smart contract instances in memory (less overhead for loading contracts to memory)
contract-memory-cache-size = "16384" # 16GB
```



## Health check

`mantlemint` implements `/health` endpoint. It is useful if you want to suppress traffics being routed to `mantlemint` nodes still syncing or unavailable due to whatever reason.

The endpoint will response:
- `200 OK` if mantlemint sync status is up-to date (i.e. syncing using websocket from RPC)
- `400 NOK` if mantlemint is still syncing past blocks, and is not ready to serve the latest state yet.

Please note that mantlemint still is able to serve queries while `/health` returns `NOK`.

## Default Indexes

- `/index/tx/by_height/{height}`: List all transactions and their responses in a block. Equivalent to `tendermint/block?height=xxx`, with tx responses base64-decoded for better usability.
- `/index/tx/by_hash/{txHash}`: Get transaction and its response by hash. Equivalent to `lcd/txs/{hash}`, but without hitting RPC.

## Notable Differences from [core](https://github.com/terra-money/core)

- Uses a forked [tendermint/tm-db](https://github.com/terra-money/tm-db/commit/c71e8b6e9f20d7f5be32527db4a92ae19ac0d2b2): Disables unncessary mutexes in `prefixdb` methods
- Replaces ABCIClient with [NewConcurrentQueryClient](https://github.com/terra-money/mantlemint/blob/main/mantlemint/client.go#L110): Removal of mutexes allow better concurrency, even during block injection
- Uses single batch-protected db: All state changes are flushed at once, making it safe to read from db during block injection
- Automatic failover: In case of block injection failure, mantlemint reverts back to the previous known state and retry
- Strictly no `tendermint`; some parameters in app.toml would not affect `mantlemint`
- Following endpoints are  not implemented
  - `GET /blocks/`
  - `GET /blocks/latest`
  - `GET /txs/{hash}`
  - `GET /txs`
  - `GET /validatorset`
  - All `POST` variants


## FAQ

### Q1. Can I use public RPC (http://public-node.terra.dev:26657) as RPC and WS endpoints?

While you can, we do NOT recommend doing so. We only expose public node as a seed node for p2p, and its http/ws connection may not be stable. It is safer to have your own RPC 

### Q2. Can I convert existing core's database to mantlemint?

No. Mantlemint's db structure is NOT compatible with core's. 

### Q3. Mantlemint doesn't support tendermint queries like /blocks, /txs, but I still need them. What should I do?

You can route those traffic to an existing RPC that you use for `RPC_ENDPOINTS` and `WS_ENDPOINTS` with a load balancer.

### Q4. Is it possible to disable archive? It takes up too much space!

Not yet, although it is on the roadmap.

### Q5. Mantlemint seems to hang up on the first block.

Processing genesis block is known to take 2+ hours -- be patient!

Also, try disabling crisis module's invariant check on genesis block creation, by providing flag `--x-crisis-skip-assert-invariants`.

### Q6. What are the system requirements to run mantlemint safely?

- 4 or more CPU cores
- At least 16GB memory, 32~64GB for genesis block creation. 
  - You can scale down to smaller system after mantlemint successfully creates the genesis block
- At least 2TB disk space, although we recommend bigger.

### Q7. Are snapshots provided?

Not yet, but we have plans to do so.

### Q8. Mantlemint becomes unresponsive when put under load

Run [this](https://github.com/YunSuk-Yeo/wasm-cache-rebuilder) at least once. This happens because since cosmwasm@0.16.6 (+wasmer@2.2.1), deserialized wasm module format changed and you need to rebuild contract cache.


## Community

- [Offical Website](https://terra.money)
- [Discord](https://discord.gg/e29HWwC2Mz)
- [Telegram](https://t.me/terra_announcements)
- [Twitter](https://twitter.com/terra_money)
- [YouTube](https://goo.gl/3G4T1z)


# License

This software is licensed under the Apache 2.0 license. Read more about it here.

© 2021 Terraform Labs, PTE LTD
