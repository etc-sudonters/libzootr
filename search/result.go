package search

import (
	"errors"
	"sudonters/libzootr/magicbean"

	"github.com/etc-sudonters/substrate/skelly/bitset32"
)

var ErrNoProgress = errors.New("no progress made")

type Result struct {
	VisitedNodes bitset32.Bitset
	ReachedNodes bitset32.Bitset
	PendingNodes bitset32.Bitset
	CrossedEdges bitset32.Bitset
	VisitedEdges []magicbean.EdgeHandle
}
