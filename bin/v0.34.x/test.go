package main

import (
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/heleveldb"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/hld"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/safe_batch"

	tmdb "github.com/tendermint/tm-db"
)

// initialize mantlemint for v0.34.x
func main() {

	var ldb, ldbErr = heleveldb.NewLevelDBDriver(&heleveldb.DriverConfig{"test", "./testdb"})
	if ldbErr != nil {
		panic(ldbErr)
	}

	// simple test
	write(ldb.Session(), "k", "v")
	print(ldb.Session(), "k")

	var hldb = hld.ApplyHeightLimitedDB(
		ldb,
		&hld.HeightLimitedDBConfig{
			Debug: true,
		},
	)

	dd, err := ldb.Get(1, []byte("k"))
	if err == nil {
		println(dd)
	} else {
		println(err)
	}

	batched := safe_batch.NewSafeBatchDB(hldb)
	batchedOrigin := batched.(safe_batch.SafeBatchDBCloser)

	hldb.SetWriteHeight(1)
	batchedOrigin.Open()
	write(batched, "a", "1")
	write(batched, "b", "1")
	write(batched, "c", "1")
	write(batched, "e", "1")
	batchedOrigin.Flush()
	hldb.ClearWriteHeight()

	hldb.SetWriteHeight(2)
	batchedOrigin.Open()
	write(batched, "a", "2")
	write(batched, "d", "1")
	batchedOrigin.Flush()
	hldb.ClearWriteHeight()

	hldb.SetWriteHeight(3)
	batchedOrigin.Open()
	write(batched, "c", "2")
	write(batched, "z", "33")
	batched.Delete([]byte("e"))
	batchedOrigin.Flush()
	hldb.ClearWriteHeight()

	hldb.SetWriteHeight(4)
	batchedOrigin.Open()
	write(batched, "b", "2")
	write(batched, "d", "2")
	write(batched, "he", "23")
	batchedOrigin.Flush()
	hldb.ClearWriteHeight()

	hldb.SetWriteHeight(5)
	batchedOrigin.Open()
	write(batched, "a", "3")
	write(batched, "c", "3")
	batchedOrigin.Flush()
	hldb.ClearWriteHeight()

	println("중간에 삭제되는 e 테스트")
	for kk := 1; kk < 6; kk++ {
		v, _ := ldb.Get(int64(kk), []byte("e"))
		println(kk, string(v), v == nil)
	}

	// println("전체 데이터")
	// 	i, e := ldb.Session().Iterator([]byte("a"), nil)
	// 	if e != nil {
	// 		panic(e)
	// 	}
	// 	for ; i.Valid(); i.Next() {
	// 		println(string(i.Key()), "-", string(i.Value()))
	// 	}
	// 	i.Close()

	println("height 별 데이터")
	for r := 0; r < 5; r++ {
		println("height : ", r+1)
		hldb.SetReadHeight(int64(r + 1))
		i, e := batched.Iterator([]byte("a"), nil)
		if e != nil {
			println(e)
			panic(e)
		}
		for ; i.Valid(); i.Next() {
			println(string(i.Key()), "-", string(i.Value()))
		}
		i.Close()
		hldb.ClearReadHeight()
	}

	println("1차 iterator 에서 범위 지정해보기")
	hldb.SetReadHeight(1)
	i, e := batched.Iterator([]byte("b"), []byte("j"))
	if e != nil {
		println(e)
		panic(e)
	}
	for ; i.Valid(); i.Next() {
		println(string(i.Key()), "-", string(i.Value()))
	}
	i.Close()
	hldb.ClearReadHeight()
}

func print(db tmdb.DB, keyString string) {
	key := []byte(keyString)

	if res, err := db.Get(key); err == nil {

		println(string(res))
	} else {
		println("err")
	}
}

func write(db tmdb.DB, keyString, dataString string) {
	key := []byte(keyString)
	value := []byte(dataString)

	if err := db.Set(key, value); err != nil {
		println(err)

	}
}
