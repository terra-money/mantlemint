package height

var key = []byte("lastKnownHeight")
var getKey = func() []byte {
	return key
}

type HeightRecord struct {
	Height uint64 `json:"height"`
}


