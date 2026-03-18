package minifier

import "strings"

type RenameStrategy interface {
	Next(original string) string
	Reset()
}

var luauKeywords = map[string]bool{
	"and": true, "break": true, "do": true, "else": true,
	"elseif": true, "end": true, "false": true, "for": true,
	"function": true, "if": true, "in": true, "local": true,
	"nil": true, "not": true, "or": true, "repeat": true,
	"return": true, "then": true, "true": true, "until": true,
	"while": true, "continue": true,
}

func generateShortName(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz"
	for {
		name := ""
		i := n
		for {
			name = string(chars[i%26]) + name
			i = i/26 - 1
			if i < 0 {
				break
			}
		}
		if !luauKeywords[name] {
			return name
		}
		n++
	}
}

type ShortRenamer struct {
	counter int
}

func (r *ShortRenamer) Next(_ string) string {
	name := generateShortName(r.counter)
	r.counter++
	return name
}

func (r *ShortRenamer) Reset() { r.counter = 0 }

type ReadableRenamer struct {
	counter int
}

func (r *ReadableRenamer) Next(original string) string {
	suffix := generateShortName(r.counter)
	r.counter++
	if len(original) > 8 {
		original = original[:8]
	}
	return original + "_" + suffix
}

func (r *ReadableRenamer) Reset() { r.counter = 0 }

type NoopRenamer struct{}

func (r *NoopRenamer) Next(original string) string { return original }
func (r *NoopRenamer) Reset()                      {}

type PrefixRenamer struct {
	prefix  string
	counter int
}

func NewPrefixRenamer(prefix string) *PrefixRenamer {
	return &PrefixRenamer{prefix: prefix}
}

func (r *PrefixRenamer) Next(_ string) string {
	name := r.prefix + generateShortName(r.counter)
	r.counter++
	return name
}

func (r *PrefixRenamer) Reset() { r.counter = 0 }

type UpperCaseRenamer struct {
	counter int
}

func (r *UpperCaseRenamer) Next(_ string) string {
	name := strings.ToUpper(generateShortName(r.counter))
	r.counter++
	return name
}

func (r *UpperCaseRenamer) Reset() { r.counter = 0 }
