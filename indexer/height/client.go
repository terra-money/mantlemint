package height

import (
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmdb "github.com/tendermint/tm-db"
)

func GetLastKownHeight(indexerDB tmdb.DB) (height HeightRecord, err error) {
	heightBytes, err := indexerDB.Get(key)
	if err != nil {
		return height, err
	}
	err = tmjson.Unmarshal(heightBytes, &height)
	return
}
