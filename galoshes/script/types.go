package script

import (
	"errors"
	"fmt"
	"strings"
)

var ErrTypeReoccurs = errors.New("type reoccurs in itself")

type Type interface {
	StrictlyEq(Type) bool
	String() string
}

type TypeVar uint64

func (this TypeVar) StrictlyEq(other Type) bool {
	tv, isTv := other.(TypeVar)
	return isTv && this == tv
}

func (this TypeVar) String() string {
	return fmt.Sprintf("TypeVar{%d}", uint64(this))
}

type TypeVoid struct{}

func (this TypeVoid) String() string {
	return "Void"
}

func (this TypeVoid) StrictlyEq(other Type) bool {
	_, isTypeVoid := other.(TypeVoid)
	return isTypeVoid
}

type TypeNumber struct{}

func (this TypeNumber) String() string {
	return "Number"
}

func (this TypeNumber) StrictlyEq(other Type) bool {
	_, isTypeNumber := other.(TypeNumber)
	return isTypeNumber
}

type TypeString struct{}

func (this TypeString) String() string {
	return "String"
}

func (this TypeString) StrictlyEq(other Type) bool {
	_, isTypeString := other.(TypeString)
	return isTypeString
}

type TypeBool struct{}

func (this TypeBool) String() string {
	return "Bool"
}

func (this TypeBool) StrictlyEq(other Type) bool {
	_, isTypeBool := other.(TypeBool)
	return isTypeBool
}

type TypeTuple struct {
	Types []Type
}

func (this TypeTuple) String() string {
	str := &strings.Builder{}
	str.WriteString("Tuple{")
	for i := range this.Types {
		fmt.Fprintf(str, "%s,", this.Types[i])
	}
	str.WriteRune('}')
	return str.String()
}

func (this TypeTuple) StrictlyEq(other Type) bool {
	tt, isTT := other.(TypeTuple)
	if !isTT || len(tt.Types) != len(this.Types) {
		return false
	}
	return allTypesEq(this.Types, tt.Types)
}

func allTypesEq(t1, t2 []Type) bool {
	for i := range t1 {
		if !t1[i].StrictlyEq(t2[i]) {
			return false
		}
	}

	return true
}
