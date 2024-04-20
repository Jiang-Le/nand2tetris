package main

import (
	"fmt"
	"os"
)

func main() {
	f, err := os.Open("/Users/jiel/project/nand2tetris/projects/10/ArrayTest/Main.jack")
	if err != nil {
		fmt.Println(err)
		return
	}
	tokenizer := NewTokenizer(f)

	for tokenizer.HasMoreTokens() {
		tokenizer.Advance()
		token := tokenizer.Token()
		switch token.TokenType() {
		case KEYWORD:
			fmt.Println(token.Keyword())
		case SYMBOL:
			fmt.Println(token.Symbol())
		case IDENTIFIER:
			fmt.Println(token.Identifier())
		case INT_CONST:
			fmt.Println(token.IntVal())
		case STRING_CONST:
			fmt.Println(token.StringVal())
		case ERR_IDENTIFIER:
			fmt.Printf("err token: %s\n", token.StringVal())
		}
	}
}
