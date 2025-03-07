package table

import (
	"fmt"
	"math"
	"reflect"

	"github.com/etc-sudonters/substrate/skelly/bitset32"

	"github.com/etc-sudonters/substrate/mirrors"
)

var INVALID_ROWID RowId = math.MaxUint32
var INVALID_COLUMNID ColumnId = math.MaxUint32

type RowId uint32
type ColumnId uint32
type Value interface{}

type ColumnFactory func() Column

// core column interface
type Column interface {
	Membership() bitset32.Bitset
	Get(e RowId) Value
	Set(e RowId, c Value)
	Unset(e RowId)
	ScanFor(Value) bitset32.Bitset
	Len() int
}

type ColumnMetadata interface {
	Type() reflect.Type
	Id() ColumnId
	Kind() ColumnKind
}

type ColumnData struct {
	id     ColumnId
	typ    reflect.Type
	column Column
	kind   ColumnKind
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

func (c ColumnData) Kind() ColumnKind {
	return c.kind
}

func BuildColumn(attr string, col Column, typ reflect.Type) *ColumnBuilder {
	if col == nil {
		panic("nil column")
	}

	if typ == nil {
		panic("nil type information")
	}

	b := new(ColumnBuilder)
	b.column = col
	b.typ = typ
	b.attr = attr
	return b
}

func BuildColumnOf[T Value](attr string, col Column) *ColumnBuilder {
	return BuildColumn(attr, col, mirrors.TypeOf[T]())
}

type DDL func() *ColumnBuilder

type ColumnBuilder struct {
	attr   string
	typ    reflect.Type
	column Column
	index  Index
	Kind   ColumnKind
}

type ColumnKind uint8

const (
	ColumnUnknown ColumnKind = iota
	ColumnInt
	ColumnUint
	ColumnFloat
	ColumnString
	ColumnBoolean
	ColumnComposite
)

func (c *ColumnBuilder) Index(i Index) *ColumnBuilder {
	if c.index != nil {
		// let me know when to support multiple indexes
		panic(fmt.Errorf("column already indexed: %s", c.typ.Name()))
	}
	c.index = i
	return c
}

func (c ColumnBuilder) Type() reflect.Type {
	return c.typ
}

func (c ColumnBuilder) build(id ColumnId) ColumnData {
	if c.Kind == ColumnUnknown {
		c.Kind = reflectKindToColumnKind(c.typ.Kind())
	}

	return ColumnData{
		id:     id,
		typ:    c.typ,
		column: c.column,
		kind:   c.Kind,
	}
}

func reflectKindToColumnKind(kind reflect.Kind) ColumnKind {
	switch kind {
	case reflect.Bool:
		return ColumnBoolean
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return ColumnInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return ColumnUint
	case reflect.Float32, reflect.Float64:
		return ColumnFloat
	case reflect.String:
		return ColumnString
	case reflect.Array, reflect.Map, reflect.Interface, reflect.Slice, reflect.Struct:
		return ColumnComposite
	default:
		return ColumnUnknown
	}
}
