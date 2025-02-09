package bitset32

import (
	"iter"
	"math/bits"
)

type IterOf[T ~uint32] interface {
	All(func(T) bool)
	Buckets(func(uint32) bool)
	UntilEmpty(func(T) bool)
}

func Iter(b *Bitset) iter32 {
	return iter32{b}
}

func IterT[T ~uint32](b *Bitset) iter32T[T] {
	return iter32T[T]{Iter(b)}
}

type iter32 struct {
	set *Bitset
}

func (i iter32) All(yield func(v uint32) bool) {
	all(i.set)(yield)
}

func (i iter32) Buckets(yield func(v uint32) bool) {
	parts := ToRawParts(*i.set)
	for _, bucket := range parts {
		if !yield(bucket) {
			break
		}
	}
}

func (i iter32) UntilEmpty(yield func(uint32) bool) {
	for !i.set.IsEmpty() {
		if !yield(i.set.Pop()) {
			return
		}
	}
}

type iter32T[T ~uint32] struct {
	iter32
}

func (i iter32T[T]) UntilEmpty(yield func(T) bool) {
	for x := range i.iter32.UntilEmpty {
		if !yield(T(x)) {
			return
		}
	}
}

func (i iter32T[T]) All(yield func(v T) bool) {
	for x := range i.iter32.All {
		if !yield(T(x)) {
			break
		}
	}
}

func all(set *Bitset) iter.Seq[uint32] {
	return func(yield func(v uint32) bool) {
		parts := ToRawParts(*set)
	iter:
		for p, part := range parts {
			for part != 0 {
				tz := bits.TrailingZeros32(part)
				if !yield(uint32(tz + (p * 32))) {
					break iter
				}
				part ^= (1 << tz)
			}
		}
	}
}
