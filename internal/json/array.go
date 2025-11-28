package json

import (
	"fmt"
	"iter"
)

type ArrayParser struct {
	p *Parser
}

func (this *ArrayParser) ReadInt() (int, error) {
	number, err := this.p.ReadInt()
	if err != nil {
		return 0, err
	}
	maybeReadComma(this.p)
	return number, nil
}

func (this *ArrayParser) ReadFloat() (float64, error) {
	number, err := this.p.ReadFloat()
	if err != nil {
		return 0, err
	}
	maybeReadComma(this.p)
	return number, nil
}

func (this *ArrayParser) ReadString() (string, error) {
	str, err := this.p.ReadString()
	if err != nil {
		return "", err
	}
	maybeReadComma(this.p)
	return str, nil
}

func (this *ArrayParser) ReadBool() (bool, error) {
	boolean, err := this.p.ReadBool()
	if err != nil {
		return false, err
	}
	maybeReadComma(this.p)
	return boolean, nil
}

func (this *ArrayParser) ReadEnd() error {
	_, err := this.p.expect(ARR_CLOSE)
	if err != nil {
		return err
	}
	this.p.Next()
	maybeReadComma(this.p)
	return nil
}

func (this *ArrayParser) DiscardValue() error {
	if err := this.p.Discard(); err != nil {
		return fmt.Errorf("unexpected end of array: %w", err)
	}
	maybeReadComma(this.p)
	return nil
}

func (this *ArrayParser) More() bool {
	return this.p.curr.Kind != ARR_CLOSE
}

func (this *ArrayParser) ReadArray() (*ArrayParser, error) {
	return this.p.ReadArray()
}

func (this *ArrayParser) ReadObject() (*ObjectParser, error) {
	return this.p.ReadObject()
}

func (this *ArrayParser) DiscardRemaining() error {
	var i int
	for this.More() {
		if err := this.DiscardValue(); err != nil {
			return err
		}
		i++
	}
	this.ReadEnd()
	return nil
}

func (this *ArrayParser) Current() Token {
	return this.p.curr
}

func (this *ArrayParser) Peek() Token {
	return this.p.peek
}

func (this *ArrayParser) Strings(err *error) iter.Seq2[int, string] {
	return ReadArrayValues(this, (*ArrayParser).ReadString, err)
}

func (this *ArrayParser) Numbers(err *error) iter.Seq2[int, float64] {
	return ReadArrayValues(this, (*ArrayParser).ReadFloat, err)
}

func (this *ArrayParser) Ints(err *error) iter.Seq2[int, int] {
	return ReadArrayValues(this, (*ArrayParser).ReadInt, err)
}

func ReadArrayValues[T any](arr *ArrayParser, read func(*ArrayParser) (T, error), err *error) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		var i int
		for arr.More() {
			val, readErr := read(arr)
			if readErr != nil {
				*err = readErr
				return
			}
			if !yield(i, val) {
				return
			}
		}
	}
}

func IntoSlice[T any](arr *ArrayParser, read func(*ArrayParser) (T, error)) ([]T, error) {
	var err error
	var items []T
	for _, item := range ReadArrayValues(arr, read, &err) {
		items = append(items, item)
	}
	return items, err
}
