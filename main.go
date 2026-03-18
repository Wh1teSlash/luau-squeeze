package main

import (
	"fmt"

	"github.com/Wh1teSlash/luau-parser/ast"
	"github.com/Wh1teSlash/luau-parser/lexer"
	"github.com/Wh1teSlash/luau-parser/parser"
	"github.com/Wh1teSlash/luau-squeeze/beautifier"
	"github.com/Wh1teSlash/luau-squeeze/minifier"
)

func main() {
	input := `
	print("PROMETHEUS Benchmark")
print("Based On IronBrew Benchmark")
local Iterations = 100000
print("Iterations: " .. tostring(Iterations))

print("CLOSURE testing.")
local Start = os.clock()
local TStart = Start
for Idx = 1, Iterations do
    (function()
        if not true then
            print("Hey gamer.")
        end
    end)()
end
print("Time:", os.clock() - Start .. "s")

print("SETTABLE testing.")
Start = os.clock()
local T = {}
for Idx = 1, Iterations do
    T[tostring(Idx)] = "EPIC GAMER " .. tostring(Idx)
end

print("Time:", os.clock() - Start .. "s")

print("GETTABLE testing.")
Start = os.clock()
for Idx = 1, Iterations do
    T[1] = T[tostring(Idx)]
end

print("Time:", os.clock() - Start .. "s")
print("Total Time:", os.clock() - TStart .. "s")
	`
	lex := lexer.New(input)
	factory := ast.NewFactory()
	parser := parser.New(lex, factory)
	program := parser.ParseProgram()

	output := minifier.Minify(program)
	fmt.Println(output)

	beautified := beautifier.Beautify(program)
	fmt.Println(beautified)
}
