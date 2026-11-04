// Command templwrap stacks the attributes of templ elements, one attribute per line with ">"
// on its own line, which is what `templ fmt` produces for a tag that was broken up by hand.
// It reads a templ file on stdin and writes it to stdout; run `templ fmt` on the result.
// A file with nothing to change comes back byte for byte.
//
//	templwrap < views/page.templ | templ fmt -stdin-filepath views/page.templ
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

// config says which tags are stacked and what else is tidied.
type config struct {
	width int  // stack a tag whose opening tag would end beyond this column
	tab   int  // columns a tab counts for
	attrs int  // stack a tag with at least this many attributes (0: only by width)
	sort  bool // put the attributes of every tag in a fixed order
}

func main() {
	var c config
	flag.IntVar(&c.width, "width", 100, "stack tags whose opening tag ends beyond this column")
	flag.IntVar(&c.tab, "tab", 4, "columns a tab counts for")
	flag.IntVar(&c.attrs, "attrs", 3, "stack tags with at least this many attributes (0: only by width)")
	flag.BoolVar(&c.sort, "sort", false, "order attributes: id, class; name, type, slot; others; x-*; hx-*; data-*, aria-*")
	flag.Parse()

	src, err := io.ReadAll(os.Stdin)
	if err != nil {
		fail(err)
	}
	out, err := wrap(string(src), c)
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

// wrap returns src with the attributes of the tags that qualify stacked.
func wrap(src string, c config) (string, error) {
	tf, err := parser.ParseString(src)
	if err != nil {
		return "", err
	}

	changed := false
	v := visitor.New()
	visitElement := v.Element
	v.Element = func(e *parser.Element) error {
		if c.sort && sortAttributes(e.Attributes) {
			changed = true
		}
		if !e.IndentAttrs && stack(src, e, c) {
			e.IndentAttrs = true
			changed = true
		}
		// ">{ x }</a>" after the last attribute looks odd: children go on their own lines.
		// templ drops whitespace at the start and end of an element's content either way,
		// so the generated code stays the same. pre and textarea keep theirs.
		if e.IndentAttrs && !e.IndentChildren && !keepsWhitespace(e.Name) {
			e.IndentChildren = true
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

func keepsWhitespace(name string) bool {
	return strings.EqualFold(name, "pre") || strings.EqualFold(name, "textarea")
}

// stack reports whether the element's attributes go one per line.
func stack(src string, e *parser.Element, c config) bool {
	if len(e.Attributes) < 2 {
		return false
	}
	return c.attrs > 0 && len(e.Attributes) >= c.attrs || tooLong(src, e.OpenTagRange, c)
}

// tooLong reports whether the opening tag, which has to start and end on one line, would
// reach past the width if it stood at the indentation of its line. Siblings before it on
// the same line don't count, or a short tag after a long one would be stacked too.
func tooLong(src string, r parser.Range, c config) bool {
	from, to := int(r.From.Index), int(r.To.Index)
	if from < 0 || to > len(src) || from >= to || strings.Contains(src[from:to], "\n") {
		return false
	}
	start := strings.LastIndexByte(src[:from], '\n') + 1
	col := 0
	for _, r := range src[start:from] {
		if r != '\t' && r != ' ' {
			break
		}
		if r == '\t' {
			col += c.tab
		} else {
			col++
		}
	}
	return col+utf8.RuneCountInString(src[from:to]) > c.width
}
