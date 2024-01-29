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
		switch tokenizer.TokenType() {
		case KEYWORD:
			fmt.Println(tokenizer.Keyword())
		case SYMBOL:
			fmt.Println(tokenizer.Symnol())
		case IDENTIFIER:
			fmt.Println(tokenizer.Identifier())
		case INT_CONST:
			fmt.Println(tokenizer.IntVal())
		case STRING_CONST:
			fmt.Println(tokenizer.StringVal())
		case ERR_IDENTIFIER:
			fmt.Printf("err token: %s\n", tokenizer.StringVal())
		}
	}
}
