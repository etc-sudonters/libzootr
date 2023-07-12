package bag

import (
	"fmt"
	"reflect"

	"golang.org/x/exp/constraints"
)

func Min[A constraints.Ordered](a, b A) A {
	if a < b {
		return a
	}
	return b
}

func Contains[E comparable, T ~[]E](needle E, haystack T) bool {
	for i := range haystack {
		if needle == haystack[i] {
			return true
		}
	}

	return false
}

func NiceTypeName(t reflect.Type) string {
	if t.Kind() != reflect.Pointer {
		return t.Name()
	}

	t = t.Elem()
	return fmt.Sprintf("&%s", t.Name())
}
