package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CodeWriter struct {
	bufWriter                *bufio.Writer
	jmpFlagCounter           int64
	callReturnAddressCounter int64
	filename                 string
	curFuncName              string
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
	return codeWriter
}

func (w *CodeWriter) SetFileName(filename string) {
	w.bufWriter.Flush()
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	w.bufWriter = bufio.NewWriter(file)
	w.filename = filepath.Base(filename)
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

func (w *CodeWriter) WriteInit() {
	// SP=256
	w.writeLine("@256")
	w.writeLine("D=A")
	w.writeLine("@SP")
	w.writeLine("M=D")

	w.WriteCall("Sys.init", 0)
}

func (w *CodeWriter) WriteRawFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	_, err = io.Copy(w.bufWriter, file)
	if err != nil {
		panic(err)
	}
}

func (w *CodeWriter) WriteLabel(label string) {
	w.writeLine(fmt.Sprintf("(%s)", w.getCurFuncLabel(label)))
}

func (w *CodeWriter) WriteGoto(label string) {
	w.writeJmp(w.getCurFuncLabel(label))
}

func (w *CodeWriter) WriteIf(label string) {
	w.writeLine(popD())
	w.writeLine("@" + w.getCurFuncLabel(label))
	w.writeLine("D;JNE")
}

func (w *CodeWriter) WriteCall(funcName string, n int32) {
	returnAddressLabel := w.getCurFuncName() + ".return." + w.getCallReturnAddressCount()
	w.writeLine(pushValue(returnAddressLabel))
	w.writeLine(pushRegSegment("LCL"))
	w.writeLine(pushRegSegment("ARG"))
	w.writeLine(pushRegSegment("THIS"))
	w.writeLine(pushRegSegment("THAT"))

	// ARG = SP - n - 5
	w.writeLine("@SP")
	w.writeLine("D=M")
	w.writeLine(fmt.Sprintf("@%d", n))
	w.writeLine("D=D-A")
	w.writeLine("@5")
	w.writeLine("D=D-A")
	w.writeLine("@ARG")
	w.writeLine("M=D")

	// LCL = SP
	w.writeLine("@SP")
	w.writeLine("D=M")
	w.writeLine("@LCL")
	w.writeLine("M=D")

	// goto funcName
	w.writeJmp(funcName)

	// return address label
	w.writeLine(fmt.Sprintf("(%s)", returnAddressLabel))
}

func (w *CodeWriter) WriteReturn() {
	// FRAME = LCL
	w.writeLine("@LCL")
	w.writeLine("D=M")
	w.writeLine("@13")
	w.writeLine("M=D")

	// RET = *(FRAME-5)
	w.writeLine("D=M")
	w.writeLine("@5")
	w.writeLine("D=D-A")
	w.writeLine("A=D")
	w.writeLine("D=M")
	w.writeLine("@14")
	w.writeLine("M=D")

	// *ARG = pop()
	w.writeLine(popToMemSegment("ARG", 0))

	// SP = ARG + 1
	w.writeLine("@ARG")
	w.writeLine("D=M+1")
	w.writeLine("@SP")
	w.writeLine("M=D")

	w.writeSourceSubToTarget("13", "THAT", 1)
	w.writeSourceSubToTarget("13", "THIS", 2)
	w.writeSourceSubToTarget("13", "ARG", 3)
	w.writeSourceSubToTarget("13", "LCL", 4)

	// goto RET
	w.writeLine("@14")
	w.writeLine("A=M")
	w.writeLine("0;JMP")

}

func (w *CodeWriter) WriteFunction(funcName string, k int64) {
	w.writeLine(fmt.Sprintf("(%s)", funcName))
	for i := int64(0); i < k; i++ {
		w.writeLine(pushValue("0"))
	}
	w.enterFunc(funcName)
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
	if segment == "constant" {
		w.writeLine(fmt.Sprintf("@%d", index))
		w.writeLine("D=A")
		w.writeLine(pushD())
	} else if segment == "argument" {
		w.writeLine(pushMemSegment("ARG", index))
	} else if segment == "local" {
		w.writeLine(pushMemSegment("LCL", index))
	} else if segment == "this" {
		w.writeLine(pushMemSegment("THIS", index))
	} else if segment == "that" {
		w.writeLine(pushMemSegment("THAT", index))
	} else if segment == "temp" {
		w.writeLine(pushRegSegment(fmt.Sprintf("%d", 5+index)))
	} else if segment == "pointer" {
		if index == 0 {
			w.writeLine(pushRegSegment("3"))
		} else if index == 1 {
			w.writeLine(pushRegSegment("4"))
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
		w.writeLine(popToMemSegment("ARG", index))
	} else if segment == "local" {
		w.writeLine(popToMemSegment("LCL", index))
	} else if segment == "this" {
		w.writeLine(popToMemSegment("THIS", index))
	} else if segment == "that" {
		w.writeLine(popToMemSegment("THAT", index))
	} else if segment == "temp" {
		w.writeLine(popToRegSegment(fmt.Sprintf("%d", 5+index)))
	} else if segment == "pointer" {
		if index == 0 {
			w.writeLine(popToRegSegment("3"))
		} else if index == 1 {
			w.writeLine(popToRegSegment("4"))
		}
	} else if segment == "static" {
		w.writeLine(strings.Join([]string{
			popD(),
			"@" + w.getStaticName(index),
			"M=D",
		}, "\n"))
	}
}

func (w *CodeWriter) writeJmp(label string) {
	w.writeLine("@" + label)
	w.writeLine("0;JMP")
}

func (w *CodeWriter) writeSourceSubToTarget(sourceAddress, targetAddress string, subVal int32) {
	w.writeLine("@" + sourceAddress)
	w.writeLine("D=M")
	w.writeLine(fmt.Sprintf("@%d", subVal))
	w.writeLine("D=D-A")
	w.writeLine("A=D")
	w.writeLine("D=M")
	w.writeLine("@" + targetAddress)
	w.writeLine("M=D")
}

func (w *CodeWriter) getJumpFlagCount() string {
	v := w.jmpFlagCounter
	w.jmpFlagCounter += 1
	return fmt.Sprintf("%d", v)
}

func (w *CodeWriter) getCallReturnAddressCount() string {
	v := w.callReturnAddressCounter
	w.callReturnAddressCounter += 1
	return fmt.Sprintf("%d", v)
}

func (w *CodeWriter) getStaticName(index int64) string {
	return fmt.Sprintf("%s.%d", w.filename, index)
}

func (w *CodeWriter) enterFunc(funcName string) {
	w.curFuncName = funcName
}

func (w *CodeWriter) getCurFuncName() string {
	return w.curFuncName
}

func (w *CodeWriter) getCurFuncLabel(label string) string {
	return w.getCurFuncName() + "." + label
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

func popToRegSegment(reg string) string {
	return strings.Join([]string{
		popD(),
		fmt.Sprintf("@%s", reg),
		"M=D",
	}, "\n")
}

func pushMemSegment(seg string, index int64) string {
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

func pushRegSegment(reg string) string {
	return strings.Join([]string{
		fmt.Sprintf("@%s", reg),
		"D=M",
		pushD(),
	}, "\n")
}

func pushValue(val string) string {
	return strings.Join([]string{
		fmt.Sprintf("@%s", val),
		"D=A",
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
