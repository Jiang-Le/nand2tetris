package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

type Token struct {
	val       string
	tokenType TokenType
}

type TokenType string

const (
	KEYWORD        TokenType = "keyword"
	SYMBOL         TokenType = "symbol"
	IDENTIFIER     TokenType = "identifier"
	INT_CONST      TokenType = "int_const"
	STRING_CONST   TokenType = "string_const"
	EOF            TokenType = "eof"
	ERR_IDENTIFIER TokenType = "error_identifier"
)

type Keyword string

const (
	CLASS       Keyword = "class"
	CONSTRUCTOR Keyword = "constructor"
	FUNCTION    Keyword = "function"
	METHOD      Keyword = "method"
	FIELD       Keyword = "field"
	STATIC      Keyword = "static"
	VAR         Keyword = "var"
	INT         Keyword = "int"
	CHAR        Keyword = "char"
	BOOLEAN     Keyword = "boolean"
	VOID        Keyword = "void"
	TRUE        Keyword = "true"
	FALSE       Keyword = "false"
	NULL        Keyword = "null"
	THIS        Keyword = "this"
	LET         Keyword = "let"
	DO          Keyword = "do"
	IF          Keyword = "if"
	ELSE        Keyword = "else"
	WHILE       Keyword = "while"
	RETURN      Keyword = "return"
)

var (
	_keywords = []string{
		"class", "constructor", "function", "method",
		"field", "static", "var", "int", "char", "boolean",
		"void", "true", "false", "null", "this", "let", "do",
		"if", "else", "while", "return",
	}
	_symbols = "{}()[].,;+-*/&|<>=~"
)

type Tokenizer struct {
	eof       bool
	bufReader *bufio.Reader
	token     Token
}

func NewTokenizer(reader io.Reader) Tokenizer {
	bufReader := bufio.NewReader(reader)
	tokenizer := Tokenizer{
		eof:       false,
		bufReader: bufReader,
	}
	return tokenizer
}

func (t *Tokenizer) HasMoreTokens() bool {
	return !t.eof
}

func (t *Tokenizer) Advance() error {
	var r rune
	var err error
	// 跳过所有的space
	for {
		r, err = t.nextChar()
		if err != nil {
			t.eof = true
			t.token = Token{
				tokenType: EOF,
			}
			return nil
		}
		if isSpace(r) {
			continue
		}
		if r == '/' {
			isComment, err := t.trySkipComment()
			if err != nil {
				return err
			}
			if isComment {
				continue
			}
		}
		break
	}

	token := []rune{
		r,
	}
	if isSymbol(r) {
		t.token = Token{
			val:       string(token),
			tokenType: SYMBOL,
		}
		return nil
	}

	if r == '"' {
		for r, err = t.nextChar(); r != '"' && err == nil; r, err = t.nextChar() {
			token = append(token, r)
		}
		if err != nil {
			return err
		}
		token = append(token, r)
		t.token = Token{
			val:       string(token),
			tokenType: STRING_CONST,
		}
		return nil
	}

	for {
		c, err := t.nextChar()
		if err != nil {
			// 此次EOF不用设置，等下次advance时才是EOF
			break
		}
		if isSpace(c) || isSymbol(c) {
			t.unreadChar()
			break
		}
		token = append(token, c)
	}
	tokenType := t.parseToken(token)
	t.token = Token{
		val:       string(token),
		tokenType: tokenType,
	}
	return nil
}

func (t *Tokenizer) trySkipComment() (bool, error) {
	c, err := t.nextChar()
	if err != nil {
		return false, nil
	}
	if c == '/' {
		if err := t.readUntil('\n'); err != nil {
			return false, err
		}
		c, err = t.nextChar()
		if c != '\r' && err == nil {
			t.unreadChar()
		}
	} else if c == '*' {
		for {
			if err := t.readUntil('*'); err != nil {
				return false, err
			}
			c, err = t.nextChar()
			if err != nil {
				return false, fmt.Errorf("unexpect end of comment")
			}
			if c == '/' {
				break
			}
		}
	} else {
		t.unreadChar()
		return false, nil
	}

	return true, nil
}

func (t *Tokenizer) TokenType() TokenType {
	return t.token.tokenType
}

func (t *Tokenizer) Keyword() Keyword {
	return Keyword(t.token.val)
}

func (t *Tokenizer) Symnol() string {
	return t.token.val
}

func (t *Tokenizer) Identifier() string {
	return t.token.val
}

func (t *Tokenizer) IntVal() int64 {
	v, _ := strconv.ParseInt(t.token.val, 10, 64)
	return v
}

func (t *Tokenizer) StringVal() string {
	return t.token.val[1 : len(t.token.val)-1]
}

func (t *Tokenizer) Val() string {
	return t.token.val
}

func (t *Tokenizer) parseToken(rs []rune) TokenType {
	if isKeyWord(rs) {
		return KEYWORD
	} else if rs[0] >= '0' || rs[0] <= '9' {
		_, err := strconv.ParseInt(string(rs), 10, 64)
		if err == nil {
			return INT_CONST
		}
	} else if rs[0] == '"' && rs[len(rs)-1] == '"' {
		return STRING_CONST
	}
	return IDENTIFIER
}

func (t *Tokenizer) nextChar() (rune, error) {
	r, _, err := t.bufReader.ReadRune()
	return r, err
}

func (t *Tokenizer) unreadChar() {
	t.bufReader.UnreadRune()
}

func (t *Tokenizer) readUntil(r rune) error {
	for {
		v, err := t.nextChar()
		if err != nil {
			return err
		}
		if v == r {
			return nil
		}
	}
}

func isSymbol(r rune) bool {
	return strings.ContainsRune(_symbols, r)
}

func isSpace(c rune) bool {
	return unicode.IsSpace(c)
}

func isKeyWord(rs []rune) bool {
	for _, v := range _keywords {
		if v == string(rs) {
			return true
		}
	}
	return false
}
