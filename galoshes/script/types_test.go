package script

import (
	"strings"
	"testing"

	"github.com/etc-sudonters/substrate/peruse"
)

func TestCanAnnotateScript(t *testing.T) {

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

	t.Log(findScript)
	lexer := NewLexer(findScript)
	parser := peruse.NewParser(grammar, lexer)

	result, parseErr := parser.ParseAt(peruse.LOWEST)
	if parseErr != nil {
		t.Logf("lexer state: %#v", lexer)
		t.Logf("parser state: %#v", parser)
		t.Fatal(parseErr.Error())
	}

	find, isFind := result.(*FindNode)

	if !isFind {
		t.Logf("%#v", result)
		t.Fatal("failed to parse find")
	}

	subs := make(Substitutions, 16)
	env := NewAstEnv()
	env.AddAttr("world/placement/holds", 1, TypeNumber{})
	env.AddAttr("tokens", 2, TypeNumber{})
	env.AddAttr("names", 3, TypeString{})

	if err := Annotate(find, env, subs); err != nil {
		t.Fatalf("failed to annotate: %s", err)
	}
	ts := TypeDisplay{Sink: &strings.Builder{}}
	ts.VisitFindNode(find)
	t.Log(ts.Sink.String())

	if find.Type == nil {
		t.Fatal("expected to type entire find script")
	}

	expected := TypeTuple{[]Type{TypeNumber{}, TypeString{}}}
	originalFind := find.Type
	if err := SubstituteTypes(find, subs); err != nil {
		t.Fatalf("failed to substitute types: %s", err)
	}
	ts.Sink.Reset()
	ts.VisitFindNode(find)
	t.Log(ts.Sink.String())
	actual := find.Type
	t.Logf("expected to type find as %#v", expected)
	t.Logf("initially typed find as %#v", originalFind)
	t.Logf("applied type: %#v", actual)
	t.Logf("env: %#v", find.Env.names)
	if !expected.StrictlyEq(actual) {
		t.Fail()
	}
}

func TestApplySubsitutions(t *testing.T) {
	subs := Substitutions{
		TypeVar(1): TypeVar(3),
		TypeVar(3): TypeVar(2),
		TypeVar(2): TypeNumber{},
		TypeVar(4): TypeTuple{[]Type{TypeVar(5), TypeVar(6)}},
		TypeVar(5): TypeVar(1),
		TypeVar(6): TypeString{},
	}

	shouldBeNum, _ := Substitute(TypeVar(1), subs)
	if _, isNum := shouldBeNum.(TypeNumber); !isNum {
		t.Logf("expected to generate number binding: %#v", shouldBeNum)
		t.FailNow()
	}

	shouldBeTT, _ := Substitute(TypeVar(4), subs)
	tt, isTT := shouldBeTT.(TypeTuple)
	if !isTT {
		t.Logf("expected to generate type tuple binding: %#v", shouldBeTT)
		t.FailNow()
	}
	t.Logf("tt: %#v", tt)
	t.Logf("subs\n%#v", subs)

	if _, isNum := tt.Types[0].(TypeNumber); !isNum {
		t.Logf("expected to generate number binding: %#v", tt.Types[0])
		t.Fail()
	}

	if _, isStr := tt.Types[1].(TypeString); !isStr {
		t.Logf("expected to generate string binding: %#v", tt.Types[1])
		t.Fail()
	}
}
