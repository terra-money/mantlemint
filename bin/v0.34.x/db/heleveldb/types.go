package heleveldb

type Item struct {
	Key      []byte
	Value    []byte
	AtHeight int64
	Deleted  bool
}

var (
	cAliveDataPrefix      = []byte("aliveData/")
	cOriginalDataPrefix   = []byte("originalData/")
	cHeightSnapShotPrefix = []byte("heightSnapshot/")
)

func prefixAliveKey(key []byte) []byte {
	return append(cAliveDataPrefix, key...)
}

func prefixOriginalDataKey(key []byte) []byte {
	return append(cOriginalDataPrefix, key...)
}

func prefixHeightSnapshotKey(key []byte) []byte {
	result := make([]byte, 0, len(cHeightSnapShotPrefix)+len(key)+1)
	result = append(result, cHeightSnapShotPrefix...)
	result = append(result, key...)
	result = append(result, byte(':'))
	return result
}
