package bundle

import (
	"errors"
	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"sudonters/libzootr/internal/table"
)

var ErrExpectSingleRow = errors.New("expected exactly 1 row")

type Bundler func(bitset32.Bitset, table.Columns) (Interface, error)

func Bundle(fill bitset32.Bitset, columns table.Columns) (Interface, error) {
	if len(columns) == 0 {
		return onlyrows(fill), nil
	}

	switch fill.Len() {
	case 0:
		return Empty{}, nil
	case 1:
		return Single(fill, columns)
	default:
		return Many(fill, columns), nil
	}
}

var BundleSingle = Single

func BundleMany(fill bitset32.Bitset, columns table.Columns) (Interface, error) {
	return Many(fill, columns), nil
}

func BundleRowsOnly(fill bitset32.Bitset, _ table.Columns) (Interface, error) {
	return onlyrows(fill), nil
}

func BundleEmpty(bitset32.Bitset, table.Column) (Interface, error) {
	return Empty{}, nil
}
