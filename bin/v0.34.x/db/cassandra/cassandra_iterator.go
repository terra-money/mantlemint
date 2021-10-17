package cassandra

import (
	"bytes"
	"fmt"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/hld"
	"sort"
)

var _ hld.HeightLimitEnabledIterator = (*Iterator)(nil)

type IteratorConfiguration struct {
	PrefetchCount uint
	Reverse       bool
}

type Iterator struct {
	session       gocqlx.Session
	table         *table.Table
	maxHeight     int64
	storeKey      []byte
	start         []byte
	end           []byte
	prefetchCount uint
	reverse       bool

	seekPointer int64
	seekCache   []Item
}

func NewCassandraIterator(session gocqlx.Session, table *table.Table, maxHeight int64, start, end []byte, config IteratorConfiguration) *Iterator {
	return &Iterator{
		session:   session,
		table:     table,
		maxHeight: maxHeight,
		start:     start,
		end:       end,

		prefetchCount: config.PrefetchCount,
		reverse:       config.Reverse,
		seekPointer:   0,
		seekCache:     nil,
	}
}

func (i *Iterator) Load() error {
	ks := i.start
	err := i.session.Query(fmt.Sprintf("select partition_key,key,value,deleted,at_height from %s where key>=? and key<? AND at_height<=? PER PARTITION LIMIT 1 allow filtering", i.table.Metadata().Name), []string{
		"start", "end", "at_height",
	}).BindMap(qb.M{
		cKey:      ks,
		"start":   i.start,
		"end":     i.end,
		cAtHeight: i.maxHeight,
	}).SelectRelease(&i.seekCache)

	if err != nil {
		return err
	}

	i.seekPointer = 0
	fmt.Printf("[iterator] load count=%d\n", len(i.seekCache))

	// sort
	sort.Slice(i.seekCache, func(a, b int) bool {
		return bytes.Compare(i.seekCache[a].Key, i.seekCache[b].Key) != -1
	})

	return nil
}

func (i *Iterator) Domain() (start []byte, end []byte) {
	panic("implement me")
}

func (i *Iterator) Valid() bool {
	// nil seekCache means this iterator never loaded anything,
	// return false
	if len(i.seekCache) == 0 {
		return false
	}

	// if seekPointer is longer than the length of seekCache,
	// we've reached at the end
	// return false
	if i.seekPointer >= int64(len(i.seekCache)) {
		return false
	}

	// filter out items with Deleted = true
	// it should return somewhere during the loop
	// otherwise iterator has reached the end without finding any record
	// with Delete = false, return false in such case.
	var c int64

	defer func() {
		i.seekPointer = c
	}()

	for c = i.seekPointer; c < int64(len(i.seekCache)); c++ {
		current := i.seekCache[c]
		if current.Deleted {
			continue
		} else {
			return true
		}
	}

	return false
}

func (i *Iterator) Next() {
	i.seekPointer++
}

func (i *Iterator) Key() (key []byte) {
	return i.seekCache[i.seekPointer].Key
}

func (i *Iterator) Value() (value []byte) {
	val := i.seekCache[i.seekPointer].Value
	if val == nil {
		val = []byte{}
	}
	return val
}

func (i *Iterator) Error() error {
	return nil
	// panic("implement me")
}

func (i *Iterator) Close() error {
	return nil
	//panic("implement me")
}
