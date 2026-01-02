package table

import (
	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"github.com/etc-sudonters/substrate/slipup"
)

func Query(tbl *Table, q Q, qs ...Q) (ResultSetIter, error) {
	qb := qb{parent: tbl}

	if err := q(&qb); err != nil {
		return nil, slipup.Describe(err, "failed to build query")
	}

	for _, q := range qs {
		if err := q(&qb); err != nil {
			return nil, slipup.Describe(err, "failed to build query")
		}
	}

	admit := qb.exists.Union(qb.load)
	deny := qb.notExists
	fill := bitset32.Bitset{}

	for row, membership := range tbl.Rows {
		if !admit.Intersect(*membership).Eq(admit) || // every value in admit must be set
			!membership.Difference(deny).Eq(*membership) { // every value in deny must not be set
			continue
		}
		fill.Set(uint32(row))
	}

	requestedColumns := qb.load.Union(qb.optional)
	var columns Columns
	for cid := range bitset32.IterT[ColumnId](&requestedColumns).All {
		columns = append(columns, tbl.Cols[cid])
	}

	return IterResultSet(fill, columns)
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
	parent *Table

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
	return setColIn[T](qb.parent, &qb.optional)
}

func Load[T Value](qb *qb) error {
	return setColIn[T](qb.parent, &qb.load)
}

func Exists[T Value](qb *qb) error {
	return setColIn[T](qb.parent, &qb.exists)
}

func NotExists[T Value](qb *qb) error {
	return setColIn[T](qb.parent, &qb.notExists)
}
