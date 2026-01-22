package ruleparser

import (
	"testing"

	"github.com/etc-sudonters/substrate/peruse"
)

func TestCanLexIdentifier(t *testing.T) {
	rule := "Compiler"
	expected := []peruse.Token{{Type: TokenIdentifier, Literal: rule}}

	l := NewRulesLexer(rule)
	collected := lexUntilEofOrErr(l, t)

	toksAreEqual(expected, collected, t)
}

func TestErrsIfNonSepFollowsIdent(t *testing.T) {
	expected := []peruse.Token{
		{Type: peruse.ERR, Pos: 8, Literal: "unexpected '\\''"},
	}

	l := NewRulesLexer("Compiler'")
	collected := lexUntilEofOrErr(l, t)

	toksAreEqual(expected, collected, t)
}

func TestCanLexBoolOp(t *testing.T) {
	rule := "Compiler or Interpreter"
	l := NewRulesLexer(rule)
	expected := []peruse.Token{
		{Type: TokenIdentifier, Literal: "Compiler", Pos: 0},
		{Type: TokenOr, Literal: "or", Pos: 9},
		{Type: TokenIdentifier, Literal: "Interpreter", Pos: 12},
	}

	collected := lexUntilEofOrErr(l, t)

	toksAreEqual(expected, collected, t)
}

func TestCanLexNumber(t *testing.T) {
	l := NewRulesLexer("99")
	expected := []peruse.Token{{Type: TokenNumber, Pos: 0, Literal: "99"}}

	collected := lexUntilEofOrErr(l, t)

	toksAreEqual(expected, collected, t)
}

func TestErrsIfNonSepFollowsNumber(t *testing.T) {
	expected := []peruse.Token{
		{Type: peruse.ERR, Pos: 2, Literal: "unexpected '\\''"},
	}

	l := NewRulesLexer("99'")

	collected := lexUntilEofOrErr(l, t)

	toksAreEqual(expected, collected, t)
}

func TestCanLexParens(t *testing.T) {
	rule := "(Compiler or Interpreter)"
	expected := []peruse.Token{
		{Type: TokenOpenParen, Pos: 0, Literal: "("},
		{Type: TokenIdentifier, Pos: 1, Literal: "Compiler"},
		{Type: TokenOr, Pos: 10, Literal: "or"},
		{Type: TokenIdentifier, Pos: 13, Literal: "Interpreter"},
		{Type: TokenCloseParen, Pos: 24, Literal: ")"},
	}

	l := NewRulesLexer(rule)
	collected := lexUntilEofOrErr(l, t)

	toksAreEqual(expected, collected, t)
}

func TestErrsOnUnclosedParen(t *testing.T) {
	rule := "(Compiler"
	expected := []peruse.Token{
		{Type: TokenOpenParen, Pos: 0, Literal: "("},
		{Type: TokenIdentifier, Pos: 1, Literal: "Compiler"},
		{Type: peruse.ERR, Pos: 9, Literal: "unclosed '(' or '['"},
	}

	l := NewRulesLexer(rule)
	collected := lexUntilEofOrErr(l, t)

	toksAreEqual(expected, collected, t)
}

func TestErrsOnUnexpectedClosedParent(t *testing.T) {
	expected := []peruse.Token{
		{Type: TokenIdentifier, Pos: 0, Literal: "Compiler"},
		{Type: peruse.ERR, Pos: 9, Literal: "unexpected ')'"},
	}

	l := NewRulesLexer("Compiler)")
	collected := lexUntilEofOrErr(l, t)

	toksAreEqual(expected, collected, t)
}

func TestCanLexActualRules(t *testing.T) {
	rules := []string{
		"can_play(Song_of_Time) or (logic_shadow_mq_invisible_blades and damage_multiplier != 'ohko')",
		"Boomerang or (logic_spirit_fence_gs and Kokiri_Sword) or (((Silver_Rupee_Spirit_Temple_Child_Early_Torches, 5) or logic_spirit_fence_gs) and (Sticks or has_explosives or Slingshot or can_use(Dins_Fire)))",
		"((can_use(Megaton_Hammer) and logic_dc_hammer_floor) or has_explosives or king_dodongo_shortcuts) and (((Bombs or Progressive_Strength_Upgrade) and can_jumpslash) or deadly_bonks == 'ohko')",
		"is_child and at_day and (can_break_crate or chicken_count < 7)",
	}

	for _, rule := range rules {
		l := NewRulesLexer(rule)
		collected := lexUntilEofOrErr(l, t)

		if collected[len(collected)-1].Type == peruse.ERR {
			t.Logf("Failed lexing rule %.80q...", rule)
			t.Log("Failure", collected[len(collected)-1])
			t.Log("Collected tokens", collected)
			t.Logf("lexer state %+v", l)
			t.Fail()
		}
	}

}

func lexUntilEofOrErr(l *peruse.StringLexer, _ *testing.T) []peruse.Token {
	collected := []peruse.Token{}

	for {
		item := l.NextToken()
		if item.Type == peruse.EOF {
			break
		}

		collected = append(collected, item)

		if item.Type == peruse.ERR {
			break
		}
	}

	return collected
}

func toksAreEqual(expected, actual []peruse.Token, t *testing.T) {
	t.Log(t.Name())
	if len(expected) != len(actual) {
		t.Fail()
		t.Logf("expected:\t%d\nactual:\t%d\n", len(expected), len(actual))
	}

	for i := range len(actual) {
		if !fungible(actual[i], expected[i]) {
			t.Fail()
			t.Logf("mismatch at index %d\nexpected:\t%+v\nactual:\t%+v", i, expected[i], actual[i])

		}
	}
}

func fungible(i peruse.Token, o peruse.Token) bool {
	return i.Type == o.Type && i.Pos == o.Pos && i.Literal == o.Literal
}
