package tracking

import (
	"errors"
	"sudonters/libzootr/table/ocm"

	"github.com/etc-sudonters/substrate/slipup"
)

type Set struct {
	Nodes  Nodes
	Tokens Tokens
}

func NewTrackingSet(entities *ocm.Entities) (Set, error) {
	nodes, nodeErr := NewNodes(entities)
	if nodeErr != nil {
		nodeErr = slipup.Describe(nodeErr, "failed to retrieve nodes")
	}
	tokens, tokenErr := NewTokens(entities)
	if tokenErr != nil {
		tokenErr = slipup.Describe(tokenErr, "failed to retrieve tokens")
	}

	return Set{nodes, tokens}, errors.Join(nodeErr, tokenErr)
}
