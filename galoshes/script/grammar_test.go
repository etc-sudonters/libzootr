package script

import (
	"testing"

	"github.com/etc-sudonters/substrate/peruse"
)

func TestCanParseInsert(t *testing.T) {
	insertScript := `insert [
        [ $place-id world/placement/holds $triforce ]
    ] where [
        [ :grotto-scrub $place-id ]
        [ $triforce names "Triforce Piece" ]
    ] rules [
        [:grotto-scrub $place] [ [ $place world/node/grotto 1 ] [ $place world/node/scrub 1 ] ]
    ]
`
	expectedInserting := 1
	expectedClauses := 2
	expectedRules := 1

	lexer := NewLexer(insertScript)
	parser := peruse.NewParser(grammar, lexer)

	result, parseErr := parser.ParseAt(peruse.LOWEST)
	t.Logf("lexer state: %#v", lexer)
	t.Logf("parser state: %#v", parser)
	if parseErr != nil {
		t.Fatal(parseErr.Error())
	}

	insert, isInsert := result.(*InsertNode)

	if !isInsert {
		t.Log("failed to parse insert")
		t.FailNow()
	}

	t.Logf("%#v", insert)

	if len(insert.Inserting) != expectedInserting {
		t.Fatalf("expected to parse %d variables, parsed %d", expectedInserting, len(insert.Inserting))
	}

	if len(insert.Clauses) != expectedClauses {
		t.Logf("expected to parse %d clauses, parsed %d", expectedClauses, len(insert.Clauses))
		t.Fail()
	}

	if len(insert.Rules) != expectedRules {
		t.Logf("expected to parse %d rules, parsed %d", expectedRules, len(insert.Rules))
		t.Fail()
	}

	if parser.HasMore() {
		t.Fatalf("parser did not consume entire input")
	}
}

func TestCanParseFind(t *testing.T) {
	findScript := `find [ $place-id $place-name ]
where [
    [ :holding $place-id $token-id ]
    [ :named $place-id $place-name ]
    [ :named $token-id "Kokri Sword" ]
    [ $token-id tokens 1 ] 
]
rules [
    [:holding $place $token] [ [$place world/placement/holds $token] ]
    [:named $id $name] [ [$id names $name] ]
]
`
	expectedFinding := 2
	expectedClauses := 4
	expectedRules := 2

	lexer := NewLexer(findScript)
	parser := peruse.NewParser(grammar, lexer)

	result, parseErr := parser.ParseAt(peruse.LOWEST)
	t.Logf("lexer state: %#v", lexer)
	t.Logf("parser state: %#v", parser)
	if parseErr != nil {
		t.Fatal(parseErr.Error())
	}

	find, isFind := result.(*FindNode)

	if !isFind {
		t.Log("failed to parse find")
		t.FailNow()
	}

	t.Logf("%#v", find)

	if len(find.Finding) != expectedFinding {
		t.Fatalf("expected to parse %d variables, parsed %d", expectedFinding, len(find.Finding))
	}

	if len(find.Clauses) != expectedClauses {
		t.Logf("expected to parse %d clauses, parsed %d", expectedClauses, len(find.Clauses))
		t.Fail()
	}

	if len(find.Rules) != expectedRules {
		t.Logf("expected to parse %d rules, parsed %d", expectedRules, len(find.Rules))
		t.Fail()
	}

	if parser.HasMore() {
		t.Fatalf("parser did not consume entire input")
	}
}
