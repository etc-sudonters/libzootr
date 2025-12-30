package json

import "iter"

type ReadsArray interface {
	ReadArray() (*ArrayParser, error)
}

type ReadsObject interface {
	ReadObject() (*ObjectParser, error)
}

func ReadArrayOf[T any](this ReadsArray, read func(*ArrayParser) (T, error), err *error) iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		arr, startErr := this.ReadArray()
		if startErr != nil {
			*err = startErr
			return
		}

		for i, t := range ReadArrayValues(arr, read, err) {
			if !yield(i, t) {
				return
			}
		}

		if endErr := arr.ReadEnd(); endErr != nil {
			*err = endErr
		}
	}
}

func ReadObjectOf[T any](this ReadsObject, read func(*ObjectParser) (T, error), err *error) iter.Seq2[string, T] {
	return func(yield func(string, T) bool) {
		obj, startErr := this.ReadObject()
		if startErr != nil {
			*err = startErr
			return
		}

		for name, value := range ReadObjectProperties(obj, read, err) {
			if !yield(name, value) {
				return
			}
		}

		if endErr := obj.ReadEnd(); endErr != nil {
			*err = endErr
		}
	}

}

func ReadStringArray(this ReadsArray, err *error) iter.Seq2[int, string] {
	return ReadArrayOf(this, (*ArrayParser).ReadString, err)
}

func ReadFloatArray(this ReadsArray, err *error) iter.Seq2[int, float64] {
	return ReadArrayOf(this, (*ArrayParser).ReadFloat, err)
}

func ReadIntArray(this ReadsArray, err *error) iter.Seq2[int, int] {
	return ReadArrayOf(this, (*ArrayParser).ReadInt, err)
}

func ReadStringObject(this ReadsObject, err *error) iter.Seq2[string, string] {
	return ReadObjectOf(this, (*ObjectParser).ReadString, err)
}

func ReadFloatObject(this ReadsObject, err *error) iter.Seq2[string, float64] {
	return ReadObjectOf(this, (*ObjectParser).ReadFloat, err)
}

func ReadIntObject(this ReadsObject, err *error) iter.Seq2[string, int] {
	return ReadObjectOf(this, (*ObjectParser).ReadInt, err)
}

func ReadIntObjectInto(this ReadsObject, dest map[string]int) (err error) {
	for key, val := range ReadIntObject(this, &err) {
		dest[key] = val
	}
	return
}

func ParseStringWith[T any](r Reader, parse func(string) (T, error)) (T, error) {
	var t T
	str, err := r.ReadString()
	if err != nil {
		return t, err
	}
	return parse(str)
}

func ParseStringInto[T any](r Reader, into *T, parse func(string) (T, error)) error {
	val, err := ParseStringWith(r, parse)
	if err == nil {
		*into = val
	}
	return err
}

func ReadStringInto[T ~string](r Reader, into *T) error {
	val, err := r.ReadString()
	if err == nil {
		*into = T(val)
	}
	return err
}

func ReadBoolInto[T ~bool](r Reader, into *T) error {
	val, err := r.ReadBool()
	if err == nil {
		*into = T(val)
	}
	return err
}

func ReadFloatInto[T ~float64](r Reader, into *T) error {
	val, err := r.ReadFloat()
	if err == nil {
		*into = T(val)
	}
	return err
}

func ReadIntInto[T ~int](r Reader, into *T) error {
	val, err := r.ReadInt()
	if err == nil {
		*into = T(val)
	}
	return err
}

func ReadStringSlice(r Reader) (strs []string, err error) {
	err = ReadStringArrayInto(r, &strs)
	return
}

func ReadStringArrayInto(r Reader, strs *[]string) (err error) {
	for _, str := range ReadStringArray(r, &err) {
		*strs = append(*strs, str)
	}
	return
}

func ReduceStringArrayInto[T any](r Reader, seed T, reduce func(T, string) (T, error)) (T, error) {
	var err error
	for _, str := range ReadStringArray(r, &err) {
		seed, err = reduce(seed, str)
		if err != nil {
			return seed, err
		}
	}

	return seed, err
}
