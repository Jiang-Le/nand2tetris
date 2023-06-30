package main

import (
	"bufio"
	"errors"
	"io"
	"regexp"
	"strconv"
	"strings"
)

type Parser struct {
	reader     *bufio.Reader
	curLine    string
	curCommand Command
}

func NewParser(reader io.Reader) *Parser {
	return &Parser{
		reader: bufio.NewReader(reader),
	}
}

const (
	C_COMMENT    = 0
	C_ARITHMETIC = iota
	C_PUSH
	C_POP
	C_LABEL
	C_GOTO
	C_IF
	C_FUNCTION
	C_RETURN
	C_CALL
)

type Command struct {
	commandType int
	Arg1        string
	Arg2        int64
}

func (p *Parser) HasMoreCommands() bool {
	line, _, err := p.reader.ReadLine()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return false
		}
		panic(err)
	}
	p.curLine = string(line)
	return true
}

func (p *Parser) Advance() {
	line := strings.TrimSpace(p.curLine)
	var cmd Command
	if len(line) == 0 {
		cmd = Command{
			commandType: C_COMMENT,
		}
	} else {
		re := regexp.MustCompile(`\s+`)
		tokens := re.Split(line, -1)
		switch tokens[0] {
		case "add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not":
			cmd = parseArithmeticCommand(tokens)
		case "push":
			cmd = parsePushCommand(tokens)
		case "pop":
			cmd = parsePopCommand(tokens)
		}
	}
	p.curCommand = cmd
}

func (p *Parser) CommandType() int {
	return p.curCommand.commandType
}

func (p *Parser) Arg1() string {
	return p.curCommand.Arg1
}

func (p *Parser) Arg2() int64 {
	return p.curCommand.Arg2
}

func parseArithmeticCommand(tokens []string) Command {
	if len(tokens) != 1 {

	}
	return Command{
		commandType: C_ARITHMETIC,
		Arg1:        tokens[0],
	}
}

func parsePopCommand(tokens []string) Command {
	arg2, err := strconv.ParseInt(tokens[2], 10, 64)
	if err != nil {
		panic(err)
	}
	return Command{
		commandType: C_POP,
		Arg1:        tokens[1],
		Arg2:        arg2,
	}
}

func parsePushCommand(tokens []string) Command {
	arg2, err := strconv.ParseInt(tokens[2], 10, 64)
	if err != nil {
		panic(err)
	}
	return Command{
		commandType: C_PUSH,
		Arg1:        tokens[1],
		Arg2:        arg2,
	}
}
