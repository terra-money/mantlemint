package cassandra

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/hld"
	"time"
)

// contract: HL Enabled
var _ hld.HeightLimitEnabledDB = (*Driver)(nil)

type Driver struct {
	session               gocqlx.Session
	table                 *table.Table
	tableName             string
	archivalMode          bool
	iteratorPrefetchCount uint
}

// NewCassandraDriver creates cassandra db connection
func NewCassandraDriver(config *DriverConfig) *Driver {
	cluster := gocql.NewCluster(config.Clusters...)
	cluster.Keyspace = config.Keyspace
	cluster.Consistency = gocql.All
	cluster.Timeout = 30 * time.Second
	cluster.NumConns = 16

	// create session and wrap in gocqlx
	session, err := gocqlx.WrapSession(cluster.CreateSession())

	if err != nil {
		panic(err)
	}

	tableModel := table.Metadata{
		Name:    config.TableName,
		Columns: columns,
		PartKey: partKey,
		SortKey: sortKey,
	}

	if err := session.Query(fmt.Sprintf("create table if not exists %s (partition_key blob, key blob, value blob, deleted boolean, at_height bigint, primary key(partition_key, at_height)) with clustering order by (at_height DESC)", config.TableName), []string{}).ExecRelease(); err != nil {
		panic(err)
	}

	table := table.New(tableModel)
	if err := initializeSasiOnKey(session, config.TableName); err != nil {
		panic(err)
	}
	if err := initializeSasiOnHeight(session, config.TableName); err != nil {
		panic(err)
	}

	return &Driver{
		session:               session,
		tableName:             config.TableName,
		table:                 table,
		archivalMode:          config.ArchivalMode,
		iteratorPrefetchCount: config.IteratorPrefetchCount,
	}
}

func (d *Driver) Get(maxHeight int64, key []byte) ([]byte, error) {
	var requestHeight = hld.Height(maxHeight).CurrentOrLatest().ToInt64()
	var requestHeightMin = hld.Height(0).CurrentOrNever().ToInt64()

	// check if requestHeightMin is
	if requestHeightMin > requestHeight {
		return nil, fmt.Errorf("invalid height")
	}

	// retrieve!
	var items []Item

	err := d.table.SelectBuilder(cKey, cValue, cAtHeight, cDeleted).
		Where(
			qb.LtOrEqNamed(cAtHeight, "at_height_max"),
			qb.GtNamed(cAtHeight, "at_height_min"),
		).
		Limit(1).
		Query(d.session).
		BindMap(qb.M{
			cPartKey:        key,
			"at_height_max": requestHeight,
			"at_height_min": requestHeightMin,
		}).
		SelectRelease(&items)

	if err != nil {
		return nil, err
	}

	// in tm-db@v0.6.4, key not found is NOT an error
	if len(items) == 0 {
		return nil, nil
	}

	if len(items[0].Key) == 0 {
		return nil, nil
	} else if items[0].Deleted {
		return nil, nil
	} else {
		if items[0].Value == nil {
			items[0].Value = []byte{}
		}
		return items[0].Value, nil
	}
}

func (d *Driver) Has(maxHeight int64, key []byte) (bool, error) {
	var requestHeight = hld.Height(maxHeight).CurrentOrLatest().ToInt64()
	var requestHeightMin = hld.Height(0).CurrentOrNever().ToInt64()

	// check if requestHeightMin is
	if requestHeightMin > requestHeight {
		return false, fmt.Errorf("invalid height")
	}

	// retrieve!
	var items []Item

	getStmt, getNames := d.table.SelectBuilder(cPartKey, cKey, cValue, cDeleted).Where(
		qb.Eq(cKey),
		qb.LtOrEqNamed(cAtHeight, "at_height_max"),
		qb.GtNamed(cAtHeight, "at_height_min"),
	).Limit(1).ToCql()

	err := d.session.Query(getStmt, getNames).
		BindMap(qb.M{
			cKey:            key,
			cPartKey:        key,
			"at_height_max": requestHeight,
			"at_height_min": requestHeightMin,
		}).
		SelectRelease(&items)

	if err != nil {
		return false, err
	}

	// in tm-db@v0.6.4, key not found is NOT an error
	if len(items[0].Key) == 0 {
		return false, nil
	} else if items[0].Deleted {
		return false, nil
	} else {
		return true, nil
	}
}

func (d *Driver) Set(atHeight int64, key, value []byte) error {
	if value == nil {
		return fmt.Errorf("invalid Value")
	}

	// should never reach here, all should be batched in tiered+hld
	item := Item{
		PartitionKey: key,
		Key:          key,
		Value:        value,
		AtHeight:     atHeight,
		Deleted:      false,
	}

	err := d.session.Query(d.table.Insert()).
		BindStruct(item).
		ExecRelease()

	if err != nil {
		return err
	} else {
		return nil
	}
}

func (d *Driver) SetSync(atHeight int64, key, value []byte) error {
	return d.Set(atHeight, key, value)
	// should never reach here, all should be batched in tiered+hld
	panic("should never reach here")
}

func (d *Driver) Delete(atHeight int64, key []byte) error {
	// should never reach here, all should be batched in tiered+hld
	item := Item{
		PartitionKey: key,
		Key:          key,
		Value:        []byte{},
		AtHeight:     atHeight,
		Deleted:      true,
	}

	err := d.session.Query(d.table.Insert()).
		BindStruct(item).
		ExecRelease()

	if err != nil {
		return err
	} else {
		return nil
	}
}

func (d *Driver) DeleteSync(atHeight int64, key []byte) error {
	return d.Delete(atHeight, key)
}

func (d *Driver) Iterator(maxHeight int64, start, end []byte) (hld.HeightLimitEnabledIterator, error) {
	it := NewCassandraIterator(d.session, d.table, maxHeight, start, end, IteratorConfiguration{
		PrefetchCount: d.iteratorPrefetchCount,
		Reverse:       false,
	})

	// load the first segment, fail if error
	if initialLoadErr := it.Load(); initialLoadErr != nil {
		return nil, initialLoadErr
	}

	return it, nil
}

func (d *Driver) ReverseIterator(maxHeight int64, start, end []byte) (hld.HeightLimitEnabledIterator, error) {
	it := NewCassandraIterator(d.session, d.table, maxHeight, start, end, IteratorConfiguration{
		PrefetchCount: d.iteratorPrefetchCount,
		Reverse:       true,
	})

	// load the first segment, fail if error
	if initialLoadErr := it.Load(); initialLoadErr != nil {
		return nil, initialLoadErr
	}

	return it, nil
}

func (d *Driver) Close() error {
	d.session.Close()
	return nil
}

func (d *Driver) NewBatch(atHeight int64) hld.HeightLimitEnabledBatch {
	return NewCassandraBatch(atHeight, d)
}

// TODO: Implement me
func (d *Driver) Print() error {
	return nil
}

func (d *Driver) Stats() map[string]string {
	return nil
}
