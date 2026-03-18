package beautifier

import "strings"

type IndentStrategy interface {
	Indent(level int) string
}

type TabIndent struct{}

func (t *TabIndent) Indent(level int) string {
	return strings.Repeat("\t", level)
}

type SpaceIndent struct {
	Size int // default: 4
}

func (s *SpaceIndent) Indent(level int) string {
	size := s.Size
	if size <= 0 {
		size = 4
	}
	return strings.Repeat(" ", level*size)
}
