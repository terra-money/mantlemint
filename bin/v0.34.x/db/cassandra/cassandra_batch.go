package cassandra

import (
	"errors"
	"fmt"
	"github.com/gocql/gocql"
	"github.com/terra-money/mantlemint-provider-v0.34.x/db/hld"
)

var _ hld.HeightLimitEnabledBatch = (*Batch)(nil)

const (
	opSet = iota
	opDelete
)

type Operation struct {
	opType  int
	opKey   []byte
	opValue []byte
}

type Batch struct {
	atHeight int64
	driver   *Driver
	ops      []Operation
}

func NewCassandraBatch(atHeight int64, driver *Driver) *Batch {
	return &Batch{
		atHeight: atHeight,
		driver:   driver,
		ops:      []Operation{},
	}
}

func (b *Batch) Set(key, value []byte) error {
	if err := invariant(b); err != nil {
		return err
	} else {
		if value == nil {
			return fmt.Errorf("asdfasd")
		}
		b.ops = append(b.ops, Operation{
			opType:  opSet,
			opKey:   key,
			opValue: value,
		})

		return nil
	}
}

func (b *Batch) Delete(key []byte) error {
	if err := invariant(b); err != nil {
		panic(err)
	} else {
		b.ops = append(b.ops, Operation{
			opType: opDelete,
			opKey:  key,
		})

		return nil
	}
}

func (b *Batch) Write() error {
	driverBatch := b.driver.session.NewBatch(gocql.LoggedBatch)

	// iterate through ops in reverse order.
	// usually in cosmos-sdk deletes on the same key happens BEFORE it's reset in the same block.
	// hence process batch ops in backwards, and prevent re-deletes in case it is reset in the same block
	deletePreventCache := make(map[string]bool)

	// create radix tree for the insertion target
	//batchInsertRadixTree := NewCassandraPrevNextGenerator(b.ops)

	for i := len(b.ops) - 1; i >= 0; i-- {
		op := b.ops[i]

		cacheKey := string(op.opKey)

		switch op.opType {
		case opSet:
			deletePreventCache[cacheKey] = true

			stmt, _ := b.driver.table.Insert()
			driverBatch.Entries = append(driverBatch.Entries, gocql.BatchEntry{
				Stmt:       stmt,
				Args:       []interface{}{op.opKey, op.opKey, op.opValue, false, b.atHeight},
				Idempotent: true,
			})

		case opDelete:
			// don't re-delete an entity that is already re-set in this block
			if _, exists := deletePreventCache[cacheKey]; exists {
				continue
			}

			stmt, _ := b.driver.table.Insert()
			driverBatch.Entries = append(driverBatch.Entries, gocql.BatchEntry{
				Stmt:       stmt,
				Args:       []interface{}{op.opKey, op.opKey, op.opValue, true, b.atHeight},
				Idempotent: true,
			})

		default:
			return errors.New("unknown cassandra batch type")
		}
	}

	defer b.Close()

	return b.driver.session.ExecuteBatch(driverBatch)
}

func (b *Batch) WriteSync() error {
	return b.Write()
}

func (b *Batch) Close() error {
	b.ops = nil
	return nil
}

func invariant(b *Batch) error {
	if b.ops == nil {
		b.ops = []Operation{}

		fmt.Printf("cassandra_batch: batch(%v) is closed", b)
		return nil
	} else {
		return nil
	}
}
