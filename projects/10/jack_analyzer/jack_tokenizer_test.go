package main

import (
	"bufio"
	"fmt"
	"html"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TokenizerAllFile(dir, output string) error {
	if err := os.MkdirAll(output, 0777); err != nil {
		return err
	}
	return filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if filepath.Ext(path) != ".jack" {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, filename := filepath.Split(path)
		outputFileName := strings.TrimSuffix(filename, filepath.Ext(filename)) + "T.xml"
		of, err := os.OpenFile(filepath.Join(output, outputFileName), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		if err != nil {
			return err
		}
		defer of.Close()
		tokenizer := NewTokenizer(f)
		FormatTokenizer(tokenizer, of)
		return nil
	})
}

func TestArrayTest(t *testing.T) {
	dir := "/Users/jiel/project/nand2tetris/projects/10/ArrayTest/"
	output := "/Users/jiel/project/nand2tetris/projects/10/ArrayTest/output"
	if err := TokenizerAllFile(dir, output); err != nil {
		t.Fatal(err)
	}
}

func TestExpressionLessSquare(t *testing.T) {
	dir := "/Users/jiel/project/nand2tetris/projects/10/ExpressionLessSquare"
	output := "/Users/jiel/project/nand2tetris/projects/10/ExpressionLessSquare/output"
	if err := TokenizerAllFile(dir, output); err != nil {
		t.Fatal(err)
	}
}

func TestSquare(t *testing.T) {
	dir := "/Users/jiel/project/nand2tetris/projects/10/Square"
	output := "/Users/jiel/project/nand2tetris/projects/10/Square/output"
	if err := TokenizerAllFile(dir, output); err != nil {
		t.Fatal(err)
	}
}

func FormatTokenizer(tokenizer Tokenizer, writer io.Writer) {
	buf := bufio.NewWriter(writer)
	buf.WriteString("<tokens>\n")

	for tokenizer.HasMoreTokens() {
		err := tokenizer.Advance()
		if err != nil {
			panic(err)
		}
		switch tokenizer.TokenType() {
		case KEYWORD:
			writeLabel(buf, "keyword", string(tokenizer.Keyword()))
		case SYMBOL:
			writeLabel(buf, "symbol", string(tokenizer.Symnol()))
		case IDENTIFIER:
			writeLabel(buf, "identifier", string(tokenizer.Identifier()))
		case INT_CONST:
			writeLabel(buf, "integerConstant", strconv.FormatInt(tokenizer.IntVal(), 10))
		case STRING_CONST:
			writeLabel(buf, "stringConstant", string(tokenizer.StringVal()))
		case ERR_IDENTIFIER:
			fmt.Printf("err token: %s\n", tokenizer.StringVal())
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
