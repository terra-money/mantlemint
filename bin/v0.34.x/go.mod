module github.com/terra-money/mantlemint-provider-v0.34.x

go 1.17

require (
	github.com/cosmos/cosmos-sdk v0.44.2
	github.com/gocql/gocql v0.0.0-20210817081954-bc256bbb90de
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/hashicorp/golang-lru v0.5.4
	github.com/pkg/errors v0.9.1
	github.com/scylladb/gocqlx/v2 v2.4.0
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.13
	github.com/tendermint/tm-db v0.6.4
	github.com/terra-money/core v0.5.7
	github.com/terra-money/mantlemint v0.0.0

)

require github.com/golang/snappy v0.0.4 // indirect

replace github.com/terra-money/mantlemint => ../../

replace github.com/cosmos/ledger-cosmos-go => github.com/terra-money/ledger-terra-go v0.11.2

replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/99designs/keyring => github.com/cosmos/keyring v1.1.7-0.20210622111912-ef00f8ac3d76
