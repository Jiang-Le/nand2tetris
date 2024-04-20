package main

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"os"
	"strconv"
	"testing"
	"flag"
)

func TestTokenizer(t *testing.T) {
	flag.Parse()
	argList := flag.Args()
	if len(argList) != 2 {
		t.Fatalf("expect 2 args, got %d args", len(argList))
	}
	inputFile := argList[0]
	outputFile := argList[1]
	if err := TokenizerFile(inputFile, outputFile); err != nil {
		t.Fatal(err)
	}
}

func TokenizerFile(inputFile, outputFile string) error {
	f, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer f.Close()
	of, err := os.OpenFile(outputFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer of.Close()
	tokenizer := NewTokenizer(f)
	FormatTokenizer(tokenizer, of)
	return nil
}

func FormatTokenizer(tokenizer Tokenizer, writer io.Writer) {
	buf := bufio.NewWriter(writer)
	buf.WriteString("<tokens>\n")

	for tokenizer.HasMoreTokens() {
		err := tokenizer.Advance()
		if err != nil {
			panic(err)
		}
		token := tokenizer.Token()
		switch token.TokenType() {
		case KEYWORD:
			writeLabel(buf, "keyword", string(token.Keyword()))
		case SYMBOL:
			writeLabel(buf, "symbol", string(token.Symbol()))
		case IDENTIFIER:
			writeLabel(buf, "identifier", string(token.Identifier()))
		case INT_CONST:
			writeLabel(buf, "integerConstant", strconv.FormatInt(token.IntVal(), 10))
		case STRING_CONST:
			writeLabel(buf, "stringConstant", string(token.StringVal()))
		case ERR_IDENTIFIER:
			fmt.Printf("err token: %s\n", token.StringVal())
		}
	}

	buf.WriteString("</tokens>")
	buf.Flush()
}

func writeLabel(writer *bufio.Writer, label string, value string) {
	writer.WriteString(fmt.Sprintf("<%s>", label))
	writer.WriteString(fmt.Sprintf(" %s ", html.EscapeString(value)))
	writer.WriteString(fmt.Sprintf("</%s>\n", label))
}
