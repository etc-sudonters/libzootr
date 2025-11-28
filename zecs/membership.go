package zecs

import (
	"github.com/etc-sudonters/substrate/skelly/bitset32"
	"sudonters/libzootr/internal/query"
)

type Membership struct {
	row    bitset32.Bitset
	parent *Ocm
}

func IsMemberOf[V Value](this Membership) bool {
	id := query.MustAsColumnId[V](this.parent.eng)
	return bitset32.IsSet(&this.row, id)
}
