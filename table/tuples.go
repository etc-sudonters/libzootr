package table

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/etc-sudonters/substrate/mirrors"
)

type RowTuple struct {
	Id RowId
	ValueTuple
}

type ValueTuple struct {
	Cols   ColumnMetas
	Values Values
}

func (vt *ValueTuple) Init(cs Columns) {
	vt.Cols = make(ColumnMetas, len(cs))
	vt.Values = make(Values, len(cs))
	for i, c := range cs {
		vt.Cols[i].Id = c.Id()
		vt.Cols[i].T = c.Type()
	}
}

func (vt *ValueTuple) Load(r RowId, cs Columns) {
	for i, c := range cs {
		vt.Values[i] = c.Column().Get(r)
	}
}

func (v *ValueTuple) ColumnMap() ColumnMap {
	m := make(ColumnMap, len(v.Values))

	for i := range v.Cols {
		m[v.Cols[i].T] = v.Values[i]
	}

	return m
}

var ErrColumnNotPresent = errors.New("column not present")
var ErrCouldNotCastColumn = errors.New("could not cast column")

func Extract[T any](cm ColumnMap) (*T, error) {
	typ := reflect.TypeFor[T]()
	item, exists := cm[typ]
	if !exists {
		return nil, fmt.Errorf("%w: '%s'", ErrColumnNotPresent, typ.Name())
	}
	t, casted := item.(T)
	if !casted {
		return nil, fmt.Errorf("%w: '%s'", ErrCouldNotCastColumn, typ.Name())
	}
	return &t, nil
}

func FromColumnMap[T any](cm ColumnMap) (T, bool) {
	t, exists := cm[reflect.TypeFor[T]()]
	if !exists {
		return mirrors.Empty[T](), false
	}
	return t.(T), true
}
