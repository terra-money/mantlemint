package heleveldb

type Item struct {
	Key      []byte
	Value    []byte
	AtHeight int64
	Deleted  bool
}

const (
	cOriginalDataPrefix   = "originalData/"
	cHeightSnapShotPrefix = "heightSnapshot/"
)

func prefixOriginalDataKey(key []byte) []byte {
	return append([]byte(cOriginalDataPrefix), key...)
}

func prefixHeightSnapshotKey(key []byte) []byte {
	return append(append([]byte(cHeightSnapShotPrefix), key...), []byte(":")...)
}
