package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html"
	"io"
	"log"
	"strings"
)

type CompilationEngine struct {
	err    error
	output *bufio.Writer

	Tokenizer
	tokens []Token
	curTokenIndex int
	curToken *Token 

	inTx bool
	txBeginTokenIndex int
	txOutput *bytes.Buffer

}

func NewCompilationEngine(reader io.Reader, writer io.Writer) CompilationEngine {
	return CompilationEngine{
		Tokenizer: NewTokenizer(reader),
		output:    bufio.NewWriter(writer),
		curTokenIndex: -1,
	}
}

func (c *CompilationEngine) Error() error {
	return c.err
}

func (c *CompilationEngine) CompileClass() {
	defer c.output.Flush()

	log.Printf("CompileClass")
	c.writeLeftLabel("class")
	defer c.writeRightLabel("class")

	c.expectKeyword("class")
	c.CompileClassName()
	c.expectSymbol("{")
	for c.HaveClassVarDec() {
		c.CompileClassVarDec()
	}
	for c.HaveSubroutineDec() {
		c.CompileSubroutineDec()
	}
	c.expectSymbol("}")
	
	c.expectEOF()

}

func (c *CompilationEngine) HaveClassVarDec() bool {
	if c.err != nil {
		return false
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("check have class var dec err: %v", err)
		return false
	}
	defer c.unreadCurToken()

	if c.TokenType() == KEYWORD {
		tokenVal := c.Keyword() 
		return tokenVal == FIELD || tokenVal == STATIC
	} 
	return false
}

func (c *CompilationEngine) CompileClassVarDec() {
	if c.err != nil {
		return
	}

	log.Printf("CompileClassVarDec")

	c.writeLeftLabel("classVarDec")
	defer c.writeRightLabel("classVarDec")

	c.expectKeywords([]string{string(FIELD), string(STATIC)})
	c.CompileType()
	c.CompileVarName()
	for c.checkSymbol(",") {
		c.expectSymbol(",")
		c.CompileVarName()
	}
	c.expectSymbol(";")
}

func (c *CompilationEngine) HaveSubroutineDec() bool {
	if c.err != nil {
		return false
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("check have subroutine dec err: %v", err)
		return false
	}
	defer c.unreadCurToken()

	if c.TokenType() == KEYWORD {
		tokenVal := c.Keyword() 
		return tokenVal == CONSTRUCTOR || tokenVal == FUNCTION || tokenVal == METHOD
	} 
	return false
}

func (c *CompilationEngine) CompileSubroutineDec() {
	if c.err != nil {
		return
	}

	log.Printf("CompileSubroutineDec")
	c.writeLeftLabel("subroutineDec")
	defer c.writeRightLabel("subroutineDec")

	c.expectKeywords([]string{string(CONSTRUCTOR), string(FUNCTION), string(METHOD)})
	if c.checkKeyword("void") {
		c.expectKeyword("void")
	} else {
		c.CompileType()
	}
	c.CompileSubroutineName()
	c.expectSymbol("(")

	c.CompileParameterList()

	c.expectSymbol(")")

	c.CompileSubroutineBody()
}

func (c *CompilationEngine) HaveParameterList() bool {
	if c.err != nil {
		return false
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("check have parameter list err: %v", err)
		return false
	}
	defer c.unreadCurToken()
	return c.TokenType() != SYMBOL
}

func (c *CompilationEngine) CompileParameterList() {
	if c.err != nil {
		return
	}

	c.writeLeftLabel("parameterList")
	defer c.writeRightLabel("parameterList")
	log.Printf("CompileParameterList")

	if !c.HaveParameterList() {
		return
	}

	c.CompileType()
	c.CompileVarName()
	for c.checkSymbol(",") {
		c.expectSymbol(",")
		c.CompileType()
		c.CompileVarName()
	}
}

func (c *CompilationEngine) CompileSubroutineBody() {
	if c.err != nil {
		return
	}

	log.Printf("CompileSubroutineBody")

	c.writeLeftLabel("subroutineBody")
	defer c.writeRightLabel("subroutineBody")

	c.expectSymbol("{")

	for c.checkKeyword("var") {
		c.CompileVarDec()
	}

	c.CompileStatements("}")
	
	c.expectSymbol("}")
}

func (c *CompilationEngine) CompileStatements(symbol string) {
	if c.err != nil {
		return
	}

	log.Printf("CompileStatements")

	c.writeLeftLabel("statements")
	defer c.writeRightLabel("statements")

	for !c.checkSymbol(symbol) && c.Error() == nil {
		c.CompileStatement()
	}
}

func (c *CompilationEngine) CompileStatement() {
	if c.err != nil {
		return
	}

	log.Printf("CompileStatement")

	if c.checkKeyword("let") {
		c.CompileLetStatement()
	} else if c.checkKeyword("if") {
		c.CompileIfStatement()
	} else if c.checkKeyword("while") {
		c.CompileWhileStatement()
	} else if c.checkKeyword("do") {
		c.CompileDoStatement()
	} else if c.checkKeyword("return") {
		c.CompileReturnStatement()
	} else {
		c.err = fmt.Errorf("expect statement, token: %s", c.currentToken())
	}
}

func (c *CompilationEngine) CompileLetStatement() {
	
	if c.err != nil {
		return
	}

	log.Printf("CompileLetStatement")
	c.writeLeftLabel("letStatement")
	defer c.writeRightLabel("letStatement")

	c.expectKeyword("let")

	c.CompileVarName()

	if c.checkSymbol("[") {
		c.expectSymbol("[")
		c.CompileExpression("]")
		c.expectSymbol("]")
	}
	c.expectSymbol("=")
	c.CompileExpression(";")
	c.expectSymbol(";")
}

func (c *CompilationEngine) CompileIfStatement() {
	
	if c.err != nil {
		return
	}

	log.Printf("CompileIfStatement")
	c.writeLeftLabel("ifStatement")
	defer c.writeRightLabel("ifStatement")

	c.expectKeyword("if")

	c.expectSymbol("(")

	c.CompileExpression(")")
	
	c.expectSymbol(")")
	
	c.expectSymbol("{")
	
	c.CompileStatements("}")
	
	c.expectSymbol("}")

	if c.checkKeyword("else") {
		c.expectKeyword("else")
		
		c.expectSymbol("{")

		c.CompileStatements("}")

		c.expectSymbol("}")
	}
}

func (c *CompilationEngine) CompileWhileStatement() {
	
	if c.err != nil {
		return
	}

	log.Printf("CompileWhileStatement")
	c.writeLeftLabel("whileStatement")
	defer c.writeRightLabel("whileStatement")

	c.expectKeyword("while")
	
	c.expectSymbol("(")

	c.CompileExpression(")")
	
	c.expectSymbol(")")

	c.expectSymbol("{")
	
	c.CompileStatements("}")
	
	c.expectSymbol("}")
}

func (c *CompilationEngine) CompileDoStatement() {

	if c.err != nil {
		return
	}

	log.Printf("CompileDoStatement")
	c.writeLeftLabel("doStatement")
	defer c.writeRightLabel("doStatement")

	c.expectKeyword("do")

	c.CompileSubroutineCall()

	c.expectSymbol(";")
}

func (c *CompilationEngine) CompileReturnStatement() {

	if c.err != nil {
		return
	}

	log.Printf("CompileReturnStatement")
	c.writeLeftLabel("returnStatement")
	defer c.writeRightLabel("returnStatement")

	c.expectKeyword("return")

	if !c.checkSymbol(";") {
		c.CompileExpression(";")
	}

	c.expectSymbol(";")
}

func ( c *CompilationEngine) CompileExpression(endSymbols ...string) {

	if c.err != nil {
		return
	}

	log.Printf("CompileExpression")
	c.writeLeftLabel("expression")
	defer c.writeRightLabel("expression")

	c.CompileTerm()
	for !c.checkSymbols(endSymbols) && c.Error() == nil {
		c.CompileOp()
		c.CompileTerm()
	}
}

func (c *CompilationEngine) CompileTerm() {
	if c.err != nil {
		return 
	}

	log.Printf("CompileTerm")
	c.writeLeftLabel("term")
	defer c.writeRightLabel("term")

	if c.checkTokenType(INT_CONST) {
		c.expectIntegerConstant()
		return
	} else if (c.checkTokenType(STRING_CONST)) {
		c.expectStringConstant()
		return
	} else if (c.checkKeywords([]string{"true", "false", "null", "this"})) {
		c.expectKeywords([]string{"true", "false", "null", "this"})
		return
	} else if (c.checkSymbols([]string{"-", "~"})){
		c.expectSymbols([]string{"-", "~"})
		c.CompileTerm()
		return
	} else if (c.checkSymbol("(")) {
		c.expectSymbol("(")
		
		c.CompileExpression(")")

		c.expectSymbol(")")
		return
	}
	c.begin()
	c.expectIdentifier()
	if c.checkSymbol("[") { // 变量带下标
		c.commit()
		c.expectSymbol("[")

		c.CompileExpression("]")

		c.expectSymbol("]")
	} else if c.checkSymbols([]string{"(", "."}) { // 函数调用
		c.rollback() // 回退之前的解析，重新按SubroutineCall进行解析
		c.CompileSubroutineCall()
	} else { // 普通变量
		c.commit()
	}
}

func (c *CompilationEngine) CompileSubroutineCall() {
	if c.err != nil {
		return
	}

	log.Printf("CompileSubroutineCall")

	c.expectIdentifier() // 可能是函数名、类名或者变量名
	if c.checkSymbol(".") {
		c.expectSymbol(".")
		c.CompileSubroutineName()
	}

	c.expectSymbol("(")

	c.CompileExpressionList(")")
	
	c.expectSymbol(")")
}

func (c *CompilationEngine) CompileExpressionList(endSymbol string) {
	if c.err != nil {
		return
	}

	log.Printf("CompileExpressionList")
	c.writeLeftLabel("expressionList")
	defer c.writeRightLabel("expressionList")

	for !c.checkSymbol(endSymbol) && c.Error() == nil {
		c.CompileExpression(endSymbol, ",")
		if c.checkSymbol(",") {
			c.expectSymbol(",")
		}
	}
}

func (c *CompilationEngine) CompileOp() {

	if c.err != nil {
		return
	}

	log.Printf("CompileOp")

	c.expectSymbols([]string{
		"+", "-", "*", "/", "&", "|", "<", ">", "=",
	})
}

func (c *CompilationEngine) CompileUnaryOp() {
	if c.err != nil {
		return
	}

	log.Printf("CompileUnaryOp")

	c.expectSymbols([]string{
		"-", "~",
	})
}

func (c *CompilationEngine) CompileKeywordConstant() {
	if c.err != nil {
		return
	}

	log.Printf("CompileKeywordConstant")

	c.expectKeywords([]string {
		"true", "false", "null","this",
	})
}

func (c *CompilationEngine) CompileVarDec() {
	if c.err != nil {
		return 
	}

	log.Printf("CompileVarDec")
	c.writeLeftLabel("varDec")
	defer c.writeRightLabel("varDec")

	c.expectKeyword("var")
	c.CompileType()
	c.CompileVarName()
	for c.checkSymbol(",") && c.Error() == nil {
		c.expectSymbol(",")
		c.CompileVarName()
	}
	c.expectSymbol(";")
}

func (c *CompilationEngine) CompileType() {
	if c.err != nil {
		return
	}

	log.Printf("CompileType")

	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("expect type, got err: %v", err)
		return
	}
	if c.TokenType() == KEYWORD {
		switch c.Keyword() {
		case INT:
			fallthrough
		case CHAR:
			fallthrough
		case BOOLEAN:
			c.writeLabelValue("keyword", string(c.Keyword()))
			return
		default:
			c.err = fmt.Errorf("expect type, got: %v, token: %s", c.Keyword(), c.currentToken())
			return
		}
	}
	if c.TokenType() == IDENTIFIER {
		c.writeLabelValue("identifier", c.Identifier())
		return
	}
	c.err = fmt.Errorf("expect type, got type: %s, val: %s, token: %s", c.TokenType(), c.Val(), c.currentToken())
}

func (c *CompilationEngine) CompileVarName() {
	if c.err != nil {
		return
	}

	log.Printf("CompileVarName")

	c.expectIdentifier()
}

func (c *CompilationEngine) CompileSubroutineName() {
	if c.err != nil {
		return
	}
	c.expectIdentifier()
}

func (c *CompilationEngine) CompileClassName() {
	if c.err != nil {
		return
	}
	c.expectIdentifier()
}


func (c *CompilationEngine) currentToken() *Token {
	return c.curToken
}

func (c *CompilationEngine) begin() {
	c.inTx = true
	c.txBeginTokenIndex = c.curTokenIndex
	c.txOutput = new(bytes.Buffer)
}

func (c *CompilationEngine) rollback() {
	if !c.inTx {
		panic("cannot rollback, not in tx")
	}
	c.inTx = false
	c.curTokenIndex = c.txBeginTokenIndex
}

func (c *CompilationEngine) commit() {
	if !c.inTx {
		panic("cannot commit, not in tx")
	}
	c.inTx = false
	c.output.WriteString(c.txOutput.String())
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
		c.err = fmt.Errorf("expect keyword, got %s, line: %d, col: %d", c.Val(), c.currentToken().line, c.currentToken().col)
		return
	}

	c.writeLabelValue("keyword", keyword)
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
		c.err = fmt.Errorf("expect keyword, got %s, line: %d, col: %d", c.Val(), c.currentToken().line, c.currentToken().col)
		return ""
	}
	for _, k := range keyword {
		if k == string(c.Keyword()) {
			c.writeLabelValue("keyword", k)
			return c.Keyword()
		}
	}
	c.err = fmt.Errorf("expect keyword: %s, got %s, line: %d, col: %d", strings.Join(keyword, "|"), c.Val(), c.currentToken().line, c.currentToken().col)
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
		c.err = fmt.Errorf("expect identifier, got %s, line: %d, col: %d", c.Val(), c.currentToken().line, c.currentToken().col)
		return ""
	}

	c.writeLabelValue("identifier", c.Identifier())
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
	if c.TokenType() != SYMBOL || c.Symbol() != symbol {
		c.err = fmt.Errorf("expect symbol '%s', got: %v, line: %d, col: %d", symbol, c.Val(), c.currentToken().line, c.currentToken().col)
		return
	}
	c.writeLabelValue("symbol", html.EscapeString(c.Symbol()))
}

func (c *CompilationEngine) expectSymbols(symbols []string) string {
	if c.err != nil {
		return ""
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("expect symbols: %v, got err: %v", symbols, err)
		return ""
	}
	if c.TokenType() != SYMBOL {
		c.err = fmt.Errorf("expect symbol, got: %s, line: %d, col: %d", c.TokenType(), c.currentToken().line, c.currentToken().col)
	}
	for _, symbol := range symbols {
		if symbol == c.Symbol() {
			c.writeLabelValue("symbol", html.EscapeString(c.Symbol()))
			return symbol
		}
	}
	c.err = fmt.Errorf("expect symbols: %v, got: %s, line: %d, col: %d", symbols, c.Symbol(), c.currentToken().line, c.currentToken().col)
	return ""
}

func (c *CompilationEngine) expectIntegerConstant() {
	if c.err != nil {
		return 
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("expect interger constant, got err: %v", err)
		return 
	}
	if c.TokenType() != INT_CONST {
		c.err = fmt.Errorf("expect interger constant, got: %s, line: %d, col: %d", c.TokenType(), c.currentToken().line, c.currentToken().col)
		return
	}
	c.writeLabelValue("integerConstant", fmt.Sprintf("%d", c.IntVal()))
}

func (c *CompilationEngine) expectStringConstant() {
	if c.err != nil {
		return 
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("expect string constant, got err: %v", err)
		return 
	}
	if c.TokenType() != STRING_CONST {
		c.err = fmt.Errorf("expect string constant, got: %s, line: %d, col: %d", c.TokenType(), c.currentToken().line, c.currentToken().col)
		return
	}
	c.writeLabelValue("stringConstant", c.StringVal())
}

func (c *CompilationEngine) expectEOF() {
	if c.err != nil {
		return 
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("expect eof, got err: %v", err)
		return
	}
	if c.TokenType() != EOF {
		c.err = fmt.Errorf("expect eof, got: %s", c.TokenType())
		return
	}
}

func (c *CompilationEngine) checkTokenType(tokenType TokenType) bool {
	if c.err != nil {
		return false
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("check interger constant, got err: %v", err)
		return false
	}
	defer c.unreadCurToken()
	return c.TokenType() == tokenType
}

func (c *CompilationEngine) checkKeyword(keyword string) bool {
	if c.err != nil {
		return false
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("check keyword '%s', got err: %v", keyword, err)
		return false
	}
	defer c.unreadCurToken()
	if c.TokenType() != KEYWORD || c.Keyword() != Keyword(keyword) {
		return false
	}
	return true
}

func (c *CompilationEngine) checkKeywords(keywords []string) bool {
	if c.err != nil {
		return false
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("check keyword %v, got err: %v", keywords, err)
		return false
	}
	defer c.unreadCurToken()
	if c.TokenType() != KEYWORD {
		return false
	}
	for _, keyword := range keywords {
		if c.Keyword() == Keyword(keyword) {
			return true
		}
	}
	return false
}

func (c *CompilationEngine) checkSymbol(symbol string) bool {
	if c.err != nil {
		return false
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("check symbol '%s', got err: %v", symbol, err)
		return false
	}
	defer c.unreadCurToken()
	if c.TokenType() != SYMBOL || c.Symbol() != symbol {
		return false
	}
	return true
}

func (c *CompilationEngine) checkSymbols(symbols []string) bool {
	if c.err != nil {
		return false
	}
	if err := c.moveNextToken(); err != nil {
		c.err = fmt.Errorf("check symbols %v, got err: %v", symbols, err)
		return false
	}
	defer c.unreadCurToken()
	if c.TokenType() != SYMBOL {
		return false
	}
	for _, symbol := range symbols {
		if c.Symbol() == symbol {
			return true
		}
	}
	return false
}

func (c *CompilationEngine) moveNextToken() error {
	if c.curTokenIndex < len(c.tokens) - 1 {
		c.curTokenIndex += 1
		c.curToken = &c.tokens[c.curTokenIndex]
		return nil
	}
	
	if c.HasMoreTokens() {
		if err := c.Advance(); err != nil {
			return err
		}
		c.tokens = append(c.tokens, c.token)
		c.curTokenIndex += 1
		c.curToken = &c.tokens[c.curTokenIndex]
	}
	return nil
}

func (c *CompilationEngine) unreadCurToken() {
	if c.curTokenIndex <= -1 {
		panic("no more token to unread")
	}
	c.curTokenIndex -= 1
	c.curToken =  &c.tokens[c.curTokenIndex]
}

func (c *CompilationEngine) TokenType() TokenType {
	return c.curToken.TokenType()
}

func (c *CompilationEngine) Keyword() Keyword {
	return c.curToken.Keyword()
}

func (c *CompilationEngine) Symbol() string {
	return c.curToken.Symbol()
}

func (c *CompilationEngine) Identifier() string {
	return c.curToken.Identifier()
}

func (c *CompilationEngine) IntVal() int64 {
	return c.curToken.IntVal()
}

func (c *CompilationEngine) StringVal() string {
	return c.curToken.StringVal()
}

func (c *CompilationEngine) Val() string {
	return c.curToken.Val()	
}

func (c *CompilationEngine) writeLeftLabel(label string) {
	c.writeValue(fmt.Sprintf("<%s>\n", label))
}

func (c *CompilationEngine) writeRightLabel(label string) {
	c.writeValue(fmt.Sprintf("</%s>\n", label))
}

func (c *CompilationEngine) writeLabelValue(label string, value string) {
	c.writeValue(fmt.Sprintf("<%s>", label))
	c.writeValue(value)
	c.writeValue(fmt.Sprintf("</%s>\n", label))
}

func (c *CompilationEngine) writeValue(value string) {
	if c.inTx {
		c.txOutput.WriteString(value)
	} else {
		c.output.WriteString(value)
	}
}


