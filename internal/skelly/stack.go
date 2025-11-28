package skelly

import (
	"errors"
)

var (
	ErrStackEmpty = errors.New("stack empty")
	ErrStackFull  = errors.New("stack full")
)

func NewStack[T any](size int) Stack[T] {
	return Stack[T]{
		items: make([]T, size),
	}
}

type Stack[T any] struct {
	items []T
	ptr   int
}

func (this *Stack[T]) Reset() {
	this.ptr = 0
}

func (this *Stack[T]) Slice(start, count int) []T {
	return this.items[start : start+count]
}

func (this *Stack[T]) PopN(n int) {
	this.ptr -= n
}

func (this *Stack[T]) Top() T {
	if this.ptr < 1 {
		panic(ErrStackEmpty)
	}

	return this.items[this.ptr-1]
}

func (this *Stack[T]) Push(item T) {
	if this.ptr == len(this.items) {
		panic(ErrStackFull)
	}

	this.items[this.ptr] = item
	this.ptr++
}

func (this *Stack[T]) Pop() T {
	if this.ptr == 0 {
		panic(ErrStackEmpty)
	}
	this.ptr--
	return this.items[this.ptr]
}
