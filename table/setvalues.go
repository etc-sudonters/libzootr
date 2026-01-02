package table

import (
	"errors"
	"reflect"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"github.com/etc-sudonters/substrate/slipup"
)

func InsertRow(tbl *Table, vs ...Value) (RowId, error) {
	rid := tbl.InsertRow()
	err := SetRowValues(tbl, rid, vs)
	return rid, err
}

func SetRowValues(tbl *Table, rid RowId, vs Values) error {
	if len(vs) == 0 {
		return nil
	}

	ids := bitset32.Bitset{}
	cols := make(ColumnIds, len(vs))
	var updateErr error

	for idx, v := range vs {
		if v == nil {
			panic("insert of nil value")
		}
		vty := reflect.TypeOf(v)
		id, err := tbl.ColumnIdFor(vty)
		if err != nil {
			updateErr = errors.Join(updateErr, err)
			continue
		}
		if !bitset32.Set(&ids, id) {
			updateErr = errors.Join(updateErr, slipup.Createf("value for column %q already provided", vty.Name()))
			continue
		}
		cols[idx] = id
	}

	if updateErr != nil {
		return updateErr
	}

	for idx, cid := range cols {
		tbl.SetValue(rid, cid, vs[idx])
	}

	return nil
}

func UnsetRowValues(tbl *Table, rowId RowId, colIds ColumnIds) error {
	var unsetErr error
	for _, id := range colIds {
		if err := tbl.UnsetValue(rowId, id); err != nil {
			unsetErr = errors.Join(unsetErr, err)
		}
	}

	return unsetErr
}

type MakesColId = func(*Table) (ColumnId, error)

func UnsetRowValuesFor(tbl *Table, rowId RowId, cols ...MakesColId) error {
	if len(cols) == 0 {
		return nil
	}

	var colErr error
	colIds := make(ColumnIds, len(cols))
	for idx, f := range cols {
		cid, err := f(tbl)
		if err != nil {
			colErr = errors.Join(colErr, err)
			continue
		}
		colIds[idx] = cid
	}

	if colErr != nil {
		return colErr
	}

	return UnsetRowValues(tbl, rowId, colIds)
}
