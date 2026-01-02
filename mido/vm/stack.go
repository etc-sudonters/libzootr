package vm

import (
	"errors"
)

var (
	ErrStackEmpty = errors.New("stack empty")
	ErrStackFull  = errors.New("stack full")
)

func newstack[T any](size int) stack[T] {
	return stack[T]{
		items: make([]T, size),
	}
}

type stack[T any] struct {
	items []T
	ptr   int
}

func (this *stack[T]) slice(start, count int) []T {
	return this.items[start : start+count]
}

func (this *stack[T]) popN(n int) {
	this.ptr -= n
}

func (this *stack[T]) push(item T) {
	if this.ptr == len(this.items) {
		panic(ErrStackFull)
	}

	this.items[this.ptr] = item
	this.ptr++
}

func (this *stack[T]) pop() T {
	if this.ptr == 0 {
		panic(ErrStackEmpty)
	}
	this.ptr--
	return this.items[this.ptr]
}
