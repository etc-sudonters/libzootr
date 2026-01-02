package indexes

import (
	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"sudonters/libzootr/internal/table"

	"github.com/etc-sudonters/substrate/mirrors"
)

type TableHashingFunc[T comparable] func(table.Value) (T, bool)
type HashingFunc[TComponent any, TIndex comparable] func(TComponent) (TIndex, bool)

func TableHasherFrom[TComponent any, TIndex comparable](f HashingFunc[TComponent, TIndex]) TableHashingFunc[TIndex] {
	return func(v table.Value) (TIndex, bool) {
		if concrete, casted := v.(TComponent); casted {
			return f(concrete)
		}
		return mirrors.Empty[TIndex](), false
	}
}

func CreateUniqueHashIndex[TValue any, TIndex comparable](f HashingFunc[TValue, TIndex]) *UniqueHashIndex[TIndex] {
	return NewUniqueHashIndex(TableHasherFrom(f))
}

func NewUniqueHashIndex[T comparable](i TableHashingFunc[T]) *UniqueHashIndex[T] {
	return &UniqueHashIndex[T]{
		members: make(map[T]table.RowId, 8),
		hasher:  i,
	}
}

type UniqueHashIndex[TIndex comparable] struct {
	members map[TIndex]table.RowId
	hasher  TableHashingFunc[TIndex]
}

func (h *UniqueHashIndex[TIndex]) Set(e table.RowId, c table.Value) {
	idx, ok := h.hasher(c)
	if !ok {
		return
	}

	h.members[idx] = e
}

func (h *UniqueHashIndex[TIndex]) Unset(e table.RowId) {
	var key TIndex
	var found bool
	for k, row := range h.members {
		if row == e {
			key = k
			found = true
			break
		}
	}

	if found {
		delete(h.members, key)
	}
}

func (h *UniqueHashIndex[TIndex]) Rows(c table.Value) (b bitset32.Bitset) {
	idx, hashed := h.hasher(c)
	if hashed {
		entity, exists := h.members[idx]
		if exists {
			b.Set(uint32(entity))
		}
	}

	return
}
