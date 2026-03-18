package minifier

import (
	"fmt"
	"strings"

	"github.com/Wh1teSlash/luau-parser/ast"
)

type minifyPrinter struct {
	builder strings.Builder
	last    byte
}

func newMinifyPrinter() *minifyPrinter {
	return &minifyPrinter{}
}

func (p *minifyPrinter) print(program *ast.Program) string {
	p.builder.Reset()
	p.last = 0
	program.Accept(p)
	return p.builder.String()
}

func isAlnum(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_'
}

func (p *minifyPrinter) write(s string) {
	if len(s) == 0 {
		return
	}
	next := s[0]
	if isAlnum(p.last) && isAlnum(next) {
		p.builder.WriteByte(' ')
	}
	if p.last == '-' && next == '-' {
		p.builder.WriteByte(' ')
	}
	p.builder.WriteString(s)
	p.last = s[len(s)-1]
}

func (p *minifyPrinter) printExprList(exprs []ast.Expr) {
	for i, expr := range exprs {
		expr.Accept(p)
		if i < len(exprs)-1 {
			p.write(",")
		}
	}
}

func (p *minifyPrinter) printParams(params []*ast.Parameter) {
	for i, param := range params {
		p.write(param.Name)
		if param.Type != nil {
			p.write(":")
			param.Type.Accept(p)
		}
		if i < len(params)-1 {
			p.write(",")
		}
	}
}

func (p *minifyPrinter) VisitProgram(node *ast.Program) any {
	for i, stmt := range node.Body {
		stmt.Accept(p)
		if i < len(node.Body)-1 {
			p.builder.WriteByte(';')
			p.last = ';'
		}
	}
	return nil
}

func (p *minifyPrinter) VisitBlock(node *ast.Block) any {
	for i, stmt := range node.Statements {
		stmt.Accept(p)
		if i < len(node.Statements)-1 {
			p.builder.WriteByte(';')
			p.last = ';'
		}
	}
	return nil
}

func (p *minifyPrinter) VisitModule(node *ast.Module) any {
	if node.Body != nil {
		node.Body.Accept(p)
	}
	return nil
}

func (p *minifyPrinter) VisitAttribute(node *ast.Attribute) any {
	p.write("@" + node.Name)
	return nil
}

func (p *minifyPrinter) VisitComment(node *ast.Comment) any { return nil }

func (p *minifyPrinter) VisitLocalAssignment(node *ast.LocalAssignment) any {
	p.write("local")
	for i, name := range node.Names {
		if i == 0 {
			p.write(name)
		} else {
			p.write(",")
			p.write(name)
		}
		if i < len(node.Types) && node.Types[i] != nil {
			p.write(":")
			node.Types[i].Accept(p)
		}
	}
	if len(node.Values) > 0 {
		p.write("=")
		p.printExprList(node.Values)
	}
	return nil
}

func (p *minifyPrinter) VisitAssignment(node *ast.Assignment) any {
	p.printExprList(node.Targets)
	p.write(node.Operator)
	p.printExprList(node.Values)
	return nil
}

func (p *minifyPrinter) VisitIfStatement(node *ast.IfStatement) any {
	p.write("if")
	node.Condition.Accept(p)
	p.write("then")
	node.Then.Accept(p)
	for _, elif := range node.ElseIfs {
		p.write("elseif")
		elif.Condition.Accept(p)
		p.write("then")
		elif.Body.Accept(p)
	}
	if node.Else != nil {
		p.write("else")
		node.Else.Accept(p)
	}
	p.write("end")
	return nil
}

func (p *minifyPrinter) VisitWhileLoop(node *ast.WhileLoop) any {
	p.write("while")
	node.Condition.Accept(p)
	p.write("do")
	node.Body.Accept(p)
	p.write("end")
	return nil
}

func (p *minifyPrinter) VisitRepeatLoop(node *ast.RepeatLoop) any {
	p.write("repeat")
	node.Body.Accept(p)
	p.write("until")
	node.Condition.Accept(p)
	return nil
}

func (p *minifyPrinter) VisitForLoop(node *ast.ForLoop) any {
	p.write("for")
	p.write(node.Variable)
	p.write("=")
	node.Start.Accept(p)
	p.write(",")
	node.End.Accept(p)
	if node.Step != nil {
		p.write(",")
		node.Step.Accept(p)
	}
	p.write("do")
	node.Body.Accept(p)
	p.write("end")
	return nil
}

func (p *minifyPrinter) VisitForInLoop(node *ast.ForInLoop) any {
	p.write("for")
	for i, v := range node.Variables {
		if i > 0 {
			p.write(",")
		}
		p.write(v)
	}
	p.write("in")
	p.printExprList(node.Iterables)
	p.write("do")
	node.Body.Accept(p)
	p.write("end")
	return nil
}

func (p *minifyPrinter) VisitDoBlock(node *ast.DoBlock) any {
	p.write("do")
	node.Body.Accept(p)
	p.write("end")
	return nil
}

func (p *minifyPrinter) VisitFunctionDef(node *ast.FunctionDef) any {
	for _, attr := range node.Attributes {
		attr.Accept(p)
	}
	p.write("function")
	p.write(node.Name)
	p.write("(")
	p.printParams(node.Parameters)
	p.write(")")
	if node.ReturnType != nil {
		p.write(":")
		node.ReturnType.Accept(p)
	}
	node.Body.Accept(p)
	p.write("end")
	return nil
}

func (p *minifyPrinter) VisitLocalFunction(node *ast.LocalFunction) any {
	for _, attr := range node.Attributes {
		attr.Accept(p)
	}
	p.write("local")
	p.write("function")
	p.write(node.Name)
	p.write("(")
	p.printParams(node.Parameters)
	p.write(")")
	if node.ReturnType != nil {
		p.write(":")
		node.ReturnType.Accept(p)
	}
	node.Body.Accept(p)
	p.write("end")
	return nil
}

func (p *minifyPrinter) VisitReturnStatement(node *ast.ReturnStatement) any {
	p.write("return")
	p.printExprList(node.Values)
	return nil
}

func (p *minifyPrinter) VisitBreakStatement(node *ast.BreakStatement) any {
	p.write("break")
	return nil
}

func (p *minifyPrinter) VisitContinueStatement(node *ast.ContinueStatement) any {
	p.write("continue")
	return nil
}

func (p *minifyPrinter) VisitTypeAlias(node *ast.TypeAlias) any {
	if node.IsExport {
		p.write("export")
	}
	p.write("type")
	p.write(node.Name)
	if len(node.Generics) > 0 {
		p.write("<")
		p.write(strings.Join(node.Generics, ","))
		p.write(">")
	}
	p.write("=")
	node.Type.Accept(p)
	return nil
}

func (p *minifyPrinter) VisitMetamethodDef(node *ast.MetamethodDef) any {
	p.write("function")
	p.write(node.Name)
	p.write("(")
	p.printParams(node.Parameters)
	p.write(")")
	node.Body.Accept(p)
	p.write("end")
	return nil
}

func (p *minifyPrinter) VisitEmptyStatement(node *ast.EmptyStatement) any { return nil }

func (p *minifyPrinter) VisitExpressionStatement(node *ast.ExpressionStatement) any {
	node.Expr.Accept(p)
	return nil
}

func (p *minifyPrinter) VisitIdentifier(node *ast.Identifier) any {
	p.write(node.Name)
	return nil
}

func (p *minifyPrinter) VisitLiteral(node *ast.Literal) any {
	if node.Type == "string" {
		p.write(fmt.Sprintf("%q", node.Value))
	} else {
		p.write(fmt.Sprintf("%v", node.Value))
	}
	return nil
}

func (p *minifyPrinter) VisitBinaryOp(node *ast.BinaryOp) any {
	node.Left.Accept(p)
	p.write(node.Op)
	node.Right.Accept(p)
	return nil
}

func (p *minifyPrinter) VisitUnaryOp(node *ast.UnaryOp) any {
	p.write(node.Op)
	node.Operand.Accept(p)
	return nil
}

func (p *minifyPrinter) VisitFunctionCall(node *ast.FunctionCall) any {
	node.Function.Accept(p)
	if len(node.Args) == 1 {
		if lit, ok := node.Args[0].(*ast.Literal); ok && lit.Type == "string" {
			str := fmt.Sprintf("%v", lit.Value)
			if !strings.Contains(str, "'") {
				p.write("'" + str + "'")
				return nil
			} else if !strings.Contains(str, "\"") {
				p.write("\"" + str + "\"")
				return nil
			}
		}
	}
	p.write("(")
	p.printExprList(node.Args)
	p.write(")")
	return nil
}

func (p *minifyPrinter) VisitMethodCall(node *ast.MethodCall) any {
	node.Object.Accept(p)
	p.write(":")
	p.write(node.Method)
	if len(node.Args) == 1 {
		if lit, ok := node.Args[0].(*ast.Literal); ok && lit.Type == "string" {
			str := fmt.Sprintf("%v", lit.Value)
			if !strings.Contains(str, "'") {
				p.write("'" + str + "'")
				return nil
			} else if !strings.Contains(str, "\"") {
				p.write("\"" + str + "\"")
				return nil
			}
		}
	}
	p.write("(")
	p.printExprList(node.Args)
	p.write(")")
	return nil
}

func (p *minifyPrinter) VisitIndexAccess(node *ast.IndexAccess) any {
	node.Table.Accept(p)
	p.write("[")
	node.Index.Accept(p)
	p.write("]")
	return nil
}

func (p *minifyPrinter) VisitFieldAccess(node *ast.FieldAccess) any {
	node.Object.Accept(p)
	p.write(".")
	p.write(node.Field)
	return nil
}

func (p *minifyPrinter) VisitTableLiteral(node *ast.TableLiteral) any {
	p.write("{")
	for i, field := range node.Fields {
		if field.Key != nil {
			if ident, ok := field.Key.(*ast.Identifier); ok {
				p.write(ident.Name)
				p.write("=")
			} else {
				p.write("[")
				field.Key.Accept(p)
				p.write("]=")
			}
		}
		field.Value.Accept(p)
		if i < len(node.Fields)-1 {
			p.write(",")
		}
	}
	p.write("}")
	return nil
}

func (p *minifyPrinter) VisitFunctionExpr(node *ast.FunctionExpr) any {
	p.write("function")
	p.write("(")
	p.printParams(node.Parameters)
	p.write(")")
	if node.ReturnType != nil {
		p.write(":")
		node.ReturnType.Accept(p)
	}
	node.Body.Accept(p)
	p.write("end")
	return nil
}

func (p *minifyPrinter) VisitTypeCast(node *ast.TypeCast) any {
	node.Value.Accept(p)
	p.write("::")
	if node.Type != nil {
		node.Type.Accept(p)
	}
	return nil
}

func (p *minifyPrinter) VisitIfExpr(node *ast.IfExpr) any {
	p.write("if")
	node.Condition.Accept(p)
	p.write("then")
	node.Then.Accept(p)
	for _, elif := range node.ElseIfs {
		p.write("elseif")
		elif.Condition.Accept(p)
		p.write("then")
		elif.Then.Accept(p)
	}
	if node.Else != nil {
		p.write("else")
		node.Else.Accept(p)
	}
	return nil
}

func (p *minifyPrinter) VisitVarArgs(node *ast.VarArgs) any {
	p.write("...")
	return nil
}

func (p *minifyPrinter) VisitParenExpr(node *ast.ParenExpr) any {
	p.write("(")
	node.Expr.Accept(p)
	p.write(")")
	return nil
}

func (p *minifyPrinter) VisitInterpolatedString(node *ast.InterpolatedString) any {
	p.write("`")
	if len(node.Segments) > 0 {
		p.write(node.Segments[0])
	}
	for i, expr := range node.Expressions {
		p.write("{")
		expr.Accept(p)
		p.write("}")
		if i+1 < len(node.Segments) {
			p.write(node.Segments[i+1])
		}
	}
	p.write("`")
	return nil
}

func (p *minifyPrinter) VisitPrimitiveType(node *ast.PrimitiveType) any {
	p.write(node.Name)
	return nil
}

func (p *minifyPrinter) VisitUnionType(node *ast.UnionType) any {
	node.Left.Accept(p)
	p.write("|")
	node.Right.Accept(p)
	return nil
}

func (p *minifyPrinter) VisitOptionalType(node *ast.OptionalType) any {
	node.BaseType.Accept(p)
	p.write("?")
	return nil
}

func (p *minifyPrinter) VisitGenericType(node *ast.GenericType) any {
	node.BaseType.Accept(p)
	p.write("<")
	for i, t := range node.Types {
		t.Accept(p)
		if i < len(node.Types)-1 {
			p.write(",")
		}
	}
	p.write(">")
	return nil
}

func (p *minifyPrinter) VisitTableType(node *ast.TableType) any {
	p.write("{")
	for i, field := range node.Fields {
		if field.IsAccess {
			p.write("[")
			field.Key.Accept(p)
			p.write("]:")
		} else if field.KeyName != "" {
			p.write(field.KeyName + ":")
		}
		field.Value.Accept(p)
		if i < len(node.Fields)-1 {
			p.write(",")
		}
	}
	p.write("}")
	return nil
}
