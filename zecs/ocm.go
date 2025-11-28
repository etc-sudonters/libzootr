package zecs

import (
	"errors"
	"fmt"
	"sudonters/libzootr/internal/bundle"
	"sudonters/libzootr/internal/query"
	"sudonters/libzootr/internal/table"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
)

type Entity = table.RowId
type DDL = table.DDL
type Value = table.Value
type Values = table.Values
type RowSet = bundle.Interface

func New() (Ocm, error) {
	eng, err := query.NewEngine()
	return Ocm{eng}, err
}

func Apply(ocm *Ocm, ddl []DDL) error {
	for i := range ddl {
		if _, err := ocm.eng.CreateColumn(ddl[i]()); err != nil {
			return err
		}
	}

	return nil
}

type makesColId func(*Ocm) table.ColumnId

func Get[T Value](ocm *Ocm) table.ColumnId {
	return query.MustAsColumnId[T](ocm.eng)
}

type Ocm struct {
	eng query.Engine
}

func (this *Ocm) GetValues(which Entity, cols ...makesColId) (Values, error) {
	if len(cols) == 0 {
		return nil, errors.New("no columns provided")
	}

	cids := make(table.ColumnIds, len(cols))
	for i := range cols {
		cids[i] = cols[i](this)
	}

	tup, err := this.eng.GetValues(which, cids)
	return tup.Values, err
}

func (this *Ocm) Proxy(which Entity) Proxy {
	return Proxy{which, this}
}

func (this *Ocm) Engine() query.Engine {
	return this.eng
}

func (this *Ocm) Query() Q {
	return Q{this, this.eng.CreateQuery(), query.RetrieveOptions{}}
}

type Q struct {
	set  *Ocm
	q    query.Query
	opts query.RetrieveOptions
}

func (this *Q) Configure(opt func(*query.RetrieveOptions)) {
	opt(&this.opts)
}

func (this *Q) Build(build BuildQuery, builds ...BuildQuery) *Q {
	build(this)
	for _, b := range builds {
		b(this)
	}
	return this
}

type BuildQuery func(*Q)

func FromSubset(subset *bitset32.Bitset) BuildQuery {
	return func(this *Q) {
		this.q.FromSubset(subset)
	}
}

func Optional[T Value](this *Q) {
	this.q.Optional(query.MustAsColumnId[T](this.set.eng))
}

func Load[T Value](this *Q) {
	this.q.Load(query.MustAsColumnId[T](this.set.eng))
}

func With[T Value](this *Q) {
	this.q.Exists(query.MustAsColumnId[T](this.set.eng))
}

func WithOut[T Value](this *Q) {
	this.q.NotExists(query.MustAsColumnId[T](this.set.eng))
}

func (this *Q) Execute() (RowSet, error) {
	return this.set.eng.Retrieve(this.q)
}

func (this *Q) Rows(yield bundle.RowIter) {
	rows, err := this.Execute()
	if err != nil {
		panic(err)
	}
	rows.All(yield)
}

func Bitset32Matching(ocm *Ocm, match ...BuildQuery) bitset32.Bitset {
	if len(match) == 0 {
		panic(errors.New("no entities specified"))
	}

	q := ocm.Query()
	q.Configure(func(config *query.RetrieveOptions) {
		config.Bundler = bundle.BundleRowsOnly
	})
	q.Build(match[0], match[1:]...)
	rows, err := q.Execute()
	if err != nil {
		panic(err)
	}
	entities := bitset32.WithBucketsFor(uint32(rows.Len()))

	for row, _ := range rows.All {
		bitset32.Set(&entities, row)
	}

	return entities

}

func SliceMatching(ocm *Ocm, match ...BuildQuery) []Entity {
	if len(match) == 0 {
		panic(errors.New("no entities specified"))
	}

	q := ocm.Query()
	q.Configure(func(config *query.RetrieveOptions) {
		config.Bundler = bundle.BundleRowsOnly
	})
	q.Build(match[0], match[1:]...)
	rows, err := q.Execute()
	if err != nil {
		panic(err)
	}
	entities := make([]Entity, 0, rows.Len())

	for row, _ := range rows.All {
		entities = append(entities, row)
	}

	return entities
}

func IndexEntities[Key ComparableValue](ocm *Ocm, query ...BuildQuery) map[Key]Entity {
	q := ocm.Query()
	q.Build(Load[Key], query...)
	rows, err := q.Execute()
	if err != nil {
		panic(err)
	}
	entities := make(map[Key]Entity, rows.Len())

	for row, tup := range rows.All {
		key := tup.Values[0].(Key)
		entities[key] = row
	}

	return entities
}

func FindOne[T ComparableValue](ocm *Ocm, key T, query ...BuildQuery) Entity {
	q := ocm.Query()
	q.Build(Load[T], query...)
	for row, tup := range q.Rows {
		k := tup.Values[0].(T)
		if k == key {
			return row
		}
	}

	panic(fmt.Errorf("did not find matching entity for key %#v", key))
}

func IndexValue[V Value](ocm *Ocm, query ...BuildQuery) map[Entity]V {
	q := ocm.Query()
	q.Build(Load[V], query...)
	rows, err := q.Execute()
	if err != nil {
		panic(err)
	}
	lookup := make(map[Entity]V, rows.Len())

	for row, tup := range rows.All {
		lookup[row] = tup.Values[0].(V)
	}

	return lookup
}
