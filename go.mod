module github.com/terra-money/mantlemint

go 1.16

replace github.com/CosmWasm/go-cosmwasm => github.com/terra-project/go-cosmwasm v0.10.5

require (
	github.com/cosmos/cosmos-sdk v0.39.3
	github.com/gocql/gocql v0.0.0-20210702075011-769848eae462
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gorilla/websocket v1.4.2
	github.com/pkg/errors v0.9.1
	github.com/scylladb/gocqlx/v2 v2.4.0
	github.com/stretchr/testify v1.6.1
	github.com/tendermint/tendermint v0.33.9
	github.com/tendermint/tm-db v0.5.2
	github.com/terra-project/core v0.4.6
)
