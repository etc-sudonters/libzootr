package table

import (
	"math"
	"reflect"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
)

var INVALID_ROWID RowId = math.MaxUint32
var INVALID_COLUMNID ColumnId = math.MaxUint32

type ColumnMeta struct {
	Id ColumnId
	T  reflect.Type
}

type ColumnMetas []ColumnMeta
type ColumnIds []ColumnId
type Columns []ColumnData
type ColumnMap map[reflect.Type]Value
type ColumnId uint32
type ColumnFactory func() Column

// core column interface
type Column interface {
	Get(e RowId) Value
	Set(e RowId, c Value)
	Unset(e RowId)
	ScanFor(Value) bitset32.Bitset
	Len() int
}

type ColumnMetadata interface {
	Type() reflect.Type
	Id() ColumnId
}

type ColumnData struct {
	id     ColumnId
	typ    reflect.Type
	column Column
}

func (c ColumnData) Column() Column {
	return c.column
}

func (c ColumnData) Type() reflect.Type {
	return c.typ
}

func (c ColumnData) Id() ColumnId {
	return c.id
}
