// Command templwrap stacks the attributes of templ elements whose opening tag is too long,
// one attribute per line with ">" on its own line, which is what `templ fmt` produces for a
// tag that was broken up by hand. It reads a templ file on stdin and writes it to stdout;
// run `templ fmt` on the result. A file with nothing to wrap comes back byte for byte.
//
//	templwrap -width 100 < views/page.templ | templ fmt -stdin-filepath views/page.templ
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/a-h/templ/parser/v2"
	"github.com/a-h/templ/parser/v2/visitor"
)

func main() {
	width := flag.Int("width", 100, "stack attributes when the opening tag ends beyond this column")
	tab := flag.Int("tab", 4, "columns a tab counts for")
	flag.Parse()

	src, err := io.ReadAll(os.Stdin)
	if err != nil {
		fail(err)
	}
	out, err := wrap(string(src), *width, *tab)
	if err != nil {
		fail(err)
	}
	if _, err := io.WriteString(os.Stdout, out); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "templwrap:", err)
	os.Exit(1)
}

// wrap returns src with the attributes of every too-long opening tag stacked.
func wrap(src string, width, tab int) (string, error) {
	tf, err := parser.ParseString(src)
	if err != nil {
		return "", err
	}

	changed := false
	v := visitor.New()
	visitElement := v.Element
	v.Element = func(e *parser.Element) error {
		if !e.IndentAttrs && len(e.Attributes) >= 2 && tooLong(src, e.OpenTagRange, width, tab) {
			e.IndentAttrs = true
			changed = true
		}
		return visitElement(e)
	}
	if err := v.VisitTemplateFile(tf); err != nil {
		return "", err
	}
	if !changed {
		return src, nil
	}

	var buf bytes.Buffer
	if err := tf.Write(&buf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// tooLong reports whether the opening tag, which has to start and end on one line, would
// reach past the width if it stood at the indentation of its line. Siblings before it on
// the same line don't count, or a short tag after a long one would be stacked too.
func tooLong(src string, r parser.Range, width, tab int) bool {
	from, to := int(r.From.Index), int(r.To.Index)
	if from < 0 || to > len(src) || from >= to || strings.Contains(src[from:to], "\n") {
		return false
	}
	start := strings.LastIndexByte(src[:from], '\n') + 1
	col := 0
	for _, c := range src[start:from] {
		if c != '\t' && c != ' ' {
			break
		}
		if c == '\t' {
			col += tab
		} else {
			col++
		}
	}
	return col+utf8.RuneCountInString(src[from:to]) > width
}
