package heleveldb

var (
	cCurrentDataPrefix     = []byte{0}
	cKeysForIteratorPrefix = []byte{1}
	cDataWithHeightPrefix  = []byte{2}
)

func prefixAliveKey(key []byte) []byte {
	return append(cCurrentDataPrefix, key...)
}

func prefixOriginalDataKey(key []byte) []byte {
	return append(cKeysForIteratorPrefix, key...)
}

func prefixHeightSnapshotKey(key []byte) []byte {
	result := make([]byte, 0, len(cDataWithHeightPrefix)+len(key))
	result = append(result, cDataWithHeightPrefix...)
	result = append(result, key...)
	return result
}
