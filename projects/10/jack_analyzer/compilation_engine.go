package main

import "io"

type CompilationEngine struct {
	tokenizer Tokenizer
}

func NewCompilationEngine(reader io.Reader) CompilationEngine {
	return CompilationEngine{
		tokenizer: NewTokenizer(reader),
	}
}

func (c *CompilationEngine) CompileClass() {

}
