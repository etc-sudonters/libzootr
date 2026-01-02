package table

import (
	"reflect"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
)

type Values []Value
type Value interface{}

type RowId uint32
type Row = bitset32.Bitset
type Rows []*Row

type Table struct {
	Cols    []ColumnData
	Rows    Rows
	indexes map[ColumnId]Index
	coltyp  map[reflect.Type]ColumnId
}

func New() *Table {
	return &Table{
		Cols:    make([]ColumnData, 0),
		Rows:    make(Rows, 0),
		indexes: make(map[ColumnId]Index, 0),
		coltyp:  make(map[reflect.Type]ColumnId, 0),
	}
}

func (tbl *Table) Lookup(c ColumnId, v Value) bitset32.Bitset {
	if idx, ok := tbl.indexes[c]; ok {
		return idx.Rows(v)
	}
	return tbl.Cols[c].column.ScanFor(v)
}

func (tbl *Table) InsertRow() RowId {
	id := RowId(len(tbl.Rows))
	tbl.Rows = append(tbl.Rows, &bitset32.Bitset{})
	return id
}

func (tbl *Table) SetValue(r RowId, c ColumnId, v Value) error {
	col := tbl.Cols[c]
	col.column.Set(r, v)
	row := tbl.Rows[r]
	row.Set(uint32(c))
	if idx, ok := tbl.indexes[c]; ok {
		idx.Set(r, v)
	}
	return nil
}

func (tbl *Table) UnsetValue(r RowId, c ColumnId) error {
	col := tbl.Cols[c]
	col.column.Unset(r)
	row := tbl.Rows[r]
	row.Unset(uint32(c))
	if idx, ok := tbl.indexes[c]; ok {
		idx.Unset(r)
	}
	return nil
}

func (tbl *Table) ColumnIdFor(ty reflect.Type) (ColumnId, error) {
	cid, exists := tbl.coltyp[ty]
	if !exists {
		return INVALID_COLUMNID, ErrColumnNotExists(ty.Name())
	}
	return cid, nil
}

func ColumnIdFor[T any](tbl *Table) (ColumnId, error) {
	return tbl.ColumnIdFor(reflect.TypeFor[T]())
}
