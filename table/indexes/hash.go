package indexes

import (
	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"sudonters/libzootr/table"
)

type hashbitmap[T comparable] map[T]bitset32.Bitset

func (h hashbitmap[T]) set(key T, which uint32) {
	members := h.membersetfor(key)
	members.Set(which)
	h[key] = members
}

func (h hashbitmap[T]) unset(which uint32) {
	for key, members := range h {
		if members.IsSet(which) {
			(&members).Unset(which)
			h[key] = members
			break
		}
	}
}

func (h hashbitmap[T]) membersetfor(key T) bitset32.Bitset {
	members, ok := h[key]
	if !ok {
		members = bitset32.Bitset{}
	}

	return members
}

type HashIndex[T comparable] struct {
	members hashbitmap[T]
	hasher  TableHashingFunc[T]
}

func CreateHashIndex[TComponent any, TIndex comparable](f HashingFunc[TComponent, TIndex]) *HashIndex[TIndex] {
	return NewHashIndex(TableHasherFrom(f))
}

func NewHashIndex[T comparable](f TableHashingFunc[T]) *HashIndex[T] {
	return &HashIndex[T]{
		members: make(hashbitmap[T], 8),
		hasher:  f,
	}
}

func (h *HashIndex[T]) Set(r table.RowId, v table.Value) {
	idx, ok := h.hasher(v)
	if !ok {
		return
	}
	h.members.set(idx, uint32(r))
}

func (h *HashIndex[T]) Unset(r table.RowId) {
	h.members.unset(uint32(r))
}

// this bitset is intersected / & / AND'd
func (h *HashIndex[T]) Rows(v table.Value) bitset32.Bitset {
	idx, ok := h.hasher(v)
	if !ok {
		return bitset32.Bitset{}
	}
	return bitset32.Copy(h.members.membersetfor(idx))
}
