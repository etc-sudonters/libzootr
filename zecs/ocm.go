package zecs

import (
	"errors"
	"fmt"
	"sudonters/libzootr/internal/query"
	"sudonters/libzootr/internal/table"
)

type Entity = table.RowId
type DDL = table.DDL
type Value = table.Value
type Values = table.Values
type RowSet = table.ResultSetIter

func New(tbl *table.Table) Ocm {
	return Ocm{query.EngineFromTable(tbl)}
}

type Ocm struct {
	eng query.Engine
}

func (this *Ocm) Proxy(which Entity) Proxy {
	return Proxy{which, this}
}

func (this *Ocm) Engine() query.Engine {
	return this.eng
}

func (this *Ocm) Query() Q {
	return Q{this, this.eng.CreateQuery()}
}

type Q struct {
	set *Ocm
	q   query.Query
}

func (this *Q) Build(build BuildQuery, builds ...BuildQuery) *Q {
	build(this)
	for _, b := range builds {
		b(this)
	}
	return this
}

type BuildQuery func(*Q)

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

func (this *Q) Rows(yield table.YieldRow) {
	rows, err := this.Execute()
	if err != nil {
		panic(err)
	}
	rows.All(yield)
}

func EntitiesMatching(ocm *Ocm, query ...BuildQuery) []Entity {
	if len(query) == 0 {
		panic(errors.New("no entities specified"))
	}

	q := ocm.Query()
	q.Build(query[0], query[1:]...)
	rows, err := q.Execute()
	if err != nil {
		panic(err)
	}
	ptrs := make([]Entity, 0, rows.Len())

	for row, _ := range rows.All {
		ptrs = append(ptrs, row)
	}

	return ptrs
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
