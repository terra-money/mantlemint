package height

var (
	key    = []byte("lastKnownHeight")
	getKey = func() []byte {
		return key
	}
)

type HeightRecord struct {
	Height uint64 `json:"height"`
}
