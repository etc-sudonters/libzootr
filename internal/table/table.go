package table

import (
	"errors"
	"fmt"
	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"reflect"

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
type Rows []Row
type ColumnMap map[reflect.Type]Value
type RowTuple struct {
	Id RowId
	ValueTuple
}

type Row = bitset32.Bitset

var copyRow = bitset32.Copy

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

var ColumnNotPresent = errors.New("column not present")
var CouldNotCastColumn = errors.New("could not cast column")

func Extract[T any](cm ColumnMap) (*T, error) {
	typ := mirrors.TypeOf[T]()
	item, exists := cm[typ]
	if !exists {
		return nil, fmt.Errorf("%w: '%s'", ColumnNotPresent, typ.Name())
	}
	t, casted := item.(T)
	if !casted {
		return nil, fmt.Errorf("%w: '%s'", CouldNotCastColumn, typ.Name())
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
		cols:    make([]ColumnData, 0),
		rows:    make(Rows, 0),
		indexes: make(map[ColumnId]Index, 0),
		attrs:   make(map[string]ColumnId, 0),
	}
}

type Table struct {
	cols    []ColumnData
	rows    Rows
	indexes map[ColumnId]Index
	attrs   map[string]ColumnId
}

func (this *Table) GetColumnByAttr(attr string) (ColumnData, error) {
	cid, exists := this.attrs[attr]
	if !exists {
		return ColumnData{}, fmt.Errorf("attribute %q not found", cid)
	}
	return this.Column(cid)
}

func (this *Table) MembersOfColumns(id ColumnId) (bitset32.Bitset, error) {
	if int(id) > len(this.cols) {
		return bitset32.Bitset{}, fmt.Errorf("column %d not found", id)
	}

	col := this.cols[id]
	return col.Column().Membership(), nil
}

func (this *Table) Column(id ColumnId) (ColumnData, error) {
	if int(id) > len(this.cols) {
		return ColumnData{}, fmt.Errorf("column %d not found", id)
	}

	return this.cols[id], nil
}

func (tbl *Table) Membership(r RowId) (bitset32.Bitset, error) {
	if int(r) >= len(tbl.rows) {
		return bitset32.Bitset{}, errors.New("row does not exist")
	}

	membership := tbl.rows[r]
	return bitset32.Copy(membership), nil
}

func (tbl *Table) Lookup(c ColumnId, v Value) bitset32.Bitset {
	if idx, ok := tbl.indexes[c]; ok {
		return idx.Rows(v)
	}
	return tbl.cols[c].column.ScanFor(v)
}

func (tbl *Table) CreateColumn(b *ColumnBuilder) (ColumnData, error) {
	id := ColumnId(len(tbl.cols))
	col := b.build(id)
	tbl.cols = append(tbl.cols, col)
	if b.index != nil {
		tbl.indexes[id] = b.index
	}
	tbl.attrs[b.attr] = id
	return col, nil
}

func (tbl *Table) InsertRow() RowId {
	id := RowId(len(tbl.rows))
	tbl.rows = append(tbl.rows, bitset32.Bitset{})
	return id
}

func (tbl *Table) SetValue(r RowId, c ColumnId, v Value) error {
	col := tbl.cols[c]
	col.column.Set(r, v)
	row := &tbl.rows[r]
	row.Set(uint32(c))
	if idx, ok := tbl.indexes[c]; ok {
		idx.Set(r, v)
	}
	return nil
}

func (tbl *Table) UnsetValue(r RowId, c ColumnId) error {
	col := tbl.cols[c]
	col.column.Unset(r)
	row := &tbl.rows[r]
	row.Unset(uint32(c))
	if idx, ok := tbl.indexes[c]; ok {
		idx.Unset(r)
	}
	return nil
}

func (tbl *Table) AttrTable() map[string]ColumnMetadata {
	attrs := make(map[string]ColumnMetadata, len(tbl.cols))

	for attr, id := range tbl.attrs {
		attrs[attr] = tbl.cols[id]
	}

	return attrs
}
