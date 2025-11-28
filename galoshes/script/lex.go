package script

import (
	"iter"
	"strings"

	"github.com/etc-sudonters/substrate/peruse"
)

type TokenType = peruse.TokenType

const (
	ERR = peruse.ERR
	EOF = peruse.EOF

	SPACE_CHARS = 0

	BAD_TOKEN TokenType = iota
	TOKEN_FIND
	TOKEN_WITH
	TOKEN_WHERE
	TOKEN_INSERT
	TOKEN_RULES
	TOKEN_TRUE
	TOKEN_FALSE
	TOKEN_COMMA         // ,
	TOKEN_OPEN_BRACKET  // [
	TOKEN_CLOSE_BRACKET // ]
	TOKEN_ASSIGN        // :-
	TOKEN_RULE          // :[a-z-]+
	TOKEN_VARIABLE      // \$[a-z-]+
	TOKEN_ATTRIBUTE     // [a-z-]+(/[a-z-])*
	TOKEN_STRING        // "[^"]"
	TOKEN_NUMBER        // [0-9]+(\.[0-9]+)
	TOKEN_COMMENT       // ;.*$

	eof        = -1
	spaceChars = " \t\n\r"
	findWord   = "find"
	withWord   = "with"
	whereWord  = "where"
	insertWord = "insert"
	rulesWord  = "rules"
	trueWord   = "true"
	falseWord  = "false"
	attrSep    = '/'
)

func TokenTypeString(t TokenType) string {
	switch t {
	case TOKEN_FIND:
		return "<FIND>"
	case TOKEN_WITH:
		return "<WITH>"
	case TOKEN_WHERE:
		return "<WHERE>"
	case TOKEN_INSERT:
		return "<INSERT>"
	case TOKEN_RULES:
		return "<RULES>"
	case TOKEN_TRUE:
		return "<TRUE>"
	case TOKEN_FALSE:
		return "<FALSE>"
	case TOKEN_COMMA:
		return "<COMMA>"
	case TOKEN_OPEN_BRACKET:
		return "<OPEN_BRACKET>"
	case TOKEN_CLOSE_BRACKET:
		return "<CLOSE_BRACKET>"
	case TOKEN_ASSIGN:
		return "<ASSIGN>"
	case TOKEN_RULE:
		return "<RULE>"
	case TOKEN_VARIABLE:
		return "<VARIABLE>"
	case TOKEN_ATTRIBUTE:
		return "<ATTRIBUTE>"
	case TOKEN_STRING:
		return "<STRING>"
	case TOKEN_NUMBER:
		return "<NUMBER>"
	case TOKEN_COMMENT:
		return "<COMMENT>"
	default:
		return "<UNKNOWN>"
	}
}

func Tokens(l *peruse.StringLexer) iter.Seq[peruse.Token] {
	return func(yield func(peruse.Token) bool) {
		for {
			token := l.NextToken()
			if token.Is(EOF) {
				return
			}
			if !yield(token) {
				return
			}
			if token.Is(ERR) {
				return
			}
		}
	}
}

type lexState struct {
	brackDepth int
}

func NewLexer(script string) *peruse.StringLexer {
	return peruse.NewLexer(script, lexScript, &lexState{})
}

func lexScript(l *peruse.StringLexer, state any) peruse.LexFn {
	s := state.(*lexState)
	r := l.Next()

	if isWhitespace(r) {
		return lexWhitespace
	}

	if isDigit(r) {
		return lexNumber
	}

	if isLetter(r) {
		return lexWord
	}

	switch r {
	case eof:
		if s.brackDepth > 0 {
			return l.Error("unclosed '['")
		}
		return nil
	case '[':
		s.brackDepth++
		return l.Emit(TOKEN_OPEN_BRACKET)
	case ']':
		s.brackDepth--
		return l.Emit(TOKEN_CLOSE_BRACKET)
	case ',':
		return l.Emit(TOKEN_COMMA)
	case '"':
		return lexString
	case '$':
		return lexVariable
	case ':':
		if l.Peek() == '-' {
			l.Next()
			return l.Emit(TOKEN_ASSIGN)
		}
		if isIdentRune(l.Peek()) {
			return lexInvoke
		}
		r = l.Next() // put correct character into error response
	case ';':
		return lexComment
	}
	return l.Error("unrecongized character %#U", r)
}

func lexWhitespace(l *peruse.StringLexer, _ any) peruse.LexFn {
	l.AcceptWhile(isWhitespace)
	l.Discard()
	return lexScript
}

func lexNumber(l *peruse.StringLexer, _ any) peruse.LexFn {
	l.AcceptWhile(isDigit)

	if l.Peek() == '.' {
		l.Next()
		l.AcceptWhile(isDigit)
	}

	if !atSeparator(l) {
		return unexpected(l.Peek(), l)
	}
	return l.Emit(TOKEN_NUMBER)
}

// opening " is already scanned
func lexString(l *peruse.StringLexer, _ any) peruse.LexFn {
	l.Discard()
	l.AcceptWhile(func(r rune) bool { return r != '"' })
	// sheer off ending "
	next := l.Emit(TOKEN_STRING)
	l.Next()
	l.Discard()
	return next
}

func lexComment(l *peruse.StringLexer, _ any) peruse.LexFn {
	l.AcceptWhile(func(r rune) bool { return r != '\n' })
	l.Next()
	l.Discard()
	return lexScript
}

func lexWord(l *peruse.StringLexer, _ any) peruse.LexFn {
	l.AcceptWhile(isIdentRune)
	if l.Peek() == attrSep {
		return lexAttr
	}

	word := l.Word()
	switch word {
	case findWord:
		return l.Emit(TOKEN_FIND)
	case withWord:
		return l.Emit(TOKEN_WITH)
	case whereWord:
		return l.Emit(TOKEN_WHERE)
	case insertWord:
		return l.Emit(TOKEN_INSERT)
	case rulesWord:
		return l.Emit(TOKEN_RULES)
	case trueWord:
		return l.Emit(TOKEN_TRUE)
	case falseWord:
		return l.Emit(TOKEN_FALSE)
	default:
		// if it's a bareword and not a keyword, it's an attribute
		return l.Emit(TOKEN_ATTRIBUTE)
	}
}

func lexAttr(l *peruse.StringLexer, _ any) peruse.LexFn {
	l.AcceptWhile(isAttrRune)
	return l.Emit(TOKEN_ATTRIBUTE)
}

func lexVariable(l *peruse.StringLexer, _ any) peruse.LexFn {
	l.Discard()
	l.AcceptWhile(isIdentRune)
	return l.Emit(TOKEN_VARIABLE)
}

func lexInvoke(l *peruse.StringLexer, _ any) peruse.LexFn {
	l.Discard()
	l.AcceptWhile(isIdentRune)
	return l.Emit(TOKEN_RULE)
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isLetter(r rune) bool {
	return 'a' <= r && r <= 'z'
}

func isAttrRune(r rune) bool {
	return isIdentRune(r) || r == attrSep
}

func isIdentRune(r rune) bool {
	return isLetter(r) || r == '-' || isDigit(r)
}

func isWhitespace(r rune) bool {
	return strings.ContainsRune(spaceChars, r)
}

func atSeparator(l *peruse.StringLexer) bool {
	r := l.Peek()
	if isWhitespace(r) {
		return true
	}

	switch r {
	case eof, '(', ')', ',', '[', ']', ':', ';':
		return true
	default:
		return false
	}
}

func unexpected(r rune, l *peruse.StringLexer) peruse.LexFn {
	return l.Error("unexpected %q", r)
}
