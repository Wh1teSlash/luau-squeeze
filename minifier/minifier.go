package minifier

import "github.com/Wh1teSlash/luau-parser/ast"

func Minify(program *ast.Program) string {
	return NewPipeline().
		Add(&CommentStripPass{}).
		Add(&RenamePass{Strategy: &ShortRenamer{}}).
		Minify(program)
}
