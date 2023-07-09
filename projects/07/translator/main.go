package main

import (
	"flag"
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

	allOutputPaths := make([]string, 0, len(allVMFile))

	var codeWriter *CodeWriter
	for i, filePath := range allVMFile {
		outputPath := filePath[:len(filePath)-len("vm")] + "tmp"
		allOutputPaths = append(allOutputPaths, outputPath)
		if i == 0 {
			codeWriter = NewCodeWriter(outputPath)
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
			switch parser.CommandType() {
			case C_ARITHMETIC:
				codeWriter.WriteArithmetic(parser.Arg1())
			case C_PUSH, C_POP:
				codeWriter.WritePushPop(parser.CommandType(), parser.Arg1(), parser.Arg2())
			case C_LABEL:
				codeWriter.WriteLabel(parser.Arg1())
			case C_GOTO:
				codeWriter.WriteGoto(parser.Arg1())
			case C_IF:
				codeWriter.WriteIf(parser.Arg1())
			case C_FUNCTION:
				codeWriter.WriteFunction(parser.Arg1(), parser.Arg2())
			case C_RETURN:
				codeWriter.WriteReturn()
			case C_CALL:
				codeWriter.WriteCall(parser.Arg1(), int32(parser.Arg2()))
			}
		}
	}
	if codeWriter != nil {
		codeWriter.Close()
	}

	if fileStat.IsDir() {
		dirName := filepath.Base(path)
		totalOutputPath := filepath.Join(path, dirName+".asm")
		totalOutputWriter := NewCodeWriter(totalOutputPath)
		totalOutputWriter.WriteInit()

		for _, outputPath := range allOutputPaths {
			totalOutputWriter.WriteRawFile(outputPath)
		}
		totalOutputWriter.Close()
	}

}
