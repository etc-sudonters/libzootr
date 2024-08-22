package internal

import (
	"io/fs"
	"os"
	"reflect"
	"regexp"
	"strings"
	"github.com/etc-sudonters/substrate/slipup"
	"github.com/etc-sudonters/substrate/mirrors"
	"muzzammil.xyz/jsonc"
)

var alphanumonly = regexp.MustCompile("[^a-z0-9]+")

type NormalizedStr string

func Normalize[S ~string](s S) NormalizedStr {
	return NormalizedStr(alphanumonly.ReplaceAllString(strings.ToLower(string(s)), ""))
}

func IsFile(e fs.DirEntry) bool {
	return e.Type()&fs.ModeType == 0
}

func ReadJsonFileStringMap(path string) (map[string]string, error) {
	t := make(map[string]string)
	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		return t, readErr
	}

	err := jsonc.Unmarshal(raw, &t)
	return t, err
}

func ReadJsonFileAs[T any](path string) (T, error) {
	var t T
	raw, readErr := os.ReadFile(path)
	if readErr != nil {
		return t, readErr
	}

	err := jsonc.Unmarshal(raw, &t)
	return t, err
}

func T[E any]() reflect.Type {
	return mirrors.TypeOf[E]()
}

func TypeAssert[T any](a any) (t T, err error) {
	t, cast := a.(T)
	if !cast {
		err = slipup.Createf("failed to cast %v to %s", a, mirrors.TypeOf[T]().Name())
	}
	err = nil
	return
}
