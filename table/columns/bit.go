package columns

import (
	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"sudonters/libzootr/table"
)

func NewBit(singleton table.Value) *Bit {
	return &Bit{t: singleton, members: &bitset32.Bitset{}}
}

func NewSizedBit(singleton table.Value, capacity uint32) *Bit {
	members := bitset32.WithBucketsFor(capacity)
	return &Bit{t: singleton, members: &members}
}

/*
 * Column backed by a bitset, consequently rows stored in this column do not
 * express unique values. Instead the presence of a row is handled by a
 * singleton value set on the column.
 */
type Bit struct {
	t       table.Value
	members *bitset32.Bitset
}

func (m Bit) Get(e table.RowId) table.Value {
	if m.members.IsSet(uint32(e)) {
		return m.t
	}
	return nil
}

func (m *Bit) Set(e table.RowId, c table.Value) {
	m.members.Set(uint32(e))
}

func (m *Bit) Unset(e table.RowId) {
	m.members.Unset(uint32(e))
}

func (m *Bit) ScanFor(c table.Value) bitset32.Bitset {
	return bitset32.Copy(*m.members)
}

func (m *Bit) Len() int {
	return m.members.Len()
}

func BitColumnOf[T any]() *table.ColumnBuilder {
	var t T
	return BitColumnUsing(t)
}

func BitColumnUsing[T any](t T) *table.ColumnBuilder {
	return table.BuildColumnOf[T](NewBit(t))
}

func SizedBitColumnOf[T any](capacity uint32) *table.ColumnBuilder {
	var t T
	return SizedBitColumnUsing(t, capacity)
}

func SizedBitColumnUsing[T any](t T, capacity uint32) *table.ColumnBuilder {
	return table.BuildColumnOf[T](NewSizedBit(t, capacity))
}
