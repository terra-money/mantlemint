package hld

import "github.com/terra-money/mantlemint-provider-v0.34.x/db/common"

var _ common.Iterator = (*HeightLimitedDBIterator)(nil)

type HeightLimitedDBIterator struct {
	oit common.Iterator
	atHeight int64
}

func NewHeightLimitedIterator(atHeight int64, oit common.Iterator) common.Iterator {
	return &HeightLimitedDBIterator{
		oit: oit,
		atHeight: atHeight,
	}
}

func (h *HeightLimitedDBIterator) Domain() (start []byte, end []byte) {
	// TODO: fix me
	return h.oit.Domain()
}

func (h *HeightLimitedDBIterator) Valid() bool {
	return h.oit.Valid()
}

func (h *HeightLimitedDBIterator) Next() {
	h.oit.Next()
}

func (h *HeightLimitedDBIterator) Key() (key []byte) {
	return h.oit.Key()[:len(key)-9]
}

func (h *HeightLimitedDBIterator) Value() (value []byte) {
	return h.oit.Value()
}

func (h *HeightLimitedDBIterator) Error() error {
	return h.oit.Error()
}

func (h *HeightLimitedDBIterator) Close() error {
	return h.oit.Close()
}
