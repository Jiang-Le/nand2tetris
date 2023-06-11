package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var source = flag.String("s", "", "source file path")
var target = flag.String("t", "", "output file path")

func main() {
	flag.Parse()
	reader, err := os.Open(*source)
	if err != nil {
		panic(err)
	}
	defer reader.Close()
	writer, err := os.OpenFile(*target, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer writer.Close()
	DoParser(reader, writer)
}

func DoParser(reader io.Reader, writer io.Writer) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		panic(err)
	}
	symbolParser := NewParser(bytes.NewReader(data))
	table := NewSymbolTalbe()
	codeAddress := 0
	for symbolParser.HasMoreCommands() {
		symbolParser.Advance()
		if symbolParser.CommandType() == L_COMMAND {
			table.AddEntry(symbolParser.Symbol(), codeAddress)
		} else if symbolParser.CommandType() == A_COMMAND || symbolParser.CommandType() == C_COMMAND {
			codeAddress += 1
		}
	}

	// fmt.Println("---------------")

	var codes []string
	parser := NewParser(bytes.NewReader(data))
	for parser.HasMoreCommands() {
		parser.Advance()
		// fmt.Printf("comand: %+v\n", parser.curCommand)
		switch parser.CommandType() {
		case L_COMMAND:
			continue
		case A_COMMAND:
			symbol := parser.Symbol()
			var aVal int64
			if isNumber(symbol) {
				aVal, _ = strconv.ParseInt(symbol, 10, 32)
			} else {
				if !table.Contains(symbol) {
					table.AddVariable(symbol)
				}
				aVal = int64(table.GetAddress(symbol))
			}
			fmt.Printf("A: 0%015b\n", aVal)
			codes = append(codes, fmt.Sprintf("0%015b", aVal))
		case C_COMMAND:
			code := "111"
			comp := parser.Comp()
			code += compInstructMap[comp]
			dest := parser.Dest()
			code += destRegMap[dest]
			jump := parser.Jump()
			fmt.Printf("comp: %s, dest: %s, jump: %s\n", comp, dest, jump)
			code += jumpMap[string(jump)]
			fmt.Printf("C: %s\n", code)
			codes = append(codes, code)
		}
	}

	bufWriter := bufio.NewWriter(writer)
	for _, code := range codes {
		bufWriter.WriteString(code + "\n")
	}
	bufWriter.Flush()
}

var jumpMap = map[string]string{
	"null": "000",
	"JGT":  "001",
	"JEQ":  "010",
	"JGE":  "011",
	"JLT":  "100",
	"JNE":  "101",
	"JLE":  "110",
	"JMP":  "111",
}

var destRegMap = map[string]string{
	"null": "000",
	"M":    "001",
	"D":    "010",
	"MD":   "011",
	"A":    "100",
	"AM":   "101",
	"AD":   "110",
	"AMD":  "111",
}

var compInstructMap = map[string]string{
	"0":   "0101010",
	"1":   "0111111",
	"-1":  "0111010",
	"D":   "0001100",
	"A":   "0110000",
	"M":   "1110000",
	"!D":  "0001101",
	"!A":  "0110001",
	"!M":  "1110001",
	"-D":  "0001111",
	"-A":  "0110011",
	"-M":  "1110011",
	"D+1": "0011111",
	"A+1": "0110111",
	"M+1": "1110111",
	"D-1": "0001110",
	"A-1": "0110010",
	"M-1": "1110010",
	"D+A": "0000010",
	"D+M": "1000010",
	"D-A": "0010011",
	"D-M": "1010011",
	"A-D": "0000111",
	"M-D": "1000111",
	"D&A": "0000000",
	"D&M": "1000000",
	"D|A": "0010101",
	"D|M": "1010101",
}

func isNumber(str string) bool {
	for _, c := range str {
		if !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

func NewParser(reader io.Reader) Parser {
	return Parser{
		reader: bufio.NewReader(reader),
	}
}

type Parser struct {
	reader     *bufio.Reader
	eof        bool
	curLine    string
	curCommand Command
}

type CommandType int32

const (
	A_COMMAND  CommandType = 1
	C_COMMAND  CommandType = 2
	L_COMMAND  CommandType = 3
	EMPTY_LINE CommandType = 4
	COMMENT    CommandType = 5
)

type JumpType string

const (
	Null JumpType = "null"
	JGT           = "JGT"
	JEQ           = "JEQ"
	JGE           = "JGE"
	JLT           = "JLT"
	JNE           = "JNE"
	JLE           = "JLE"
	JMP           = "JMP"
)

type Command struct {
	commandType CommandType
	symbol      string
	destType    string
	comp        string
	jump        JumpType
}

func (p *Parser) HasMoreCommands() bool {
	line, _, err := p.reader.ReadLine()
	if err != nil {
		if err == io.EOF {
			return false
		}
		panic(err)
	}
	p.curLine = string(line)
	return true
}

func (p *Parser) Advance() {
	line := p.curLine
	// fmt.Printf("raw: %s\n", line)
	line = strings.TrimSpace(line)
	curCommand := Command{}
	if len(line) == 0 {
		curCommand.commandType = EMPTY_LINE
	} else if line[0] == '@' { // A指令
		curCommand.commandType = A_COMMAND
		curCommand.symbol = string(line[1:])
	} else if line[0] == '(' { // L指令
		curCommand.commandType = L_COMMAND
		length := len(line)
		for i, c := range line[1:] {
			if c == ')' && i != length-2 {
				panic(fmt.Errorf("expect end command, got %s", line[i+1:]))
			}
		}
		curCommand.symbol = string(line[1 : length-1])
	} else if string(line[0:2]) == "//" {
		curCommand.commandType = COMMENT
	} else { // C指令
		curCommand.commandType = C_COMMAND
		commentIndex := strings.Index(line, "//")
		if commentIndex != -1 {
			line = strings.TrimSpace(line[:commentIndex])
		}
		segments := strings.Split(line, ";")
		if len(segments) > 2 {
			panic(fmt.Errorf("invalid grammar for '%s'", line))
		}
		calCommand := segments[0]
		calCommandSegments := strings.Split(calCommand, "=")
		if len(calCommandSegments) > 2 {
			panic(fmt.Errorf("invalid grammar for '%s'", line))
		}
		var jump string
		if len(segments) == 2 {
			jump = segments[1]
		} else {
			jump = "null"
		}

		var destReg string
		var comp string
		if len(calCommandSegments) == 2 {
			destReg = calCommandSegments[0]
			comp = calCommandSegments[1]
		} else {
			destReg = "null"
			comp = calCommandSegments[0]
		}

		curCommand.destType = destReg
		curCommand.comp = comp
		curCommand.jump = JumpType(jump)
	}
	p.curCommand = curCommand
}

func (p *Parser) CommandType() CommandType {
	return p.curCommand.commandType
}

func (p *Parser) Symbol() string {
	return p.curCommand.symbol
}

func (p *Parser) Dest() string {
	return p.curCommand.destType
}

func (p *Parser) Comp() string {
	return p.curCommand.comp
}

func (p *Parser) Jump() JumpType {
	return p.curCommand.jump
}

func NewSymbolTalbe() SymbolTable {
	initTable := map[string]int{
		"SP":     0,
		"LCL":    1,
		"ARG":    2,
		"THIS":   3,
		"THAT":   4,
		"SCREEN": 16384,
		"KBD":    24576,
	}
	for i := 0; i <= 15; i++ {
		initTable[fmt.Sprintf("R%d", i)] = i
	}

	return SymbolTable{
		table:           initTable,
		variableAddress: 16,
	}
}

type SymbolTable struct {
	table           map[string]int
	variableAddress int
}

func (s *SymbolTable) AddEntry(symbol string, address int) {
	s.table[symbol] = address
}

func (s *SymbolTable) AddVariable(symbol string) {
	s.table[symbol] = s.variableAddress
	s.variableAddress += 1
}

func (s *SymbolTable) Contains(symbol string) bool {
	_, ok := s.table[symbol]
	return ok
}

func (s *SymbolTable) GetAddress(symbol string) int {
	address := s.table[symbol]
	return address
}
