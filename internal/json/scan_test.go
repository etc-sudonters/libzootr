package json

import (
	"fmt"
	"testing"
)

func (this *Scanner) DiscardAll() error {
	for this.Scan() {
	}
	return this.Err()
}

func TestScansEscapedQuote(t *testing.T) {
	if len([]byte("\"")) != 1 {
		t.Fail()
	}
}

func TestScansEmptyString(t *testing.T) {
	json := []byte{'"', '"'}
	scanner := scanner(json)
	if err := scanner.DiscardAll(); err != nil {
		t.Fatal(err)
	}
}

func TestScansNewLinesProperly(t *testing.T) {
	json := []byte("[\n\r{\n\"property\"\r:\"\nsome\r\nnew\nlines\r\"}\r\n//acomment\r\n]")
	t.Log(string(json))
	expectedLine := 11 // 1 + \n + \r + \n + \r + \n + \r\n + \n + \r + \r\n + \r\n
	scanner := scanner(json)

	if err := scanner.DiscardAll(); err != nil {
		t.Fatal(err)
	}

	if scanner.line != expectedLine {
		t.Logf("Expected %d lines but counted %d", expectedLine, scanner.line)
		t.Fail()
	}
}

func TestCanScanJson(t *testing.T) {
	tests := []severalTokens{
		makeSeveralTokens("{}", expectChar('{'), expectChar('}')),
		makeSeveralTokens("[]", expectChar('['), expectChar(']')),
		makeSeveralTokens(`["str"]`, expectChar('['), expectString("str"), expectChar(']')),
		makeSeveralTokens(
			`{"prop": "value"}`,
			expectChar('{'),
			expectString("prop"),
			expectChar(':'),
			expectString("value"),
			expectChar('}'),
		),
		makeSeveralTokens("[1,-2,3.3]",
			expectChar('['),
			expectNumber(1), expectChar(','),
			expectNumber(-2), expectChar(','),
			expectFloat("3.3"),
			expectChar(']'),
		),
		makeSeveralTokens(
			`[{"prop":"val"}, 1, null, false, "string"] # comment`,
			expectChar('['),
			expectChar('{'),
			expectString("prop"), expectChar(':'),
			expectString("val"),
			expectChar('}'), expectChar(','),
			expectNumber(1), expectChar(','),
			expectNull(), expectChar(','),
			expectFalse(), expectChar(','),
			expectString("string"),
			expectChar(']'),
			expectComment(" comment"),
		),
		makeSeveralTokens(`{
            "prop": {"nested": 1},
            "another": [[1]]
        }
        //comment
`,
			expectChar('{'),
			expectString("prop"), expectChar(':'), expectChar('{'), expectString("nested"), expectChar(':'), expectNumber(1), expectChar('}'), expectChar(','),
			expectString("another"), expectChar(':'), expectChar('['), expectChar('['), expectNumber(1), expectChar(']'), expectChar(']'),
			expectChar('}'),
			expectComment("comment"),
		),
	}

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("scan: %q", test.input), func(t *testing.T) {
			scanner := scanner([]byte(test.input))

			for scanner.Scan() {
				scanned, body := scanner.Lexeme()
				test.accept(t, scanned, string(body))
			}

			if err := scanner.Err(); err != nil {
				t.Fatalf("scanner err: %s", err)
			}

			if len(test.scanned) != test.scans {
				t.Fail()
				t.Logf("expected %d scans but only did %d", len(test.scanned), test.scans)
			}
		})
	}
}

func TestCanScanJsonAtom(t *testing.T) {

	tests := []singleToken{

		{"{", scanned_obj_open, "{"},
		{"}", scanned_obj_close, "}"},
		{"[", scanned_arr_open, "["},
		{"]", scanned_arr_close, "]"},
		{",", scanned_comma, ","},
		{":", scanned_colon, ":"},
		{`"a string"`, scanned_string, "a string"},
		{`"a string\""`, scanned_string, `a string"`},
		{"\"a string\\\"\"", scanned_string, `a string"`},
		{"true", scanned_true, "true"},
		{"false", scanned_false, "false"},
		{"null", scanned_null, "null"},
		{"1", scanned_number, "1"},
		{"-1", scanned_number, "-1"},
		{".1", scanned_number, ".1"},
		{"1.1", scanned_number, "1.1"},
		{"-.1", scanned_number, "-.1"},
		{"-1.1", scanned_number, "-1.1"},
		{"//comment", scanned_comment, "comment"},
		{"//comment\n", scanned_comment, "comment"},
		{"#comment", scanned_comment, "comment"},
		{"#comment\n", scanned_comment, "comment"},
		{"\n\rfalse\r\n", scanned_false, "false"},
	}

	for _, test := range tests {
		test := test
		t.Run(fmt.Sprintf("scan %q", test.input), func(t *testing.T) {
			scanner := scanner([]byte(test.input))
			scanned := scanner.Scan()
			if !scanned {
				t.Log("failed to scan", scanner.Err())
				t.FailNow()
			}

			token, body := scanner.Lexeme()
			if !test.isExpected(token, string(body)) {
				t.Logf("expected %s but found %s", test.scanned, token)
				t.Logf("expected %q but found %q", test.body, string(body))
				t.Fail()
			}

			secondScan := scanner.Scan()
			if secondScan {
				token, body := scanner.Lexeme()
				t.Fatalf("expected only 1 scan, second scan found %s: %q", token, string(body))
			}
		})
	}
}
