package table

import (
	"errors"
	"fmt"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
)

var ErrExpectSingleRow = errors.New("expected exactly 1 row")

func IterResultSet(fill bitset32.Bitset, columns Columns) (ResultSetIter, error) {
	switch fill.Len() {
	case 0:
		return Empty{}, nil
	case 1:
		return Single(fill, columns)
	default:
		return Many(fill, columns), nil
	}
}

type YieldRow = func(RowId, ValueTuple) bool

type ResultSetIter interface {
	All(YieldRow)
	Len() int
}

type Empty struct{}

func (e Empty) All(YieldRow) {}
func (e Empty) Len() int     { return 0 }

func Many(fill bitset32.Bitset, columns Columns) ResultSetIter {
	return &many{fill, columns}
}

func Single(fill bitset32.Bitset, columns Columns) (ResultSetIter, error) {
	if fill.Len() != 1 {
		return nil, fmt.Errorf("%w: had %d", ErrExpectSingleRow, fill.Len())
	}

	tup := new(RowTuple)
	tup.Id = RowId(fill.Elems()[0])
	tup.Init(columns)
	tup.Load(tup.Id, columns)
	s := single(*tup)
	return &s, nil
}

type many struct {
	fill    bitset32.Bitset
	columns Columns
}

func (r *many) All(yield YieldRow) {
	vt := new(ValueTuple)
	vt.Init(r.columns)

	biter := bitset32.Iter(&r.fill)
	for rowId := range biter.All {
		vt.Load(RowId(rowId), r.columns)
		if !yield(RowId(rowId), *vt) {
			return
		}
	}

}

func (r many) Len() int {
	return r.fill.Len()
}

type single RowTuple

func (s *single) All(yield YieldRow) {
	yield(s.Id, s.ValueTuple)
}

func (s single) Len() int {
	return 1
}
