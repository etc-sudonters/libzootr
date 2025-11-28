package query

import (
	"sudonters/libzootr/galoshes/script"
	"sudonters/libzootr/internal/table"
)

/*
   Goals:
   1. find [ $id ] [[ $id /name "Kokri Sword" ]]
   2. find [ $id $name ] where [[ $id /name $name ]]
   3. find [ $id $name ] where [
       [ $id /name name ]
       [ $id /world/placement true ]
   ]
   4. find [ $id $name ] where [
       [ :named $id $name ]
       [ :placement $id ]
   ] rules [
       [ [:named $id $name] [[ $id /name $name ]] ]
       [ [:placement $id] [[ $id /world/placement true ]]
   ]
   5. find [ $place-id $place-name ] where [
       [ :named $place-id $place-name ]
       [ :holding $place-id $token-id ]
       [ :named $token-id "Kokri Sword" ]
   ]
*/

type Expr interface {
	Type() table.ColumnKind
}

type Variable struct {
	Id uint32
	T  table.ColumnKind
}

type Number float64
type String string
type Bool bool
type Attr struct {
	Id table.ColumnId
	T  table.ColumnKind
}

type FindQuery struct {
	selecting []Expr
	vars      []string
}

func RunFindQuery(find FindQuery, tbl *table.Table, context map[string]table.Value) ([][]table.Value, error) {
	results := make([][]table.Value, 0, 16)
	if context == nil {
		context = make(map[string]table.Value)
	}

	return results, nil
}

func BuildFind(find *script.FindNode) FindQuery {
	panic("not implemented")
}
