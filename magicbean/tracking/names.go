package tracking

import (
	"fmt"
	"sudonters/libzootr/components"
	"sudonters/libzootr/zecs"
)

type NameTable map[zecs.Entity]components.Name

func NameTableFrom(ocm *zecs.Ocm, match zecs.BuildQuery, matches ...zecs.BuildQuery) (NameTable, error) {
	q := ocm.Query()
	q.Build(zecs.Load[components.Name])
	q.Build(match, matches...)

	rows, err := q.Execute()
	if err != nil {
		return nil, fmt.Errorf("while constructing name table: %w", err)
	}

	names := make(NameTable, rows.Len())

	for entity, tup := range rows.All {
		name := tup.Values[0].(components.Name)
		names[entity] = name
	}

	return names, nil
}
