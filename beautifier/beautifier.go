package beautifier

import "github.com/Wh1teSlash/luau-parser/ast"

type Config struct {
	Indent                    IndentStrategy
	SpacesAroundOperators     bool
	BlankLineBetweenFunctions bool
	Semicolons                bool
	TrailingNewline           bool
}

var DefaultConfig = Config{
	Indent:                    &TabIndent{},
	SpacesAroundOperators:     true,
	BlankLineBetweenFunctions: true,
	Semicolons:                false,
	TrailingNewline:           true,
}

type Beautifier struct {
	config Config
}

func New(config Config) *Beautifier {
	if config.Indent == nil {
		config.Indent = &TabIndent{}
	}
	return &Beautifier{config: config}
}

func (b *Beautifier) Print(program *ast.Program) string {
	p := newBeautifyPrinter(b.config)
	return p.print(program)
}

func Beautify(program *ast.Program) string {
	return New(DefaultConfig).Print(program)
}
