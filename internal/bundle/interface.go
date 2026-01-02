package bundle

import (
	"fmt"
	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"sudonters/libzootr/internal/table"
)

type RowIter = func(table.RowId, table.ValueTuple) bool

type Interface interface {
	All(RowIter)
	Len() int
}

type Empty struct{}

func (e Empty) All(RowIter) {}
func (e Empty) Len() int    { return 0 }

func Many(fill bitset32.Bitset, columns table.Columns) Interface {
	return &many{fill, columns}
}

func Single(fill bitset32.Bitset, columns table.Columns) (Interface, error) {
	if fill.Len() != 1 {
		return nil, fmt.Errorf("%w: had %d", ErrExpectSingleRow, fill.Len())
	}

	tup := new(table.RowTuple)
	tup.Id = table.RowId(fill.Elems()[0])
	tup.Init(columns)
	tup.Load(tup.Id, columns)
	s := single(*tup)
	return &s, nil
}

type many struct {
	fill    bitset32.Bitset
	columns table.Columns
}

func (r *many) All(yield RowIter) {
	vt := new(table.ValueTuple)
	vt.Init(r.columns)

	biter := bitset32.Iter(&r.fill)
	for rowId := range biter.All {
		vt.Load(table.RowId(rowId), r.columns)
		if !yield(table.RowId(rowId), *vt) {
			return
		}
	}

}

func (r many) Len() int {
	return r.fill.Len()
}

type single table.RowTuple

func (s *single) All(yield RowIter) {
	yield(s.Id, s.ValueTuple)
}

func (s single) Len() int {
	return 1
}
