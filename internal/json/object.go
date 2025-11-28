package json

import (
	"fmt"
	"iter"
)

type ObjectParser struct {
	p *Parser
}

func (this *ObjectParser) ReadPropertyName() (string, error) {
	str, err := this.p.ReadString()
	if err != nil {
		return "", err
	}
	if this.p.curr.Kind != colon {
		return str, this.p.unexpected(&this.p.curr)
	}
	this.p.Next()
	return str, nil
}

func (this *ObjectParser) ReadInt() (int, error) {
	number, err := this.p.ReadInt()
	if err != nil {
		return 0, err
	}
	maybeReadComma(this.p)
	return number, nil
}

func (this *ObjectParser) ReadFloat() (float64, error) {
	number, err := this.p.ReadFloat()
	if err != nil {
		return 0, err
	}
	maybeReadComma(this.p)
	return number, nil
}

func (this *ObjectParser) ReadString() (string, error) {
	str, err := this.p.ReadString()
	if err != nil {
		return "", err
	}
	maybeReadComma(this.p)
	return str, nil
}

func (this *ObjectParser) ReadBool() (bool, error) {
	boolean, err := this.p.ReadBool()
	if err != nil {
		return false, err
	}
	maybeReadComma(this.p)
	return boolean, nil
}

func (this *ObjectParser) DiscardValue() error {
	if err := this.p.Discard(); err != nil {
		return fmt.Errorf("unexpected end of object: %w", err)
	}
	maybeReadComma(this.p)
	return nil
}

func (this *ObjectParser) ReadEnd() error {
	_, err := this.p.expect(OBJ_CLOSE)
	if err != nil {
		return err
	}
	this.p.Next()
	maybeReadComma(this.p)
	return nil
}

func (this *ObjectParser) More() bool {
	return this.p.curr.Kind != OBJ_CLOSE
}

func (this *ObjectParser) Current() Token {
	return this.p.curr
}

func (this *ObjectParser) Peek() Token {
	return this.p.peek
}

func (this *ObjectParser) ReadObject() (*ObjectParser, error) {
	return this.p.ReadObject()
}

func (this *ObjectParser) ReadArray() (*ArrayParser, error) {
	return this.p.ReadArray()
}

func (this *ObjectParser) DiscardRemaining() error {
	for this.More() {
		if _, err := this.ReadPropertyName(); err != nil {
			return err
		} else {
		}
		if err := this.DiscardValue(); err != nil {
			return err
		}
	}
	this.ReadEnd()
	return nil
}

func (this *ObjectParser) Strings(err *error) iter.Seq2[string, string] {
	return ReadObjectProperties(this, (*ObjectParser).ReadString, err)
}

func (this *ObjectParser) Floats(err *error) iter.Seq2[string, float64] {
	return ReadObjectProperties(this, (*ObjectParser).ReadFloat, err)
}

func (this *ObjectParser) Ints(err *error) iter.Seq2[string, int] {
	return ReadObjectProperties(this, (*ObjectParser).ReadInt, err)
}

func (this *ObjectParser) Keys(err *error) iter.Seq[string] {
	return func(yield func(string) bool) {
		for this.More() {
			prop, propertyErr := this.ReadPropertyName()
			if propertyErr != nil {
				*err = propertyErr
				return
			}
			if discardErr := this.DiscardValue(); discardErr != nil {
				*err = discardErr
				return
			}
			if !yield(prop) {
				return
			}
		}
		if endErr := this.ReadEnd(); endErr != nil {
			*err = endErr
		}
	}
}

func ReadObjectProperties[T any](obj *ObjectParser, read func(*ObjectParser) (T, error), err *error) iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		for obj.More() {
			name, propertyErr := obj.ReadPropertyName()
			if propertyErr != nil {
				*err = propertyErr
				return
			}
			val, valueErr := read(obj)
			if valueErr != nil {
				*err = valueErr
				return
			}
			if !yield(name, val) {
				return
			}
		}
	}
}
