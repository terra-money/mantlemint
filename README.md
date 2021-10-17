# terra-money/mantlemint


## What is Mantlemint?

Mantlemint is a fast core optimized for serving massive user queries. Native query performance on RPC is very slow and is not suitable for massive query handling, due to the inefficiencies introduced by IAVL tree. Mantlemint is running on `fauxMerkleTree` mode, basically removing the IAVL inefficiencies while using the same core to compute the same module outputs. 

If you are looking to serve any kind of public node accepting varying degrees of end-user queries, it is recommended that you run a mantlemint instance alongside of your RPC. While mantlemint is indeed faster at resolving queries, due to the absence of IAVL tree and native tendermint, it cannot join p2p network by itself. Rather, you would have to relay finalized blocks to mantlemint, using RPC's websocket.

## Currently supported terra-money/core versions
- columbus-5 | terra-money/core@0.5.x | tendermint v0.34.x


# LICENSE

Apache


