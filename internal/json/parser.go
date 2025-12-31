package json

import (
	"fmt"
	"io"
	"strconv"
	"sudonters/libzootr/internal"

	"github.com/etc-sudonters/substrate/slipup"
)

type Reader interface {
	ReadsArray
	ReadsObject
	ReadString() (string, error)
	ReadInt() (int, error)
	ReadFloat() (float64, error)
	ReadBool() (bool, error)
}

var _ Reader = (*Parser)(nil)
var _ Reader = (*ArrayParser)(nil)
var _ Reader = (*ObjectParser)(nil)

type Token struct {
	Kind Kind
	Body []byte
}

func (this Token) dump() map[string]any {
	return map[string]any{
		"kind": this.Kind.String(),
		"body": string(this.Body),
	}
}

type Kind uint8

func (this Kind) String() string {
	return scanned(this).String()
}

const (
	OBJ_OPEN  = Kind(scanned_obj_open)
	OBJ_CLOSE = Kind(scanned_obj_close)
	ARR_OPEN  = Kind(scanned_arr_open)
	ARR_CLOSE = Kind(scanned_arr_close)
	STRING    = Kind(scanned_string)
	COMMENT   = Kind(scanned_comment)
	NUMBER    = Kind(scanned_number)
	TRUE      = Kind(scanned_true)
	FALSE     = Kind(scanned_false)
	NULL      = Kind(scanned_null)
	comma     = Kind(scanned_comma)
	colon     = Kind(scanned_colon)
	EOF       = Kind(scanned_eof)
	ERR       = Kind(scanned_err)
)

type Parser struct {
	scanner    *Scanner
	curr, peek Token
}

func ParserFrom(r io.Reader) *Parser {
	return NewParser(NewScanner(r))
}

func NewParser(scanner *Scanner) *Parser {
	p := new(Parser)
	p.scanner = scanner
	p.Next()
	p.Next()
	return p
}

func (this *Parser) Dump() map[string]any {
	return map[string]any{
		"current-token": this.curr.dump(),
		"next-token":    this.peek.dump(),
		"scanner":       this.scanner.Dump(),
	}

}

func (this *Parser) makeError(cause error) error {
	return this.scanner.makePositionedError(cause)
}

func (this *Parser) Unexpected(t *Token) error {
	return this.makeError(fmt.Errorf("unexpected token %s: %q", scanned(t.Kind), string(t.Body)))
}

func (this *Parser) Discard() error {
	switch this.curr.Kind {
	case ARR_OPEN:
		if arr, err := this.ReadArray(); err != nil {
			return err
		} else {
			return arr.DiscardRemaining()
		}
	case OBJ_OPEN:
		if obj, err := this.ReadObject(); err != nil {
			return err
		} else {
			return obj.DiscardRemaining()
		}
	default:
		this.Next()
		return nil
	}
}

func (this *Parser) Next() bool {
	defer func() {
		if r := recover(); r != nil {
			var panicwith error
			switch r := r.(type) {
			case error:
				panicwith = this.makeError(r)
			default:
				panicwith = this.scanner.makePositionedError(slipup.Createf("%v", r))
			}
			panic(panicwith)
		}
	}()

	if this.peek.Kind == EOF {
		return false
	}

	for {
		lexeme, err := this.scanner.Next()
		internal.PanicOnError(err)

		if lexeme.scanned == scanned_comment {
			continue
		}

		this.curr = this.peek
		this.peek.Kind = Kind(lexeme.scanned)
		this.peek.Body = make([]byte, len(lexeme.body))
		copy(this.peek.Body, lexeme.body)
		return true
	}
}

func (this *Parser) Current() Token {
	return this.curr
}

func (this *Parser) Peek() Token {
	return this.peek
}

func (this *Parser) ReadObject() (*ObjectParser, error) {
	_, err := this.expect(OBJ_OPEN)
	if err != nil {
		return nil, err
	}

	this.Next()
	return &ObjectParser{this}, nil
}

func (this *Parser) ReadArray() (*ArrayParser, error) {
	_, err := this.expect(ARR_OPEN)
	if err != nil {
		return nil, err
	}
	this.Next()
	return &ArrayParser{this}, nil
}

func (this *Parser) ReadString() (string, error) {
	token, err := this.expect(STRING)
	if err != nil {
		return "", err
	}

	this.Next()
	return string(token.Body), nil
}

func (this *Parser) ReadInt() (int, error) {
	token, err := this.expect(NUMBER)
	if err != nil {
		return 0, err
	}

	number, err := strconv.Atoi(string(token.Body))
	if err != nil {
		return 0, this.scanner.makePositionedError(fmt.Errorf("failed to parse number: %w", err))
	}

	this.Next()
	return number, nil
}

func (this *Parser) ReadFloat() (float64, error) {
	token, err := this.expect(NUMBER)
	if err != nil {
		return 0, err
	}

	number, err := strconv.ParseFloat(string(token.Body), 64)
	if err != nil {
		return 0, this.scanner.makePositionedError(fmt.Errorf("failed to parse number: %w", err))
	}
	this.Next()
	return number, nil
}

func (this *Parser) ReadBool() (bool, error) {
	if this.curr.Kind == TRUE {
		this.Next()
		return true, nil
	} else if this.curr.Kind == FALSE {
		this.Next()
		return false, nil
	} else {
		return false, this.Unexpected(&this.curr)
	}
}

func (this *Parser) expect(expected Kind) (Token, error) {
	if this.curr.Kind == EOF {
		return this.curr, io.EOF
	}

	if this.curr.Kind != expected {
		return Token{Kind: ERR}, this.Unexpected(&this.curr)
	}

	return this.curr, nil
}

func maybeReadComma(parser *Parser) {
	if parser.curr.Kind == comma {
		parser.Next()
	}
}
