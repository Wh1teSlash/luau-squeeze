package beautifier

import (
	"fmt"
	"strings"

	"github.com/Wh1teSlash/luau-parser/ast"
)

type beautifyPrinter struct {
	builder strings.Builder
	indent  int
	config  Config
}

func newBeautifyPrinter(config Config) *beautifyPrinter {
	return &beautifyPrinter{config: config}
}

func (p *beautifyPrinter) print(program *ast.Program) string {
	p.builder.Reset()
	p.indent = 0
	program.Accept(p)
	result := p.builder.String()
	if p.config.TrailingNewline && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}

func (p *beautifyPrinter) write(s string) {
	p.builder.WriteString(s)
}

func (p *beautifyPrinter) writeLine(s string) {
	p.builder.WriteString(s)
	p.builder.WriteByte('\n')
}

func (p *beautifyPrinter) writeIndent() {
	p.write(p.config.Indent.Indent(p.indent))
}

func (p *beautifyPrinter) writeStmtEnd() {
	if p.config.Semicolons {
		p.write(";")
	}
	p.write("\n")
}

func (p *beautifyPrinter) writeOp(op string) {
	if p.config.SpacesAroundOperators {
		p.write(" " + op + " ")
	} else {
		p.write(op)
	}
}

func (p *beautifyPrinter) printExprList(exprs []ast.Expr) {
	for i, expr := range exprs {
		expr.Accept(p)
		if i < len(exprs)-1 {
			p.write(", ")
		}
	}
}

func (p *beautifyPrinter) printParams(params []*ast.Parameter) {
	for i, param := range params {
		p.write(param.Name)
		if param.Type != nil {
			p.write(": ")
			param.Type.Accept(p)
		}
		if i < len(params)-1 {
			p.write(", ")
		}
	}
}

func (p *beautifyPrinter) isFunctionStmt(stmt ast.Stmt) bool {
	switch stmt.(type) {
	case *ast.FunctionDef, *ast.LocalFunction:
		return true
	}
	return false
}

func (p *beautifyPrinter) VisitProgram(node *ast.Program) any {
	for i, stmt := range node.Body {
		stmt.Accept(p)
		if p.config.BlankLineBetweenFunctions && p.isFunctionStmt(stmt) && i < len(node.Body)-1 {
			p.write("\n")
		}
	}
	return nil
}

func (p *beautifyPrinter) VisitBlock(node *ast.Block) any {
	for i, stmt := range node.Statements {
		stmt.Accept(p)
		if p.config.BlankLineBetweenFunctions && p.isFunctionStmt(stmt) && i < len(node.Statements)-1 {
			p.write("\n")
		}
	}
	return nil
}

func (p *beautifyPrinter) VisitModule(node *ast.Module) any {
	if node.Body != nil {
		node.Body.Accept(p)
	}
	return nil
}

func (p *beautifyPrinter) VisitComment(node *ast.Comment) any {
	p.writeIndent()
	p.write("-- ")
	p.writeLine(strings.TrimSpace(node.Text))
	return nil
}

func (p *beautifyPrinter) VisitAttribute(node *ast.Attribute) any {
	p.write("@" + node.Name)
	return nil
}

func (p *beautifyPrinter) VisitLocalAssignment(node *ast.LocalAssignment) any {
	p.writeIndent()
	p.write("local ")
	for i, name := range node.Names {
		p.write(name)
		if i < len(node.Types) && node.Types[i] != nil {
			p.write(": ")
			node.Types[i].Accept(p)
		}
		if i < len(node.Names)-1 {
			p.write(", ")
		}
	}
	if len(node.Values) > 0 {
		p.write(" = ")
		p.printExprList(node.Values)
	}
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitAssignment(node *ast.Assignment) any {
	p.writeIndent()
	p.printExprList(node.Targets)
	p.write(" ")
	p.write(node.Operator)
	p.write(" ")
	p.printExprList(node.Values)
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitIfStatement(node *ast.IfStatement) any {
	p.writeIndent()
	p.write("if ")
	node.Condition.Accept(p)
	p.writeLine(" then")

	p.indent++
	node.Then.Accept(p)
	p.indent--

	for _, elif := range node.ElseIfs {
		p.writeIndent()
		p.write("elseif ")
		elif.Condition.Accept(p)
		p.writeLine(" then")
		p.indent++
		elif.Body.Accept(p)
		p.indent--
	}

	if node.Else != nil {
		p.writeIndent()
		p.writeLine("else")
		p.indent++
		node.Else.Accept(p)
		p.indent--
	}

	p.writeIndent()
	p.write("end")
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitWhileLoop(node *ast.WhileLoop) any {
	p.writeIndent()
	p.write("while ")
	node.Condition.Accept(p)
	p.writeLine(" do")

	p.indent++
	node.Body.Accept(p)
	p.indent--

	p.writeIndent()
	p.write("end")
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitRepeatLoop(node *ast.RepeatLoop) any {
	p.writeIndent()
	p.writeLine("repeat")

	p.indent++
	node.Body.Accept(p)
	p.indent--

	p.writeIndent()
	p.write("until ")
	node.Condition.Accept(p)
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitForLoop(node *ast.ForLoop) any {
	p.writeIndent()
	p.write("for ")
	p.write(node.Variable)
	p.write(" = ")
	node.Start.Accept(p)
	p.write(", ")
	node.End.Accept(p)
	if node.Step != nil {
		p.write(", ")
		node.Step.Accept(p)
	}
	p.writeLine(" do")

	p.indent++
	node.Body.Accept(p)
	p.indent--

	p.writeIndent()
	p.write("end")
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitForInLoop(node *ast.ForInLoop) any {
	p.writeIndent()
	p.write("for ")
	p.write(strings.Join(node.Variables, ", "))
	p.write(" in ")
	p.printExprList(node.Iterables)
	p.writeLine(" do")

	p.indent++
	node.Body.Accept(p)
	p.indent--

	p.writeIndent()
	p.write("end")
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitDoBlock(node *ast.DoBlock) any {
	p.writeIndent()
	p.writeLine("do")

	p.indent++
	node.Body.Accept(p)
	p.indent--

	p.writeIndent()
	p.write("end")
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitFunctionDef(node *ast.FunctionDef) any {
	p.writeIndent()

	for _, attr := range node.Attributes {
		attr.Accept(p)
		p.write("\n")
		p.writeIndent()
	}

	p.write("function ")
	p.write(node.Name)
	p.write("(")
	p.printParams(node.Parameters)
	p.write(")")

	if node.ReturnType != nil {
		p.write(": ")
		node.ReturnType.Accept(p)
	}
	p.write("\n")

	p.indent++
	node.Body.Accept(p)
	p.indent--

	p.writeIndent()
	p.write("end")
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitLocalFunction(node *ast.LocalFunction) any {
	p.writeIndent()

	for _, attr := range node.Attributes {
		attr.Accept(p)
		p.write("\n")
		p.writeIndent()
	}

	p.write("local function ")
	p.write(node.Name)
	p.write("(")
	p.printParams(node.Parameters)
	p.write(")")

	if node.ReturnType != nil {
		p.write(": ")
		node.ReturnType.Accept(p)
	}
	p.write("\n")

	p.indent++
	node.Body.Accept(p)
	p.indent--

	p.writeIndent()
	p.write("end")
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitReturnStatement(node *ast.ReturnStatement) any {
	p.writeIndent()
	p.write("return")
	if len(node.Values) > 0 {
		p.write(" ")
		p.printExprList(node.Values)
	}
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitBreakStatement(node *ast.BreakStatement) any {
	p.writeIndent()
	p.write("break")
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitContinueStatement(node *ast.ContinueStatement) any {
	p.writeIndent()
	p.write("continue")
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitTypeAlias(node *ast.TypeAlias) any {
	p.writeIndent()
	if node.IsExport {
		p.write("export ")
	}
	p.write("type ")
	p.write(node.Name)
	if len(node.Generics) > 0 {
		p.write("<")
		p.write(strings.Join(node.Generics, ", "))
		p.write(">")
	}
	p.write(" = ")
	node.Type.Accept(p)
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitMetamethodDef(node *ast.MetamethodDef) any {
	p.writeIndent()
	p.write("function ")
	p.write(node.Name)
	p.write("(")
	p.printParams(node.Parameters)
	p.write(")\n")

	p.indent++
	node.Body.Accept(p)
	p.indent--

	p.writeIndent()
	p.write("end")
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitEmptyStatement(node *ast.EmptyStatement) any {
	return nil
}

func (p *beautifyPrinter) VisitExpressionStatement(node *ast.ExpressionStatement) any {
	p.writeIndent()
	node.Expr.Accept(p)
	p.writeStmtEnd()
	return nil
}

func (p *beautifyPrinter) VisitIdentifier(node *ast.Identifier) any {
	p.write(node.Name)
	return nil
}

func (p *beautifyPrinter) VisitLiteral(node *ast.Literal) any {
	if node.Type == "string" {
		p.write(fmt.Sprintf("%q", node.Value))
	} else {
		p.write(fmt.Sprintf("%v", node.Value))
	}
	return nil
}

func (p *beautifyPrinter) VisitBinaryOp(node *ast.BinaryOp) any {
	node.Left.Accept(p)
	p.writeOp(node.Op)
	node.Right.Accept(p)
	return nil
}

func (p *beautifyPrinter) VisitUnaryOp(node *ast.UnaryOp) any {
	p.write(node.Op)
	if node.Op == "not" {
		p.write(" ")
	}
	node.Operand.Accept(p)
	return nil
}

func (p *beautifyPrinter) VisitFunctionCall(node *ast.FunctionCall) any {
	node.Function.Accept(p)
	p.write("(")
	p.printExprList(node.Args)
	p.write(")")
	return nil
}

func (p *beautifyPrinter) VisitMethodCall(node *ast.MethodCall) any {
	node.Object.Accept(p)
	p.write(":")
	p.write(node.Method)
	p.write("(")
	p.printExprList(node.Args)
	p.write(")")
	return nil
}

func (p *beautifyPrinter) VisitIndexAccess(node *ast.IndexAccess) any {
	node.Table.Accept(p)
	p.write("[")
	node.Index.Accept(p)
	p.write("]")
	return nil
}

func (p *beautifyPrinter) VisitFieldAccess(node *ast.FieldAccess) any {
	node.Object.Accept(p)
	p.write(".")
	p.write(node.Field)
	return nil
}

func (p *beautifyPrinter) VisitTableLiteral(node *ast.TableLiteral) any {
	if len(node.Fields) == 0 {
		p.write("{}")
		return nil
	}
	p.writeLine("{")
	p.indent++
	for i, field := range node.Fields {
		p.writeIndent()
		if field.Key != nil {
			if ident, ok := field.Key.(*ast.Identifier); ok {
				p.write(ident.Name)
				p.write(" = ")
			} else {
				p.write("[")
				field.Key.Accept(p)
				p.write("] = ")
			}
		}
		field.Value.Accept(p)
		if i < len(node.Fields)-1 {
			p.write(",")
		}
		p.write("\n")
	}
	p.indent--
	p.writeIndent()
	p.write("}")
	return nil
}

func (p *beautifyPrinter) VisitFunctionExpr(node *ast.FunctionExpr) any {
	p.write("function(")
	p.printParams(node.Parameters)
	p.write(")")

	if node.ReturnType != nil {
		p.write(": ")
		node.ReturnType.Accept(p)
	}
	p.write("\n")

	p.indent++
	node.Body.Accept(p)
	p.indent--

	p.writeIndent()
	p.write("end")
	return nil
}

func (p *beautifyPrinter) VisitTypeCast(node *ast.TypeCast) any {
	node.Value.Accept(p)
	p.write(" :: ")
	if node.Type != nil {
		node.Type.Accept(p)
	}
	return nil
}

func (p *beautifyPrinter) VisitIfExpr(node *ast.IfExpr) any {
	p.write("if ")
	node.Condition.Accept(p)
	p.write(" then ")
	node.Then.Accept(p)
	for _, elif := range node.ElseIfs {
		p.write(" elseif ")
		elif.Condition.Accept(p)
		p.write(" then ")
		elif.Then.Accept(p)
	}
	if node.Else != nil {
		p.write(" else ")
		node.Else.Accept(p)
	}
	return nil
}

func (p *beautifyPrinter) VisitVarArgs(node *ast.VarArgs) any {
	p.write("...")
	return nil
}

func (p *beautifyPrinter) VisitParenExpr(node *ast.ParenExpr) any {
	p.write("(")
	node.Expr.Accept(p)
	p.write(")")
	return nil
}

func (p *beautifyPrinter) VisitInterpolatedString(node *ast.InterpolatedString) any {
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

func (p *beautifyPrinter) VisitPrimitiveType(node *ast.PrimitiveType) any {
	p.write(node.Name)
	return nil
}

func (p *beautifyPrinter) VisitUnionType(node *ast.UnionType) any {
	node.Left.Accept(p)
	p.write(" | ")
	node.Right.Accept(p)
	return nil
}

func (p *beautifyPrinter) VisitOptionalType(node *ast.OptionalType) any {
	node.BaseType.Accept(p)
	p.write("?")
	return nil
}

func (p *beautifyPrinter) VisitGenericType(node *ast.GenericType) any {
	node.BaseType.Accept(p)
	p.write("<")
	for i, t := range node.Types {
		t.Accept(p)
		if i < len(node.Types)-1 {
			p.write(", ")
		}
	}
	p.write(">")
	return nil
}

func (p *beautifyPrinter) VisitTableType(node *ast.TableType) any {
	p.write("{ ")
	for i, field := range node.Fields {
		if field.IsAccess {
			p.write("[")
			field.Key.Accept(p)
			p.write("]: ")
		} else if field.KeyName != "" {
			p.write(field.KeyName + ": ")
		}
		field.Value.Accept(p)
		if i < len(node.Fields)-1 {
			p.write(", ")
		}
	}
	p.write(" }")
	return nil
}
