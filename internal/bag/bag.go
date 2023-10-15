package bag

// Items in this package have no obvious home but are useful across a variety
// of domains

import (
	"fmt"
	"reflect"

	"golang.org/x/exp/constraints"
)

func Max[A constraints.Ordered](a, b A) A {
	if a > b {
		return a
	}
	return b
}

// returns a if a < b otherwise b
func Min[A constraints.Ordered](a, b A) A {
	if a < b {
		return a
	}
	return b
}

// determines if E is present in T
func Contains[E comparable, T ~[]E](needle E, haystack T) bool {
	for i := range haystack {
		if needle == haystack[i] {
			return true
		}
	}

	return false
}

func Map[A any, AT ~[]A, B any](as AT, f func(A) B) []B {
	bs := make([]B, len(as))

	for i, a := range as {
		bs[i] = f(a)
	}

	return bs
}

// returns the name of the type represented, if it is a pointer & is prefixed
func NiceTypeName(t reflect.Type) string {
	if t == nil {
		return "nil"
	}

	if t.Kind() != reflect.Pointer {
		return t.Name()
	}

	t = t.Elem()
	return fmt.Sprintf("&%s", t.Name())
}
