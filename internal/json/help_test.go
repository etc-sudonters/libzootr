package json

import (
	"bytes"
	"fmt"
	"testing"
)

type singleToken struct {
	input   string
	scanned scanned
	body    string
}

func (this singleToken) isExpected(scanned scanned, body string) bool {
	return this.scanned == scanned && this.body == body
}

func scanner(src []byte) *Scanner {
	return NewScanner(bytes.NewReader(src))
}

type severalTokens struct {
	input   string
	scanned []scanned
	body    []string
	scans   int
}

func (this *severalTokens) expect(scanned scanned, body string) {
	this.scanned = append(this.scanned, scanned)
	this.body = append(this.body, body)
}

func (this *severalTokens) accept(t *testing.T, scanned scanned, body string) {
	t.Helper()

	if scanned != this.scanned[this.scans] || body != this.body[this.scans] {
		t.Fail()
		t.Logf("expected(%3d) %q but found %q", this.scans, this.scanned[this.scans], scanned)
		t.Logf("expected(%3d) %q but found %q", this.scans, this.body[this.scans], body)
	}
	this.scans++
}

func makeSeveralTokens(input string, atoms ...func(*severalTokens)) severalTokens {
	var tokens severalTokens
	tokens.input = input
	for i := range atoms {
		atoms[i](&tokens)
	}

	return tokens
}

func expectAtom(scanned scanned, body string) func(*severalTokens) {
	return func(test *severalTokens) {
		test.expect(scanned, body)
	}
}

var chars = map[byte]scanned{
	':': scanned_colon,
	',': scanned_comma,
	'{': scanned_obj_open,
	'}': scanned_obj_close,
	'[': scanned_arr_open,
	']': scanned_arr_close,
}

func expectChar(char byte) func(*severalTokens) {
	scanned := chars[char]
	return expectAtom(scanned, string(char))
}

func expectString(body string) func(*severalTokens) {
	return expectAtom(scanned_string, body)
}

func expectNumber(body int) func(*severalTokens) {
	return expectAtom(scanned_number, fmt.Sprintf("%d", body))
}

func expectFloat(body string) func(*severalTokens) {
	return expectAtom(scanned_number, body)
}

func expectFalse() func(*severalTokens) {
	return expectAtom(scanned_false, "false")
}

func expectNull() func(*severalTokens) {
	return expectAtom(scanned_null, "null")
}

func expectComment(body string) func(*severalTokens) {
	return expectAtom(scanned_comment, body)
}
