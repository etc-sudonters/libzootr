package internal

import "iter"

func EmptySeq[T any]() iter.Seq[T] {
	return func(func(T) bool) {}
}

func EmptySeq2[T1 any, T2 any]() iter.Seq2[T1, T2] {
	return func(func(T1, T2) bool) {}
}

func PanicOnError(e error) {
	if e != nil {
		panic(e)
	}
}
