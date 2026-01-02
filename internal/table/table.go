package table

import (
	"errors"
	"fmt"
	"reflect"
	"sudonters/libzootr/internal/skelly/bitset32"

	"github.com/etc-sudonters/substrate/mirrors"
)

type ColumnMeta struct {
	Id ColumnId
	T  reflect.Type
}
type ColumnMetas []ColumnMeta
type ColumnIds []ColumnId
type Columns []ColumnData
type Values []Value
type Row = bitset32.Bitset
type Rows []*Row
type ColumnMap map[reflect.Type]Value
type RowTuple struct {
	Id RowId
	ValueTuple
}

func (vt *ValueTuple) Init(cs Columns) {
	vt.Cols = make(ColumnMetas, len(cs))
	vt.Values = make(Values, len(cs))
	for i, c := range cs {
		vt.Cols[i].Id = c.Id()
		vt.Cols[i].T = c.Type()
	}
}

func (vt *ValueTuple) Load(r RowId, cs Columns) {
	for i, c := range cs {
		vt.Values[i] = c.Column().Get(r)
	}
}

var ErrColumnNotPresent = errors.New("column not present")
var ErrCouldNotCastColumn = errors.New("could not cast column")

func Extract[T any](cm ColumnMap) (*T, error) {
	typ := mirrors.TypeOf[T]()
	item, exists := cm[typ]
	if !exists {
		return nil, fmt.Errorf("%w: '%s'", ErrColumnNotPresent, typ.Name())
	}
	t, casted := item.(T)
	if !casted {
		return nil, fmt.Errorf("%w: '%s'", ErrCouldNotCastColumn, typ.Name())
	}
	return &t, nil
}

type ValueTuple struct {
	Cols   ColumnMetas
	Values Values
}

func FromColumnMap[T any](cm ColumnMap) (T, bool) {
	t, exists := cm[mirrors.T[T]()]
	if !exists {
		return mirrors.Empty[T](), false
	}
	return t.(T), true
}

func (v *ValueTuple) ColumnMap() ColumnMap {
	m := make(ColumnMap, len(v.Values))

	for i := range v.Cols {
		m[v.Cols[i].T] = v.Values[i]
	}

	return m
}

func New() *Table {
	return &Table{
		Cols:    make([]ColumnData, 0),
		Rows:    make(Rows, 0),
		indexes: make(map[ColumnId]Index, 0),
	}
}

type Table struct {
	Cols    []ColumnData
	Rows    Rows
	indexes map[ColumnId]Index
}

func (tbl *Table) Lookup(c ColumnId, v Value) bitset32.Bitset {
	if idx, ok := tbl.indexes[c]; ok {
		return idx.Rows(v)
	}
	return tbl.Cols[c].column.ScanFor(v)
}

func (tbl *Table) CreateColumn(b *ColumnBuilder) (ColumnData, error) {
	id := ColumnId(len(tbl.Cols))
	col := b.build(id)
	tbl.Cols = append(tbl.Cols, col)
	if b.index != nil {
		tbl.indexes[id] = b.index
	}
	return col, nil
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
