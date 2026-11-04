package main

import (
	"strings"
	"testing"
)

const long = `package views

templ A(x string) {
	<div>
		<wa-button href="/add" { pageLink("/add")... } size="s" variant="brand">
			{ x }
		</wa-button>
	</div>
}
`

func TestWrapStacksLongTag(t *testing.T) {
	out, err := wrap(long, 60, 4)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"\t\t<wa-button\n\t\t\thref=\"/add\"\n\t\t\t{ pageLink(\"/add\")... }\n\t\t\tsize=\"s\"\n\t\t\tvariant=\"brand\"\n",
		"\t\t>\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestWrapKeepsFileWithNothingToWrap(t *testing.T) {
	// Not in templ fmt's style (no space before "/>"): must come back untouched.
	src := "package views\n\ntempl A() {\n\t<img src=\"a\" alt=\"b\"  />\n}\n"
	out, err := wrap(src, 100, 4)
	if err != nil {
		t.Fatal(err)
	}
	if out != src {
		t.Errorf("changed a file with nothing to wrap:\n%s", out)
	}
}

func TestWrapLeavesSingleAttribute(t *testing.T) {
	src := "package views\n\ntempl A() {\n\t<a href=\"" + strings.Repeat("x", 200) + "\">hi</a>\n}\n"
	out, err := wrap(src, 100, 4)
	if err != nil {
		t.Fatal(err)
	}
	if out != src {
		t.Errorf("stacked a tag with one attribute:\n%s", out)
	}
}

func TestWrapIsIdempotent(t *testing.T) {
	once, err := wrap(long, 60, 4)
	if err != nil {
		t.Fatal(err)
	}
	twice, err := wrap(once, 60, 4)
	if err != nil {
		t.Fatal(err)
	}
	if once != twice {
		t.Errorf("second run changed the output:\n%s\n---\n%s", once, twice)
	}
}

func TestWrapReportsSyntaxErrors(t *testing.T) {
	if _, err := wrap("package views\n\ntempl A( {\n", 100, 4); err == nil {
		t.Error("no error for a broken file")
	}
}
