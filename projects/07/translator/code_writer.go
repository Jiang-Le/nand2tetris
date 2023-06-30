package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type CodeWriter struct {
	bufWriter      *bufio.Writer
	jmpFlagCounter int64
	filename       string
}

func NewCodeWriter(path string) *CodeWriter {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	_, filename := filepath.Split(path)
	codeWriter := &CodeWriter{
		bufWriter: bufio.NewWriter(file),
		filename:  filepath.Base(filename),
	}
	// codeWriter.init()
	return codeWriter
}

func (w *CodeWriter) init() {
	w.writeLine(strings.Join([]string{
		"@256", // 初始化栈顶
		"D=A",
		"@SP",
		"M=D",
	}, "\n"))
}

func (w *CodeWriter) SetFileName(filename string) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	w.bufWriter = bufio.NewWriter(file)
}

func (w *CodeWriter) WriteArithmetic(command string) {
	switch command {
	case "add":
		w.writeAdd()
	case "sub":
		w.writeSub()
	case "neg":
		w.writeNeg()
	case "eq":
		w.writeEq()
	case "gt":
		w.writeGt()
	case "lt":
		w.writeLt()
	case "and":
		w.writeAnd()
	case "or":
		w.writeOr()
	case "not":
		w.writeNot()
	}
}

func (w *CodeWriter) Close() {
	w.bufWriter.Flush()
}

func (w *CodeWriter) WritePushPop(commandType int, segment string, index int64) {
	switch commandType {
	case C_PUSH:
		w.writePush(segment, index)
	case C_POP:
		w.writePop(segment, index)
	}
}

func (w *CodeWriter) writeAdd() {
	w.writeLine(popM())
	w.writeLine("D=M")
	w.writeLine(popM())
	w.writeLine("M=D+M")
	w.writeLine(spAdd1())
}

func (w *CodeWriter) writeSub() {
	w.writeLine(popM())
	w.writeLine("D=-M")
	w.writeLine(popM())
	w.writeLine("M=D+M")
	w.writeLine(spAdd1())
}

func (w *CodeWriter) writeNeg() {
	w.writeLine(popM())
	w.writeLine("M=-M")
	w.writeLine(spAdd1())
}

func (w *CodeWriter) writeEq() {
	v := w.getJumpFlagCount()
	w.writeLine(popM())
	w.writeLine("D=-M")
	w.writeLine(popM())
	w.writeLine(strings.Join([]string{
		"D=D+M",
		"@writeTrue." + v,
		"D;JEQ",
		"D=0",
		"@writeFalse." + v,
		"0;JMP",
		"(" + "writeTrue." + v + ")",
		"D=-1",
		"(" + "writeFalse." + v + ")"},
		"\n"))
	w.writeLine(pushD())
}

func (w *CodeWriter) writeGt() {
	v := w.getJumpFlagCount()
	w.writeLine(popM())
	w.writeLine("D=-M")
	w.writeLine(popM())
	w.writeLine(strings.Join([]string{
		"D=D+M",
		"@writeTrue." + v,
		"D;JGT",
		"D=0",
		"@writeFalse." + v,
		"0;JMP",
		"(" + "writeTrue." + v + ")",
		"D=-1",
		"(" + "writeFalse." + v + ")"},
		"\n"))
	w.writeLine(pushD())
}

func (w *CodeWriter) writeLt() {
	v := w.getJumpFlagCount()
	w.writeLine(popM())
	w.writeLine("D=-M")
	w.writeLine(popM())
	w.writeLine(strings.Join([]string{
		"D=D+M",
		"@writeTrue." + v,
		"D;JLT",
		"D=0",
		"@writeFalse." + v,
		"0;JMP",
		"(" + "writeTrue." + v + ")",
		"D=-1",
		"(" + "writeFalse." + v + ")"},
		"\n"))
	w.writeLine(pushD())
}

func (w *CodeWriter) writeAnd() {
	w.writeLine(popM())
	w.writeLine("D=M")
	w.writeLine(popM())
	w.writeLine("D=D&M")
	w.writeLine(pushD())
}

func (w *CodeWriter) writeOr() {
	w.writeLine(popM())
	w.writeLine("D=M")
	w.writeLine(popM())
	w.writeLine("D=D|M")
	w.writeLine(pushD())
}

func (w *CodeWriter) writeNot() {
	w.writeLine(popM())
	w.writeLine("D=!M")
	w.writeLine(pushD())
}

func (w *CodeWriter) writeLine(line string) {
	w.bufWriter.WriteString(line)
	w.bufWriter.WriteString("\n")
}

func (w *CodeWriter) writePush(segment string, index int64) {
	fmt.Println("writePush")
	if segment == "constant" {
		w.writeLine(fmt.Sprintf("@%d", index))
		w.writeLine("D=A")
		w.writeLine(pushD())
	} else if segment == "argument" {
		fmt.Printf("argument: %d\n", index)
		fmt.Println(pushToMemSegment("ARG", index))
		w.writeLine(pushToMemSegment("ARG", index))
	} else if segment == "local" {
		w.writeLine(pushToMemSegment("LCL", index))
	} else if segment == "this" {
		w.writeLine(pushToMemSegment("THIS", index))
	} else if segment == "that" {
		w.writeLine(pushToMemSegment("THAT", index))
	} else if segment == "temp" {
		w.writeLine(pushToRegSegment(5 + index))
	} else if segment == "pointer" {
		if index == 0 {
			w.writeLine(pushToRegSegment(3))
		} else if index == 1 {
			w.writeLine(pushToRegSegment(4))
		}
	} else if segment == "static" {
		w.writeLine(strings.Join([]string{
			"@" + w.getStaticName(index),
			"D=M",
			pushD(),
		}, "\n"))
	}
}

func (w *CodeWriter) writePop(segment string, index int64) {
	if segment == "constant" {
		w.writeLine(popM())
	} else if segment == "argument" {
		fmt.Println(popToMemSegment("ARG", index))
		w.writeLine(popToMemSegment("ARG", index))
	} else if segment == "local" {
		w.writeLine(popToMemSegment("LCL", index))
	} else if segment == "this" {
		w.writeLine(popToMemSegment("THIS", index))
	} else if segment == "that" {
		w.writeLine(popToMemSegment("THAT", index))
	} else if segment == "temp" {
		w.writeLine(popToRegSegment(5 + index))
	} else if segment == "pointer" {
		if index == 0 {
			w.writeLine(popToRegSegment(3))
		} else if index == 1 {
			w.writeLine(popToRegSegment(4))
		}
	} else if segment == "static" {
		w.writeLine(strings.Join([]string{
			popD(),
			"@" + w.getStaticName(index),
			"M=D",
		}, "\n"))
	}
}

func (w *CodeWriter) getJumpFlagCount() string {
	v := w.jmpFlagCounter
	w.jmpFlagCounter += 1
	return fmt.Sprintf("%d", v)
}

func (w *CodeWriter) getStaticName(index int64) string {
	return fmt.Sprintf("%s.%d", w.filename, index)
}

func popToMemSegment(seg string, index int64) string {
	return strings.Join([]string{
		"@" + seg,
		"D=M",
		fmt.Sprintf("@%d", index),
		"D=D+A",
		"@15",
		"M=D",
		popD(),
		"@15",
		"A=M",
		"M=D",
	}, "\n")
}

func popToRegSegment(regIndex int64) string {
	return strings.Join([]string{
		popD(),
		fmt.Sprintf("@%d", regIndex),
		"M=D",
	}, "\n")
}

func pushToMemSegment(seg string, index int64) string {
	return strings.Join([]string{
		"@" + seg,
		"D=M",
		fmt.Sprintf("@%d", index),
		"D=D+A",
		"A=D",
		"D=M",
		pushD(),
	}, "\n")
}

func pushToRegSegment(regIndex int64) string {
	return strings.Join([]string{
		fmt.Sprintf("@%d", regIndex),
		"D=M",
		pushD(),
	}, "\n")
}

// popM 将SP减1，同时栈顶元素的值在M中
func popM() string {
	return strings.Join([]string{
		"@SP",
		"M=M-1",
		"A=M",
	}, "\n")

}

// pushD 将SP减1，同时栈顶元素的值放到D中
func popD() string {
	return strings.Join([]string{
		"@SP",
		"M=M-1",
		"A=M",
		"D=M",
	}, "\n")
}

// pushD 将D中的值放入栈顶，同时将SP+1
func pushD() string {
	return strings.Join([]string{
		"@SP",
		"A=M",
		"M=D",
		"@SP",
		"M=M+1",
	}, "\n")
}

// spAdd1 将SP增加1
func spAdd1() string {
	return strings.Join([]string{
		"@SP",
		"M=M+1",
	},
		"\n",
	)
}
