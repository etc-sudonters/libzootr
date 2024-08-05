package query

import (
	"errors"
	"fmt"
	"math"
	"reflect"
	"sudonters/zootler/internal/bundle"
	"sudonters/zootler/internal/table"
	"sudonters/zootler/internal/table/columns"

	"github.com/etc-sudonters/substrate/skelly/bitset"
)

var _ Engine = (*engine)(nil)

var ErrInvalidQuery = errors.New("query is not supported")
var ErrColumnExists = errors.New("column exists already")
var ErrColumnNotExist = errors.New("column does not exist")

func errNotExist(t reflect.Type) error {
	return fmt.Errorf("%w: %s", ErrColumnNotExist, t.Name())
}

func errExists(t reflect.Type) error {
	return fmt.Errorf("%w: %s", ErrColumnExists, t.Name())
}

func errNotExists(t reflect.Type) error {
	return fmt.Errorf("%w: %s", ErrColumnNotExist, t.Name())
}

type columnIndex map[reflect.Type]table.ColumnId
type Entry struct{}

type Query interface {
	Exists(typ reflect.Type)
	NotExists(typ reflect.Type)
	Load(typ reflect.Type)
}

type Lookup interface {
	Lookup(v table.Value)
	Load(typ reflect.Type)
}

type queryCore struct {
	errs  []error
	types columnIndex
	load  *bitset.Bitset64
}

func (b *queryCore) Load(typ reflect.Type) {
	b.set(typ, b.load)
}

func (b *queryCore) set(typ reflect.Type, s *bitset.Bitset64) {
	if id, ok := b.types[typ]; ok {
		s.Set(uint64(id))
	} else {
		b.errs = append(b.errs, fmt.Errorf("%s: %w", typ.Name(), ErrColumnNotExist))
	}
}

type lookup struct {
	queryCore
	lookups []table.Value
}

func (l *lookup) Lookup(v table.Value) {
	l.lookups = append(l.lookups, v)
}

type builder struct {
	queryCore
	exists    *bitset.Bitset64
	notExists *bitset.Bitset64
}

func (b *builder) Exists(typ reflect.Type) {
	b.set(typ, b.exists)
}

func (b *builder) NotExists(typ reflect.Type) {
	b.set(typ, b.notExists)
}

func build(b *builder) predicate {
	return predicate{
		exists:    b.exists.Union(*b.load),
		notExists: bitset.Copy(*b.notExists),
	}
}

type predicate struct {
	exists    bitset.Bitset64
	notExists bitset.Bitset64
}

func (p predicate) admit(row *bitset.Bitset64) bool {
	if !p.exists.Intersect(*row).Eq(p.exists) {
		return false
	}

	return row.Difference(p.notExists).Eq(*row)
}

type Engine interface {
	CreateQuery() Query
	CreateLookup() Lookup
	CreateColumn(c *table.ColumnBuilder) (table.ColumnId, error)
	CreateColumnIfNotExists(c *table.ColumnBuilder) (table.ColumnId, error)
	InsertRow(vs ...table.Value) (table.RowId, error)
	Retrieve(b Query) (bundle.Interface, error)
	Lookup(l Lookup) (bundle.Interface, error)
	SetValues(r table.RowId, vs table.Values) error
	UnsetValues(r table.RowId, cs table.ColumnIds) error
}

func ExtractTable(e Engine) (*table.Table, error) {
	if eng, ok := e.(*engine); ok {
		return eng.tbl, nil
	}

	return nil, errors.ErrUnsupported
}

func NewEngine() (*engine, error) {
	eng := &engine{
		columnIndex: columnIndex{nil: 0},
		tbl:         table.New(),
	}

	if _, err := eng.CreateColumn(columns.BitColumnOf[Entry]()); err != nil {
		return nil, err
	}

	return eng, nil
}

type engine struct {
	columnIndex map[reflect.Type]table.ColumnId
	tbl         *table.Table
}

func (e engine) CreateQuery() Query {
	return &builder{
		queryCore: queryCore{
			types: e.columnIndex,
			errs:  nil,
			load:  &bitset.Bitset64{},
		},
		exists:    &bitset.Bitset64{},
		notExists: &bitset.Bitset64{},
	}
}

func (e engine) CreateLookup() Lookup {
	return &lookup{
		queryCore: queryCore{
			types: e.columnIndex,
			errs:  nil,
			load:  &bitset.Bitset64{},
		},
		lookups: nil,
	}
}

func (e *engine) CreateColumn(c *table.ColumnBuilder) (table.ColumnId, error) {
	if _, ok := e.columnIndex[c.Type()]; ok {
		return 0, errExists(c.Type())
	}

	col, err := e.tbl.CreateColumn(c)
	if err != nil {
		return 0, err
	}

	e.columnIndex[col.Type()] = col.Id()
	return col.Id(), nil
}

func (e *engine) CreateColumnIfNotExists(c *table.ColumnBuilder) (table.ColumnId, error) {
	id, ok := e.columnIndex[c.Type()]
	if ok {
		return id, nil
	}

	return e.CreateColumn(c)
}

func (e *engine) InsertRow(vs ...table.Value) (table.RowId, error) {
	id := e.tbl.InsertRow()
	if len(vs) > 0 {
		e.SetValues(id, vs)
	}
	return id, nil
}

func (e engine) Retrieve(b Query) (bundle.Interface, error) {
	q, ok := b.(*builder)
	if !ok {
		return nil, fmt.Errorf("%T: %w", b, ErrInvalidQuery)
	}

	if q.errs != nil {
		return nil, errors.Join(q.errs...)
	}

	predicate := build(q)
	fill := bundle.Fill{}

	for row, possessed := range e.tbl.Rows {
		if predicate.admit(possessed) {
			fill.Set(uint64(row))
		}
	}

	var columns table.Columns
	for _, col := range q.load.Elems() {
		columns = append(columns, e.tbl.Cols[col])
	}

	if fill.Len() != 1 {
		return bundle.RowOrdered(fill, columns), nil
	} else {
		return bundle.SingleRow(fill, columns)
	}
}

func saturatedSet(numBuckets uint64) bitset.Bitset64 {
	buckets := make([]uint64, numBuckets)
	for i := range buckets {
		buckets[i] = math.MaxUint64
	}
	return bitset.FromRaw(buckets...)
}

func (e engine) Lookup(l Lookup) (bundle.Interface, error) {
	L, ok := l.(*lookup)
	if !ok {
		return nil, fmt.Errorf("%T: %w", l, ErrInvalidQuery)
	}

	if L.errs != nil {
		return nil, errors.Join(L.errs...)
	}

	entities := saturatedSet(uint64(len(e.tbl.Rows) / 64))

	for _, lookup := range L.lookups {
		colId, colIdErr := e.intoColId(lookup)
		if colIdErr != nil {
			return nil, colIdErr
		}
		rows := e.tbl.Lookup(colId, lookup)
		entities = entities.Intersect(rows)
	}

	fill := bundle.Filled(entities)
	var columns table.Columns
	for _, col := range L.load.Elems() {
		columns = append(columns, e.tbl.Cols[col])
	}

	if fill.Len() != 1 {
		return bundle.RowOrdered(fill, columns), nil
	} else {
		return bundle.SingleRow(fill, columns)
	}
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

		id, err := e.intoColId(v)
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

func (e *engine) UnsetValues(r table.RowId, cs table.ColumnIds) error {
	if len(cs) == 0 {
		return nil
	}

	for _, id := range cs {
		e.tbl.UnsetValue(r, id)
	}

	return nil
}

func (e *engine) intoColId(v table.Value) (table.ColumnId, error) {
	typ := reflect.TypeOf(v)
	id, ok := e.columnIndex[typ]
	if !ok {
		return 0, errNotExists(typ)
	}

	return id, nil
}
