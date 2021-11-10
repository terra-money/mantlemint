package heleveldb

var (
	cCurrentDataPrefix     = []byte("currentData/")
	cKeysForIteratorPrefix = []byte("keysForIterator/")
	cDataWithHeightPrefix  = []byte("dataWithHeight/")
)

func prefixAliveKey(key []byte) []byte {
	return append(cCurrentDataPrefix, key...)
}

func prefixOriginalDataKey(key []byte) []byte {
	return append(cKeysForIteratorPrefix, key...)
}

func prefixHeightSnapshotKey(key []byte) []byte {
	result := make([]byte, 0, len(cDataWithHeightPrefix)+len(key)+1)
	result = append(result, cDataWithHeightPrefix...)
	result = append(result, key...)
	result = append(result, byte(':'))
	return result
}
