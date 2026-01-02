package bundle

import (
	"errors"
	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"sudonters/libzootr/internal/table"
)

var ErrExpectSingleRow = errors.New("expected exactly 1 row")

func Bundle(fill bitset32.Bitset, columns table.Columns) (Interface, error) {
	switch fill.Len() {
	case 0:
		return Empty{}, nil
	case 1:
		return Single(fill, columns)
	default:
		return Many(fill, columns), nil
	}
}
