package heleveldb

var (
	cCurrentDataPrefix     = []byte{0}
	cKeysForIteratorPrefix = []byte{1}
	cDataWithHeightPrefix  = []byte{2}
)

func prefixCurrentDataKey(key []byte) []byte {
	return append(cCurrentDataPrefix, key...)
}

func prefixKeysForIteratorKey(key []byte) []byte {
	return append(cKeysForIteratorPrefix, key...)
}

func prefixDataWithHeightKey(key []byte) []byte {
	result := make([]byte, 0, len(cDataWithHeightPrefix)+len(key))
	result = append(result, cDataWithHeightPrefix...)
	result = append(result, key...)
	return result
}
