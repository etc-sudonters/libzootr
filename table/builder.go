package table

import (
	"fmt"
	"reflect"

	"github.com/etc-sudonters/substrate/slipup"
)

type ErrColumnExists string

func (this ErrColumnExists) Error() string {
	return fmt.Sprintf("column for type %q already exists", string(this))
}

type ErrColumnNotExists string

func (this ErrColumnNotExists) Error() string {
	return fmt.Sprintf("column for type %q does not exist", string(this))
}

func FromDDL(ddls ...DDL) (*Table, error) {
	build := builder{New()}
	for _, ddl := range ddls {
		if _, err := build.WithColumn(ddl()); err != nil {
			return nil, slipup.Describe(err, "failed to apply DDL")
		}
	}

	return build.tbl, nil
}

type builder struct {
	tbl *Table
}

func (this *builder) WithColumn(b *ColumnBuilder) (ColumnData, error) {
	coltyp := b.Type()
	if _, ok := this.tbl.coltyp[coltyp]; ok {
		return ColumnData{}, ErrColumnExists(coltyp.Name())
	}

	id := ColumnId(len(this.tbl.Cols))
	col := b.build(id)
	this.tbl.Cols = append(this.tbl.Cols, col)
	this.tbl.coltyp[coltyp] = id
	if b.index != nil {
		this.tbl.indexes[id] = b.index
	}
	return col, nil
}

func BuildColumn(col Column, typ reflect.Type) *ColumnBuilder {
	if col == nil {
		panic("nil column")
	}

	if typ == nil {
		panic("nil type information")
	}

	b := new(ColumnBuilder)
	b.column = col
	b.typ = typ
	return b
}

func BuildColumnOf[T Value](col Column) *ColumnBuilder {
	return BuildColumn(col, reflect.TypeFor[T]())
}

type DDL func() *ColumnBuilder

type ColumnBuilder struct {
	typ    reflect.Type
	column Column
	index  Index
}

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
	return ColumnData{
		id:     id,
		typ:    c.typ,
		column: c.column,
	}
}
