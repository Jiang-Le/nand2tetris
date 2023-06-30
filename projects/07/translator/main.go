package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var flagPath = flag.String("input", "", "")

func main() {
	flag.Parse()

	path := *flagPath

	fileStat, err := os.Stat(path)
	if err != nil {
		panic(err)
	}

	allVMFile := []string{}
	if fileStat.IsDir() {
		filepath.Walk(path, func(path string, info fs.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(path, ".vm") {
				allVMFile = append(allVMFile, path)
			}
			return nil
		})
	} else {
		allVMFile = append(allVMFile, path)
	}

	var codeWriter *CodeWriter
	for i, filePath := range allVMFile {
		outputPath := filePath[:len(filePath)-len("vm")] + "asm"
		if i == 0 {
			codeWriter = NewCodeWriter(outputPath)
			defer codeWriter.Close()
		} else {
			codeWriter.SetFileName(outputPath)
		}
		file, err := os.Open(filePath)
		if err != nil {
			panic(err)
		}
		parser := NewParser(file)
		for parser.HasMoreCommands() {
			parser.Advance()
			fmt.Println(parser.CommandType())
			switch parser.CommandType() {
			case C_ARITHMETIC:
				codeWriter.WriteArithmetic(parser.Arg1())
			case C_PUSH, C_POP:
				codeWriter.WritePushPop(parser.CommandType(), parser.Arg1(), parser.Arg2())
			}
		}
	}
}
