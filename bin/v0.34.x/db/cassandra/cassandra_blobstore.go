package cassandra

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
	"github.com/scylladb/gocqlx/v2/table"
)

type BlobStoreDriver struct {
	session   gocqlx.Session
	table     *table.Table
	tableName string
}

func NewBlobStoreCassandraDriver(config *BlobStoreDriverConfig) *BlobStoreDriver {

	fmt.Println(config)

	cluster := gocql.NewCluster(config.Clusters...)
	cluster.Keyspace = config.Keyspace
	cluster.Consistency = gocql.LocalQuorum

	// create session and wrap in gocqlx
	session, err := gocqlx.WrapSession(cluster.CreateSession())

	if err != nil {
		panic(err)
	}

	tableModel := table.Metadata{
		Name:    config.TableName,
		Columns: blobStoreColumns,
		PartKey: blobStorePartKey,
	}

	if err := session.Query(fmt.Sprintf("create table if not exists %s (partition_key blob, data blob, primary key(partition_key))", config.TableName), []string{}).ExecRelease(); err != nil {
		panic(err)
	}

	table := table.New(tableModel)

	return &BlobStoreDriver{
		session:   session,
		tableName: config.TableName,
		table:     table,
	}
}

func (d *BlobStoreDriver) Get(key []byte) ([]byte, error) {
	var blobs [1]Blob

	getStmt, getNames := d.table.SelectBuilder(cPartKey, cData).Where(
		qb.Eq(cPartKey),
	).Limit(1).ToCql()

	err := d.session.Query(getStmt, getNames).
		BindMap(qb.M{
			cPartKey: key,
		}).
		SelectRelease(&blobs)

	if err != nil {
		return nil, err
	}

	if len(blobs[0].PartitionKey) == 0 {
		return nil, fmt.Errorf("blob not found")
	} else {
		return blobs[0].Data, nil
	}
}

func (d *BlobStoreDriver) Set(key []byte, data []byte) error {
	blob := Blob{
		PartitionKey: key,
		Data:         data,
	}

	err := d.session.Query(d.table.Insert()).
		BindStruct(blob).
		ExecRelease()

	if err != nil {
		return err
	} else {
		return nil
	}
}

func (d *BlobStoreDriver) Close() error {
	d.session.Close()
	return nil
}
