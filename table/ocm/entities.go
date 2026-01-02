package ocm

import (
	"iter"
	"sudonters/libzootr/table"
	"sudonters/libzootr/table/columns"

	"github.com/etc-sudonters/substrate/slipup"
)

type entity struct{}
type Entity = table.RowId
type Component = table.Value
type ComparableComponent = table.ComparableValue
type Components table.Values

func (this *Components) Add(cs ...Component) {
	*this = append(*this, cs...)
}

func DDL() []table.DDL {
	return []table.DDL{
		columns.BitColumnOf[entity],
	}
}

func NewEntities(tbl *table.Table) *Entities {
	return &Entities{tbl}
}

type Entities struct {
	tbl *table.Table
}

func (this *Entities) Proxy(e Entity) (Proxy, error) {
	if !this.tbl.IsRow(table.RowId(e)) {
		return Proxy{}, slipup.Createf("entity %v does not exist", e)
	}

	return Proxy{e: e, src: this}, nil
}

func (this *Entities) Query(qs ...table.Q) (table.ResultSetIter, error) {
	return table.Query(this.tbl, table.Exists[entity], qs...)
}

func KeyedEntities[Key Component](ents *Entities, qs ...table.Q) (iter.Seq2[Key, Entity], error) {
	qs = append(qs, table.Load[Key])
	keyIdx := len(qs) - 1
	entities, err := ents.Query(qs...)
	if err != nil {
		return nil, err
	}

	return func(yield func(Key, Entity) bool) {
		for id, row := range entities.All {
			key := row.Values[keyIdx].(Key)
			if !yield(key, id) {
				return
			}
		}
	}, nil
}

func FindOne[K ComparableComponent](ents *Entities, key K, qs ...table.Q) (Entity, error) {
	qs = append(qs, table.Exists[entity])
	id, err := table.FindOne(ents.tbl, key, qs...)
	return id, err
}

func (this *Entities) Matching(qs ...table.Q) (iter.Seq[Entity], error) {
	ids, qErr := table.QueryRowIds(this.tbl, table.Exists[entity], qs...)
	if qErr != nil {
		return nil, qErr
	}

	return func(yield func(Entity) bool) {
		for id, _ := range ids.All {
			if !yield(Entity(id)) {
				return
			}
		}

	}, nil
}

func (this *Entities) CreateEntity() (Proxy, error) {
	id, err := table.InsertRow(this.tbl, entity{})
	if err != nil {
		return Proxy{}, slipup.Describe(err, "failed to create entity")
	}

	return Proxy{e: id, src: this}, nil
}

type Proxy struct {
	e   Entity
	src *Entities
}

func (this Proxy) Entity() Entity { return this.e }

func (this Proxy) Attach(components ...Component) error {
	return table.SetRowValues(this.src.tbl, this.e, components)
}

func (this Proxy) AttachFrom(components Components) error {
	return table.SetRowValues(this.src.tbl, this.e, table.Values(components))
}

func (this Proxy) Values(components ...table.MakesColId) (table.ValueTuple, error) {
	cols := make(table.ColumnIds, len(components))
	for i, mid := range components {
		cid, err := mid(this.src.tbl)
		if err != nil {
			return table.ValueTuple{}, err
		}
		cols[i] = cid
	}
	return table.GetValues(this.src.tbl, this.e, cols)
}

func IndexedComponent[V Component](ents *Entities) (iter.Seq2[Entity, V], error) {
	rows, err := ents.Query(table.Load[V])
	if err != nil {
		return nil, err
	}

	return func(yield func(Entity, V) bool) {
		for row, tup := range rows.All {
			v := tup.Values[0].(V) // there's only our load present
			if !yield(row, v) {
				return
			}
		}
	}, nil
}
