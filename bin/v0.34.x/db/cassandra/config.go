package cassandra

import (
	"github.com/gocql/gocql"
)

type DriverConfig struct {
	Clusters              []string
	Keyspace              string
	TableName             string
	ArchivalMode          bool
	Consistency           gocql.Consistency
	IteratorPrefetchCount uint
}

func DefaultCassandraConfig() *DriverConfig {
	return &DriverConfig{
		Clusters:              []string{"127.0.0.1"},
		Keyspace:              "",
		ArchivalMode:          true,
		Consistency:           gocql.LocalQuorum,
		IteratorPrefetchCount: 100,
	}
}

func (config *DriverConfig) WithClusters(clusters []string) *DriverConfig {
	config.Clusters = clusters
	return config
}

func (config *DriverConfig) WithKeyspace(keyspace string) *DriverConfig {
	config.Keyspace = keyspace
	return config
}

func (config *DriverConfig) WithTableName(tableName string) *DriverConfig {
	config.TableName = tableName
	return config
}

func (config *DriverConfig) WithArchivalMode(archivalMode bool) *DriverConfig {
	config.ArchivalMode = archivalMode
	return config
}

func (config *DriverConfig) WithConsistency(consistency gocql.Consistency) *DriverConfig {
	config.Consistency = consistency
	return config
}

func (config *DriverConfig) WithIteratorPrefetchCount(iteratorPrefetchCount uint) *DriverConfig {
	config.IteratorPrefetchCount = iteratorPrefetchCount
	return config
}

// ------- blob store config

type BlobStoreDriverConfig struct {
	Clusters    []string
	Keyspace    string
	TableName   string
	Consistency gocql.Consistency
}

func DefaultBlobStoreDriverConfig() *BlobStoreDriverConfig {
	return &BlobStoreDriverConfig{
		Clusters:    []string{"127.0.0.1"},
		Keyspace:    "blobstore",
		TableName:   "blobstore",
		Consistency: gocql.LocalQuorum,
	}
}

func (config *BlobStoreDriverConfig) WithClusters(clusters []string) *BlobStoreDriverConfig {
	config.Clusters = clusters
	return config
}

func (config *BlobStoreDriverConfig) WithKeyspace(keyspace string) *BlobStoreDriverConfig {
	config.Keyspace = keyspace
	return config
}

func (config *BlobStoreDriverConfig) WithTableName(tableName string) *BlobStoreDriverConfig {
	config.TableName = tableName
	return config
}

func (config *BlobStoreDriverConfig) WithConsistency(consistency gocql.Consistency) *BlobStoreDriverConfig {
	config.Consistency = consistency
	return config
}
