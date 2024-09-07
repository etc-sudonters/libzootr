package entities

import (
	"github.com/etc-sudonters/substrate/slipup"
	"sudonters/zootler/internal"
	"sudonters/zootler/internal/components"
	"sudonters/zootler/internal/query"
	"sudonters/zootler/internal/table"
)

type Map[T Entity] interface {
	Entity(components.Name) (T, error)
}

type genericmap[T Entity] struct {
	cache map[internal.NormalizedStr]T
	eng   query.Engine
	def   table.Values
	fact  func(table.RowId, components.Name, table.Values) T
}

func (e *genericmap[T]) Entity(name components.Name) (T, error) {
	normaled := internal.Normalize(name)
	if ent, exists := e.cache[normaled]; exists {
		return ent, nil
	}

	ent, err := e.eng.InsertRow(name)
	if err != nil {
		var t T
		return t, err
	}

	if err := e.eng.SetValues(ent, e.def); err != nil {
		var t T
		return t, err
	}

	t := e.fact(ent, name, e.def)
	e.cache[normaled] = t
	return t, nil
}

func (e *genericmap[T]) init(f func(query.Engine, query.Query)) error {
	q := e.eng.CreateQuery()
	f(e.eng, q)
	rows, err := e.eng.Retrieve(q)
	if err != nil {
		return slipup.Describe(err, "while initializing map")
	}

	for rid, row := range rows.All {
		m := row.ColumnMap()
		name, _ := table.FromColumnMap[components.Name](m)
		e.cache[internal.Normalize(name)] = e.fact(rid, name, e.def)
	}

	return nil
}

func LocationMap(eng query.Engine) (*genericmap[Location], error) {
	return newmap(
		eng,
		func(eng query.Engine, q query.Query) {
			q.Exists(query.MustAsColumnId[components.Location](eng))
		},
		func(id table.RowId, name components.Name, _ table.Values) Location {
			var t Location
			t.rid = id
			t.name = name
			t.eng = eng
			return t
		},
		components.Location{},
	)
}

func TokenMap(eng query.Engine) (*genericmap[Token], error) {
	return newmap(
		eng,
		func(eng query.Engine, q query.Query) {
			q.Exists(query.MustAsColumnId[components.CollectableGameToken](eng))
		},
		func(id table.RowId, name components.Name, _ table.Values) Token {
			var t Token
			t.rid = id
			t.name = name
			t.eng = eng
			return t
		},
		components.CollectableGameToken{},
	)
}

func EdgeMap(eng query.Engine) (*genericmap[Edge], error) {
	return newmap(
		eng,
		func(eng query.Engine, q query.Query) {
			q.Exists(query.MustAsColumnId[components.Edge](eng))
		},
		func(id table.RowId, name components.Name, _ table.Values) Edge {
			var t Edge
			t.rid = id
			t.name = name
			t.eng = eng
			return t
		},
		components.Edge{},
	)
}

func newmap[T Entity](
	eng query.Engine,
	q func(query.Engine, query.Query),
	fact func(table.RowId, components.Name, table.Values) T,
	def ...table.Value,
) (*genericmap[T], error) {
	var g genericmap[T]
	g.cache = make(map[internal.NormalizedStr]T, 256)
	g.def = def
	g.fact = fact
	g.eng = eng

	if err := g.init(q); err != nil {
		return &g, err
	}
	return &g, nil
}
