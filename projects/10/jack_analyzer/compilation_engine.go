package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

type CompilationEngine struct {
	err    error
	output *bufio.Writer

	Tokenizer
}

func NewCompilationEngine(reader io.Reader, writer io.Writer) CompilationEngine {
	return CompilationEngine{
		Tokenizer: NewTokenizer(reader),
		output:    bufio.NewWriter(writer),
	}
}

func (c *CompilationEngine) CompileClass() {
	c.expectKeyword("class")
	c.writeLeftLabel("class")
	c.writeLabelValue("keyword", "class")
	identifier := c.expectIdentifier()
	c.writeLabelValue("identifier", identifier)
	c.expectSymbol("{")
	c.writeLabelValue("symbol", "{")
	c.CompileClassVarDec()
	c.CompileSubroutineDec()
	c.expectSymbol("}")
	c.writeLabelValue("symbol", "}")
	c.writeRightLabel("class")
}

func (c *CompilationEngine) CompileClassVarDec() {
	if c.err != nil {
		return
	}
	keyword := c.expectKeywords([]string{string(FIELD), string(STATIC)})
	c.CompileType()
	c.CompileVarName()
	// c.expectSymbol()

}

func (c *CompilationEngine) CompileSubroutineDec() {
	if c.err != nil {
		return
	}
}

func (c *CompilationEngine) CompileType() {
	if c.err != nil {
		return
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("expect type, got err: %v", err)
		return
	}
	if c.TokenType() == KEYWORD {
		switch c.Keyword() {
		case INT:
		case CHAR:
		case BOOLEAN:
		default:
			c.err = fmt.Errorf("expect type, got: %v", c.Keyword())
			return
		}
	}
	if c.TokenType() == IDENTIFIER {

	}
}

func (c *CompilationEngine) CompileVarName() {
	if c.err != nil {
		return
	}
	c.expectIdentifier()
}

func (c *CompilationEngine) expectKeyword(keyword string) {
	if c.err != nil {
		return
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("expect keyword '%s', got err: %v", keyword, err)
		return
	}
	if c.TokenType() != KEYWORD || c.Keyword() != Keyword(keyword) {
		c.err = fmt.Errorf("expect keyword, got %s", c.Val())
		return
	}
}

func (c *CompilationEngine) expectKeywords(keyword []string) Keyword {
	if c.err != nil {
		return ""
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("expect keyword '%s', got err: %v", keyword, err)
		return ""
	}
	if c.TokenType() != KEYWORD {
		c.err = fmt.Errorf("expect keyword, got %s", c.Val())
		return ""
	}
	for _, k := range keyword {
		if k == string(c.Keyword()) {
			return c.Keyword()
		}
	}
	c.err = fmt.Errorf("expect keyword: %s, got %s", strings.Join(keyword, "|"), c.Val())
	return ""
}

func (c *CompilationEngine) expectIdentifier() string {
	if c.err != nil {
		return ""
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("expect identifier, got err: %v", err)
		return ""
	}
	if c.TokenType() != IDENTIFIER {
		c.err = fmt.Errorf("expect identifier, got %s", c.Val())
		return ""
	}
	return c.Identifier()
}

func (c *CompilationEngine) expectSymbol(symbol string) {
	if c.err != nil {
		return
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("expect symbol '%s', got err: %v", symbol, err)
		return
	}
	if c.TokenType() != SYMBOL || c.Symnol() != symbol {
		c.err = fmt.Errorf("expect symbol '%s', got: %v", symbol, c.Val())
		return
	}
}

func (c *CompilationEngine) moveNextToken() error {
	if c.HasMoreTokens() {
		return c.Advance()
	}
	return nil
}

func (c *CompilationEngine) writeLeftLabel(label string) {
	c.output.WriteString(fmt.Sprintf("<%s>\n", label))
}

func (c *CompilationEngine) writeRightLabel(label string) {
	c.output.WriteString(fmt.Sprintf("</%s>\n", label))
}

func (c *CompilationEngine) writeValue(value string) {
	c.output.WriteString(value)
}

func (c *CompilationEngine) writeLabelValue(label string, value string) {
	c.output.WriteString(fmt.Sprintf("<%s>", label))
	c.output.WriteString(value)
	c.output.WriteString(fmt.Sprintf("</%s>\n", label))
}
