package columns

import (
	"sudonters/libzootr/internal/table"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
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

func (m *Bit) Membership() bitset32.Bitset {
	return bitset32.Copy(*m.members)
}

func BitColumnOf[T any](attr string) *table.ColumnBuilder {
	var t T
	return BitColumnUsing(attr, t)
}

func BitColumnUsing[T any](attr string, t T) *table.ColumnBuilder {
	col := table.BuildColumnOf[T](attr, NewBit(t))
	col.Kind = table.ColumnBoolean
	return col
}

func SizedBitColumnOf[T any](attr string, capacity uint32) *table.ColumnBuilder {
	var t T
	return SizedBitColumnUsing(attr, t, capacity)
}

func SizedBitColumnUsing[T any](attr string, t T, capacity uint32) *table.ColumnBuilder {
	col := table.BuildColumnOf[T](attr, NewSizedBit(t, capacity))
	col.Kind = table.ColumnBoolean
	return col
}
