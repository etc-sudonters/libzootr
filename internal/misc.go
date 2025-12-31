package internal

import (
	"io/fs"
	"reflect"
	"regexp"
	"strings"

	"github.com/etc-sudonters/substrate/slipup"
)

var idcharsonly = regexp.MustCompile("[^a-z0-9]+")

type NormalizedStr string

func Normalize[S ~string](s S) NormalizedStr {
	return NormalizedStr(idcharsonly.ReplaceAllString(strings.ToLower(string(s)), ""))
}

func IsFile(e fs.DirEntry) bool {
	return e.Type()&fs.ModeType == 0
}

func T[E any]() reflect.Type {
	return reflect.TypeFor[E]()
}

func TypeAssert[Ty any](a any) (t Ty, err error) {
	t, cast := a.(Ty)
	if !cast {
		err = slipup.Createf("failed to cast %v to %s", a, T[Ty]().Name())
	}
	err = nil
	return
}
