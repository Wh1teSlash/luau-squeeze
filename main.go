package main

import (
	"os"

	"github.com/Wh1teSlash/luau-parser/ast"
	"github.com/Wh1teSlash/luau-parser/lexer"
	"github.com/Wh1teSlash/luau-parser/parser"
	"github.com/Wh1teSlash/luau-parser/visitors"
	"github.com/Wh1teSlash/luau-squeeze/beautifier"
)

func main() {
	data, err := os.ReadFile("script.lua")
	if err != nil {
		panic(err)
	}
	input := string(data)

	lex := lexer.New(input)
	factory := ast.NewFactory()
	parser := parser.New(lex, factory)
	program := parser.ParseProgram()
	printer := visitors.NewTreePrinter()

	astFile, err := os.Create("ast.lua")
	if err != nil {
		panic(err)
	}
	_, err = astFile.WriteString(printer.Print(program))
	if err != nil {
		panic(err)
	}

	beautified := beautifier.Beautify(program)

	file, err := os.Create("beautified.lua")
	if err != nil {
		panic(err)
	}
	_, err = file.WriteString(beautified)
	if err != nil {
		panic(err)
	}
}
