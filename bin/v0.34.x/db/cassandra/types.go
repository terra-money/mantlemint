package cassandra

const (
	cPartKey  = "partition_key"
	cKey      = "key"
	cValue    = "value"
	cAtHeight = "at_height"
	cDeleted  = "deleted"

	// blobstore
	cData = "data"
)

var (
	columns = []string{cPartKey, cKey, cValue, cDeleted, cAtHeight}
	partKey = []string{cPartKey}
	sortKey = []string{cAtHeight}

	blobStoreColumns = []string{cPartKey, cData}
	blobStorePartKey = []string{cPartKey}
)

type Item struct {
	PartitionKey []byte
	Key          []byte
	Value        []byte
	AtHeight     int64
	Deleted      bool
}

type Blob struct {
	PartitionKey []byte
	Data    []byte
}

