package table

type iter struct {
	tbl *Table
}

func Iterate(tbl *Table) iter {
	return iter{tbl}
}

func (this iter) UnsafeRows(yield func(RowId, *Row) bool) {
	for idx := range this.tbl.rows {
		if !yield(RowId(idx), &this.tbl.rows[idx]) {
			return
		}
	}
}

func (this iter) Rows(yield func(RowId, Row) bool) {
	for idx, row := range this.tbl.rows {
		if !yield(RowId(idx), copyRow(row)) {
			return
		}
	}
}
