package query

import (
	"errors"
	"fmt"
	"reflect"
	"sudonters/libzootr/internal/bundle"
	"sudonters/libzootr/internal/table"

	"github.com/etc-sudonters/substrate/skelly/bitset32"

	"github.com/etc-sudonters/substrate/slipup"
)

var _ Engine = (*engine)(nil)

var ErrInvalidQuery = errors.New("query is not supported")

type Entry struct{}

type Query interface {
	Optional(table.ColumnId)
	Load(table.ColumnId)
	Exists(table.ColumnId)
	NotExists(table.ColumnId)
}

type query struct {
	load      *bitset32.Bitset
	cols      table.ColumnIds
	exists    *bitset32.Bitset
	notExists *bitset32.Bitset
	optional  *bitset32.Bitset
}

func (b *query) Load(typ table.ColumnId) {
	if b.load.Set(uint32(typ)) {
		b.cols = append(b.cols, typ)
	}
}

func (b *query) Exists(typ table.ColumnId) {
	b.exists.Set(uint32(typ))
}

func (b *query) NotExists(typ table.ColumnId) {
	b.notExists.Set(uint32(typ))
}

func (b *query) Optional(typ table.ColumnId) {
	i := uint32(typ)
	if b.optional.Set(i) && !b.load.IsSet(i) {
		b.cols = append(b.cols, typ)
	}
}

func makePredicate(b *query) predicate {
	return predicate{
		exists:    b.exists.Union(*b.load),
		notExists: bitset32.Copy(*b.notExists),
	}
}

type predicate struct {
	exists    bitset32.Bitset
	notExists bitset32.Bitset
}

func (p predicate) admit(row *bitset32.Bitset) bool {
	if !p.exists.Intersect(*row).Eq(p.exists) {
		return false
	}

	return row.Difference(p.notExists).Eq(*row)
}

type Engine interface {
	CreateQuery() Query
	InsertRow(vs ...table.Value) (table.RowId, error)
	Retrieve(b Query) (bundle.Interface, error)
	GetValues(r table.RowId, cs table.ColumnIds) (table.ValueTuple, error)
	SetValues(r table.RowId, vs table.Values) error
	UnsetValues(r table.RowId, cs table.ColumnIds) error
	ColumnIdFor(reflect.Type) (table.ColumnId, bool)
}

func MustColumnIdFor(typ reflect.Type, e Engine) table.ColumnId {
	id, ok := e.ColumnIdFor(typ)
	if ok {
		return id
	}

	panic(slipup.Createf("did not have column for '%s'", typ.Name()))
}

func MustAsColumnId[T any](e Engine) table.ColumnId {
	return MustColumnIdFor(reflect.TypeFor[T](), e)
}

func ExtractTable(e Engine) (*table.Table, error) {
	if eng, ok := e.(*engine); ok {
		return eng.tbl, nil
	}

	return nil, errors.ErrUnsupported
}

func EngineFromTable(tbl *table.Table) Engine {
	return &engine{tbl}
}

type engine struct {
	tbl *table.Table
}

func (e *engine) ColumnIdFor(t reflect.Type) (table.ColumnId, bool) {
	cid, err := e.tbl.ColumnIdFor(t)
	return cid, err == nil
}

func (e engine) CreateQuery() Query {
	return &query{
		cols:      nil,
		load:      &bitset32.Bitset{},
		exists:    &bitset32.Bitset{},
		notExists: &bitset32.Bitset{},
		optional:  &bitset32.Bitset{},
	}
}

func (e *engine) InsertRow(vs ...table.Value) (table.RowId, error) {
	id := e.tbl.InsertRow()
	if len(vs) > 0 {
		if err := e.SetValues(id, vs); err != nil {
			return id, err
		}
	}
	return id, nil
}

func (e engine) Retrieve(b Query) (bundle.Interface, error) {
	q, ok := b.(*query)
	if !ok {
		return nil, fmt.Errorf("%T: %w", b, ErrInvalidQuery)
	}

	predicate := makePredicate(q)
	fill := bitset32.Bitset{}

	for row, possessed := range e.tbl.Rows {
		if predicate.admit(possessed) {
			fill.Set(uint32(row))
		}
	}

	var columns table.Columns
	for _, col := range q.cols {
		columns = append(columns, e.tbl.Cols[col])
	}

	return bundle.Bundle(fill, columns)
}

func (e *engine) SetValues(r table.RowId, vs table.Values) error {
	if len(vs) == 0 {
		return nil
	}

	ids := make(table.ColumnIds, len(vs))
	errs := make([]error, 0)

	for idx, v := range vs {
		if v == nil {
			panic("cannot insert nil value")
		}

		vty := reflect.TypeOf(v)
		id, err := e.tbl.ColumnIdFor(vty)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		ids[idx] = id
	}

	if len(errs) != 0 {
		return errors.Join(errs...)
	}

	for idx, id := range ids {
		e.tbl.SetValue(r, id, vs[idx])
	}

	return nil
}

func (e *engine) GetValues(r table.RowId, cs table.ColumnIds) (table.ValueTuple, error) {
	var vt table.ValueTuple
	vt.Cols = make(table.ColumnMetas, len(cs))
	vt.Values = make(table.Values, len(cs))
	for i, cid := range cs {
		c := e.tbl.Cols[cid]
		vt.Cols[i].Id = c.Id()
		vt.Cols[i].T = c.Type()
		vt.Values[i] = c.Column().Get(r)
	}

	return vt, nil
}

func (e *engine) UnsetValues(r table.RowId, cs table.ColumnIds) error {
	if len(cs) == 0 {
		return nil
	}

	for _, id := range cs {
		e.tbl.UnsetValue(r, id)
	}

	return nil
}
