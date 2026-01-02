package table

import (
	"errors"
	"slices"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"github.com/etc-sudonters/substrate/slipup"
)

func querycore(tbl *Table, qs []Q) (bitset32.Bitset, Columns, error) {
	qb := qb{tbl: tbl}
	fill := bitset32.Bitset{}

	for _, q := range qs {
		if err := q(&qb); err != nil {
			return fill, nil, slipup.Describe(err, "failed to build query")
		}
	}

	admit := qb.exists.Union(qb.load)
	deny := qb.notExists

	for row, membership := range tbl.Rows {
		if !admit.Intersect(*membership).Eq(admit) || // every value in admit must be set
			!membership.Difference(deny).Eq(*membership) { // every value in deny must not be set
			continue
		}
		fill.Set(uint32(row))
	}

	var columns Columns
	for _, cid := range qb.returning {
		columns = append(columns, tbl.Cols[cid])
	}

	return fill, columns, nil
}

func QueryRowIds(tbl *Table, q Q, qs ...Q) (RowIds, error) {
	qs = append(qs, q)
	fill, _, err := querycore(tbl, qs)
	return RowIds{fill}, err
}

func Query(tbl *Table, q Q, qs ...Q) (ResultSetIter, error) {
	qs = slices.Concat([]Q{q}, qs)
	fill, columns, err := querycore(tbl, qs)
	if err != nil {
		return nil, err
	}
	return IterResultSet(fill, columns)
}

type ComparableValue interface {
	comparable
	Value
}

var ErrNoRowsMatch = errors.New("no matching rows")

func FindOne[K ComparableValue](tbl *Table, key K, qs ...Q) (RowId, error) {
	rows, err := Query(tbl, Load[K], qs...)
	if err != nil {
		return INVALID_ROWID, err
	}

	for row, tup := range rows.All {
		candidate := tup.Values[0].(K)
		if key == candidate {
			return row, nil
		}
	}

	return INVALID_ROWID, ErrNoRowsMatch
}

func GetValues(tbl *Table, rowId RowId, colId ColumnIds) (ValueTuple, error) {
	var vt ValueTuple
	vt.Cols = make(ColumnMetas, len(colId))
	vt.Values = make(Values, len(colId))
	for i, cid := range colId {
		c := tbl.Cols[cid]
		vt.Cols[i].Id = c.Id()
		vt.Cols[i].T = c.Type()
		vt.Values[i] = c.Column().Get(rowId)
	}

	return vt, nil
}

type Q func(*qb) error

type qb struct {
	tbl       *Table
	returning ColumnIds

	load, exists, notExists, optional bitset32.Bitset
}

func setColIn[T Value](tbl *Table, cols *bitset32.Bitset) error {
	cid, err := ColumnIdFor[T](tbl)
	if err != nil {
		return err
	}
	bitset32.Set(cols, cid)
	return nil

}

func Optional[T Value](qb *qb) error {
	cid, err := ColumnIdFor[T](qb.tbl)
	if err != nil {
		return err
	}
	if bitset32.Set(&qb.optional, cid) && !bitset32.IsSet(&qb.load, cid) {
		qb.returning = append(qb.returning, cid)
	}
	return nil
}

func Load[T Value](qb *qb) error {
	cid, err := ColumnIdFor[T](qb.tbl)
	if err != nil {
		return err
	}
	if bitset32.Set(&qb.load, cid) && !bitset32.IsSet(&qb.optional, cid) {
		qb.returning = append(qb.returning, cid)
	}
	return nil
}

func Exists[T Value](qb *qb) error {
	return setColIn[T](qb.tbl, &qb.exists)
}

func NotExists[T Value](qb *qb) error {
	return setColIn[T](qb.tbl, &qb.notExists)
}
