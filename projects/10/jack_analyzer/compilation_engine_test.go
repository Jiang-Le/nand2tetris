package main

import (
	"flag"
	"io"
	"log"
	"os"
	"runtime"
	"testing"
)

var flagVerbose = flag.Bool("v", false, "show all debug log")
var flagInputFile = flag.String("if", "", "input file")
var flagOutputFile = flag.String("of", "", "output file")

func TestCompilationEngine(t *testing.T) {
	flag.Parse()
	if *flagInputFile == "" || *flagOutputFile == "" {
		t.Fatalf("expect input file and output file")
	}

	if !*flagVerbose {
		sysType := runtime.GOOS
		var outputWriter io.Writer
		var err error
		if sysType == "windows" {
			outputWriter, err = os.Open("NUL")
		} else {
			outputWriter, err = os.Open("/dev/null")
		}
		if err != nil {
			t.Fatalf("open null device err: %v", err)
		}
		log.SetOutput(outputWriter)
	}
	
	inputFile := *flagInputFile
	outputFile := *flagOutputFile
	inf, err := os.Open(inputFile)
	if err != nil {
		t.Fatalf("open input file err: %v", err)
	}
	defer inf.Close()
	of, err := os.OpenFile(outputFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		t.Fatalf("open output file err: %v", err)
	}
	ce := NewCompilationEngine(inf, of)
	ce.CompileClass()
	if err := ce.Error(); err != nil {
		t.Fatalf("compile err: %v", err)
	}
}