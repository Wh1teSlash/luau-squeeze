package minifier

import "github.com/Wh1teSlash/luau-parser/ast"

type Pass interface {
	Run(program *ast.Program)
}

type Pipeline struct {
	passes []Pass
}

func NewPipeline() *Pipeline {
	return &Pipeline{}
}

func (p *Pipeline) Add(pass Pass) *Pipeline {
	p.passes = append(p.passes, pass)
	return p
}

func (p *Pipeline) Minify(program *ast.Program) string {
	for _, pass := range p.passes {
		pass.Run(program)
	}
	printer := newMinifyPrinter()
	return printer.print(program)
}
