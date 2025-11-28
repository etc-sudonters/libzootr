package script

import (
	"slices"
	"testing"

	"github.com/etc-sudonters/substrate/peruse"
)

func TestLexFullScript(t *testing.T) {
	script := `; where is the kokri sword placed?
find [ $place-id $place-name ]
where [
    [ $place-id world/placement/holds $token-id ]
    [ :named $place-id $place-name]
    [ :named $token-id "Kokri Sword"]
]
rules [
    [:named $id $name] [$id names $name]
]`

	lexer := NewLexer(script)
	collected := slices.Collect(Tokens(lexer))

	if last := collected[len(collected)-1]; last.Is(ERR) {
		t.Fatalf("failed to parse: POS %d %q", last.Pos, last.Literal)
	}

	expected := expect().
		Find().OpenBracket().Variable("place-id").Variable("place-name").CloseBracket().
		Where().OpenBracket().
		OpenBracket().Variable("place-id").Attribute("world/placement/holds").Variable("token-id").CloseBracket().
		OpenBracket().Invoke("named").Variable("place-id").Variable("place-name").CloseBracket().
		OpenBracket().Invoke("named").Variable("token-id").String("Kokri Sword").CloseBracket().
		CloseBracket().
		Rules().OpenBracket().
		OpenBracket().Invoke("named").Variable("id").Variable("name").CloseBracket().
		OpenBracket().Variable("id").Attribute("names").Variable("name").CloseBracket().
		CloseBracket()

	expected.AssertEqual(t, collected)
}

func TestCanLexManyScripts(t *testing.T) {
	scripts := []string{
		`[ [:named $id $name] [$id names $name] ]`,
		`[ [:in-region $node-id $region-id ] [ $node-id world/region $region-id ] ]`,
		`[ [:in-dungeon $node-id $dungeon-id ] [
    [:in-region $node-id $dungeon-id]
    [$dungeon-id world/region/kind "Dungeon"]
]]`,
		` insert [ 100 test/scores 100 ]`,
	}

	for _, script := range scripts {
		script := script
		t.Run(script, func(t *testing.T) {

			lexer := NewLexer(script)
			collected := slices.Collect(Tokens(lexer))

			if last := collected[len(collected)-1]; last.Is(ERR) {
				t.Fatalf("failed to parse: POS %d %q", last.Pos, last.Literal)
			}
		})
	}
}

func expect() *ExpectedTokens { return new(ExpectedTokens) }

type ExpectedTokens struct {
	expected []peruse.Token
}

func (this *ExpectedTokens) Len() int {
	return len(this.expected)
}

func (this *ExpectedTokens) AssertEqual(t *testing.T, collected []peruse.Token) {
	t.Helper()

	if len(this.expected) != len(collected) {
		t.Logf("expected to lex %d tokens but lexed %d", len(this.expected), len(collected))
		t.Fail()
	}

	for i := range min(len(this.expected), len(collected)) {
		collected := collected[i]
		expected := this.expected[i]

		if collected.Type != expected.Type || collected.Literal != expected.Literal {
			t.Logf("expected to lex %#v but lexed %#v", expected, collected)
			t.Fail()
		}
	}

}

func (this *ExpectedTokens) add(t TokenType, lit string) *ExpectedTokens {
	this.expected = append(this.expected, peruse.Token{Type: t, Literal: lit})
	return this
}

func (this *ExpectedTokens) Find() *ExpectedTokens         { return this.add(TOKEN_FIND, findWord) }
func (this *ExpectedTokens) With() *ExpectedTokens         { return this.add(TOKEN_WITH, withWord) }
func (this *ExpectedTokens) Where() *ExpectedTokens        { return this.add(TOKEN_WHERE, whereWord) }
func (this *ExpectedTokens) Insert() *ExpectedTokens       { return this.add(TOKEN_INSERT, insertWord) }
func (this *ExpectedTokens) Rules() *ExpectedTokens        { return this.add(TOKEN_RULES, rulesWord) }
func (this *ExpectedTokens) True() *ExpectedTokens         { return this.add(TOKEN_TRUE, trueWord) }
func (this *ExpectedTokens) False() *ExpectedTokens        { return this.add(TOKEN_FALSE, falseWord) }
func (this *ExpectedTokens) Comma() *ExpectedTokens        { return this.add(TOKEN_COMMA, ",") }
func (this *ExpectedTokens) OpenBracket() *ExpectedTokens  { return this.add(TOKEN_OPEN_BRACKET, "[") }
func (this *ExpectedTokens) CloseBracket() *ExpectedTokens { return this.add(TOKEN_CLOSE_BRACKET, "]") }
func (this *ExpectedTokens) Assign() *ExpectedTokens       { return this.add(TOKEN_ASSIGN, ":-") }

func (this *ExpectedTokens) Invoke(name string) *ExpectedTokens {
	return this.add(TOKEN_RULE, name)
}
func (this *ExpectedTokens) Variable(name string) *ExpectedTokens {
	return this.add(TOKEN_VARIABLE, name)
}
func (this *ExpectedTokens) Attribute(name string) *ExpectedTokens {
	return this.add(TOKEN_ATTRIBUTE, name)
}
func (this *ExpectedTokens) String(content string) *ExpectedTokens {
	return this.add(TOKEN_STRING, content)
}
func (this *ExpectedTokens) Number(content string) *ExpectedTokens {
	return this.add(TOKEN_NUMBER, content)
}
