package ast

import (
	"hash/fnv"
	"sudonters/libzootr/ruleparser"
	"sudonters/libzootr/mido/symbols"
	"testing"
)

var grammar = ruleparser.NewRulesGrammar()

func TestHashNodes(t *testing.T) {
	var expectedHash uint64 = 15766529011315612546

	hasher := fnv.New64()

	var treeA, treeB Node

	treeA = Compare{
		LHS: Number(1),
		Op:  CompareEq,
		RHS: Number(2),
	}

	treeB = Compare{
		LHS: Number(1),
		Op:  CompareEq,
		RHS: Number(2),
	}

	Hash64(treeA, hasher)
	hashA := hasher.Sum64()
	hasher.Reset()

	if hashA != expectedHash {
		t.Logf("Expected hash has changed")
		t.Fail()
	}

	Hash64(treeB, hasher)
	hashB := hasher.Sum64()
	hasher.Reset()

	if hashA != hashB {
		t.Fatalf("Expected identical trees to have identical hashes:\nHash A:\t%d\nHash B:\t%d\n", hashA, hashB)
	}
}

func TestHashNodesFromSource(t *testing.T) {
	hasher := fnv.New64()
	expectedTree := Invoke{
		Target: Identifier(symbols.Index(0)),
		Args: []Node{
			Identifier(1),
			Number(20),
		}}
	Hash64(expectedTree, hasher)
	expectedHash := hasher.Sum64()
	hasher.Reset()

	syms := symbols.NewTable()
	syms.Declare("has", symbols.BUILT_IN_FUNCTION)
	syms.Declare("Bucks", symbols.TOKEN)

	parsed, err := Parse("has(Bucks, 20)", &syms, grammar)
	if err != nil {
		t.Fatal(err)
	}

	Hash64(parsed, hasher)
	parsedHash := hasher.Sum64()

	if parsedHash != expectedHash {
		t.Fatalf("Expected identical trees to have identical hashes.\nExpected:\t%d\nActual:\t%d", expectedHash, parsedHash)
	}
}
