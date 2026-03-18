package minifier

import (
	"github.com/Wh1teSlash/luau-parser/ast"
)

type CommentStripPass struct{}

func (p *CommentStripPass) Run(program *ast.Program) {
	program.Body = stripComments(program.Body)
}

func stripComments(stmts []ast.Stmt) []ast.Stmt {
	result := stmts[:0]
	for _, stmt := range stmts {
		if _, isComment := stmt.(*ast.Comment); isComment {
			continue
		}
		stripBlock(stmt)
		result = append(result, stmt)
	}
	return result
}

func stripBlock(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.IfStatement:
		s.Then.Statements = stripComments(s.Then.Statements)
		for _, elif := range s.ElseIfs {
			elif.Body.Statements = stripComments(elif.Body.Statements)
		}
		if s.Else != nil {
			s.Else.Statements = stripComments(s.Else.Statements)
		}
	case *ast.WhileLoop:
		s.Body.Statements = stripComments(s.Body.Statements)
	case *ast.RepeatLoop:
		s.Body.Statements = stripComments(s.Body.Statements)
	case *ast.ForLoop:
		s.Body.Statements = stripComments(s.Body.Statements)
	case *ast.ForInLoop:
		s.Body.Statements = stripComments(s.Body.Statements)
	case *ast.DoBlock:
		s.Body.Statements = stripComments(s.Body.Statements)
	case *ast.FunctionDef:
		s.Body.Statements = stripComments(s.Body.Statements)
	case *ast.LocalFunction:
		s.Body.Statements = stripComments(s.Body.Statements)
	case *ast.MetamethodDef:
		s.Body.Statements = stripComments(s.Body.Statements)
	}
}

type RenamePass struct {
	Strategy RenameStrategy
}

func (p *RenamePass) Run(program *ast.Program) {
	p.Strategy.Reset()
	locals := resolve(program)
	for _, sym := range locals {
		name := p.Strategy.Next(sym.originalName)
		sym.shortName = name
		for _, site := range sym.declSites {
			site(name)
		}
		for _, ident := range sym.identRefs {
			ident.Name = name
		}
	}
}
