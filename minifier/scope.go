package minifier

import "github.com/Wh1teSlash/luau-parser/ast"

type symbol struct {
	originalName string
	shortName    string
	declSites    []func(string)
	identRefs    []*ast.Identifier
}

type scope struct {
	parent  *scope
	symbols map[string]*symbol
}

func newScope(parent *scope) *scope {
	return &scope{parent: parent, symbols: make(map[string]*symbol)}
}

func (s *scope) declare(name string, sym *symbol) {
	s.symbols[name] = sym
}

func (s *scope) lookup(name string) *symbol {
	if sym, ok := s.symbols[name]; ok {
		return sym
	}
	if s.parent != nil {
		return s.parent.lookup(name)
	}
	return nil
}

type resolver struct {
	current *scope
	locals  []*symbol
}

func newResolver() *resolver {
	return &resolver{current: newScope(nil)}
}

func (r *resolver) push() { r.current = newScope(r.current) }
func (r *resolver) pop()  { r.current = r.current.parent }

func (r *resolver) declare(name string, declSite func(string)) *symbol {
	sym := &symbol{
		originalName: name,
		declSites:    []func(string){declSite},
	}
	r.current.declare(name, sym)
	r.locals = append(r.locals, sym)
	return sym
}

func (r *resolver) ref(ident *ast.Identifier) {
	if sym := r.current.lookup(ident.Name); sym != nil {
		sym.identRefs = append(sym.identRefs, ident)
	}
}

func resolve(program *ast.Program) []*symbol {
	r := newResolver()
	r.walkStmts(program.Body)
	return r.locals
}

func (r *resolver) walkStmts(stmts []ast.Stmt) {
	for _, s := range stmts {
		r.walkStmt(s)
	}
}

func (r *resolver) walkStmt(stmt ast.Stmt) {
	switch s := stmt.(type) {
	case *ast.LocalAssignment:
		for _, val := range s.Values {
			r.walkExpr(val)
		}
		for i := range s.Names {
			idx := i
			r.declare(s.Names[i], func(name string) { s.Names[idx] = name })
		}

	case *ast.LocalFunction:
		r.declare(s.Name, func(name string) { s.Name = name })
		r.push()
		for i := range s.Parameters {
			p := s.Parameters[i]
			r.declare(p.Name, func(name string) { p.Name = name })
		}
		r.walkStmts(s.Body.Statements)
		r.pop()

	case *ast.FunctionDef:
		r.push()
		for i := range s.Parameters {
			p := s.Parameters[i]
			r.declare(p.Name, func(name string) { p.Name = name })
		}
		r.walkStmts(s.Body.Statements)
		r.pop()

	case *ast.Assignment:
		for _, t := range s.Targets {
			r.walkExpr(t)
		}
		for _, v := range s.Values {
			r.walkExpr(v)
		}

	case *ast.IfStatement:
		r.walkExpr(s.Condition)
		r.push()
		r.walkStmts(s.Then.Statements)
		r.pop()
		for _, elif := range s.ElseIfs {
			r.walkExpr(elif.Condition)
			r.push()
			r.walkStmts(elif.Body.Statements)
			r.pop()
		}
		if s.Else != nil {
			r.push()
			r.walkStmts(s.Else.Statements)
			r.pop()
		}

	case *ast.WhileLoop:
		r.walkExpr(s.Condition)
		r.push()
		r.walkStmts(s.Body.Statements)
		r.pop()

	case *ast.RepeatLoop:
		r.push()
		r.walkStmts(s.Body.Statements)
		r.walkExpr(s.Condition)
		r.pop()

	case *ast.ForLoop:
		r.walkExpr(s.Start)
		r.walkExpr(s.End)
		if s.Step != nil {
			r.walkExpr(s.Step)
		}
		r.push()
		r.declare(s.Variable, func(name string) { s.Variable = name })
		r.walkStmts(s.Body.Statements)
		r.pop()

	case *ast.ForInLoop:
		for _, iter := range s.Iterables {
			r.walkExpr(iter)
		}
		r.push()
		for i := range s.Variables {
			idx := i
			r.declare(s.Variables[i], func(name string) { s.Variables[idx] = name })
		}
		r.walkStmts(s.Body.Statements)
		r.pop()

	case *ast.DoBlock:
		r.push()
		r.walkStmts(s.Body.Statements)
		r.pop()

	case *ast.ReturnStatement:
		for _, v := range s.Values {
			r.walkExpr(v)
		}

	case *ast.ExpressionStatement:
		r.walkExpr(s.Expr)

	case *ast.MetamethodDef:
		r.push()
		for i := range s.Parameters {
			p := s.Parameters[i]
			r.declare(p.Name, func(name string) { p.Name = name })
		}
		r.walkStmts(s.Body.Statements)
		r.pop()
	}
}

func (r *resolver) walkExpr(expr ast.Expr) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *ast.Identifier:
		r.ref(e)
	case *ast.BinaryOp:
		r.walkExpr(e.Left)
		r.walkExpr(e.Right)
	case *ast.UnaryOp:
		r.walkExpr(e.Operand)
	case *ast.FunctionCall:
		r.walkExpr(e.Function)
		for _, arg := range e.Args {
			r.walkExpr(arg)
		}
	case *ast.MethodCall:
		r.walkExpr(e.Object)
		for _, arg := range e.Args {
			r.walkExpr(arg)
		}
	case *ast.IndexAccess:
		r.walkExpr(e.Table)
		r.walkExpr(e.Index)
	case *ast.FieldAccess:
		r.walkExpr(e.Object)
	case *ast.TableLiteral:
		for _, field := range e.Fields {
			if field.Key != nil {
				r.walkExpr(field.Key)
			}
			r.walkExpr(field.Value)
		}
	case *ast.FunctionExpr:
		r.push()
		for i := range e.Parameters {
			p := e.Parameters[i]
			r.declare(p.Name, func(name string) { p.Name = name })
		}
		r.walkStmts(e.Body.Statements)
		r.pop()
	case *ast.TypeCast:
		r.walkExpr(e.Value)
	case *ast.IfExpr:
		r.walkExpr(e.Condition)
		r.walkExpr(e.Then)
		for _, elif := range e.ElseIfs {
			r.walkExpr(elif.Condition)
			r.walkExpr(elif.Then)
		}
		r.walkExpr(e.Else)
	case *ast.ParenExpr:
		r.walkExpr(e.Expr)
	case *ast.InterpolatedString:
		for _, ex := range e.Expressions {
			r.walkExpr(ex)
		}
	}
}
