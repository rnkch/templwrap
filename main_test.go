package main

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/a-h/templ/generator"
	"github.com/a-h/templ/parser/v2"
)

// narrow makes short tags qualify.
var narrow = config{width: 40, tab: 4, attrs: 3}

const long = `package views

templ A(x string) {
	<div>
		<wa-button href="/add" { pageLink("/add")... } size="s" variant="brand">
			{ x }
		</wa-button>
	</div>
}
`

func mustWrap(t *testing.T, src string, c config) string {
	t.Helper()
	out, err := wrap(src, c)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func TestStacksLongTag(t *testing.T) {
	out := mustWrap(t, long, config{width: 60, tab: 4})
	for _, want := range []string{
		"\t\t<wa-button\n\t\t\thref=\"/add\"\n\t\t\t{ pageLink(\"/add\")... }\n\t\t\tsize=\"s\"\n\t\t\tvariant=\"brand\"\n",
		"\t\t>\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestStacksByAttributeCount(t *testing.T) {
	src := "package v\n\ntempl A() {\n\t<p a=\"1\" b=\"2\" c=\"3\"></p>\n\t<p a=\"1\" b=\"2\"></p>\n}\n"
	out := mustWrap(t, src, config{width: 100, tab: 4, attrs: 3})
	if !strings.Contains(out, "\t<p\n\t\ta=\"1\"\n\t\tb=\"2\"\n\t\tc=\"3\"\n\t></p>") {
		t.Errorf("three attributes not stacked:\n%s", out)
	}
	if !strings.Contains(out, "\t<p a=\"1\" b=\"2\"></p>") {
		t.Errorf("two attributes stacked:\n%s", out)
	}
}

func TestChildrenOfStackedTagsGoOnTheirOwnLines(t *testing.T) {
	src := "package v\n\ntempl A(x string) {\n\t<p\n\t\ta=\"1\"\n\t\tb=\"2\"\n\t>{ x } %</p>\n}\n"
	out := mustWrap(t, src, config{width: 100, tab: 4})
	if !strings.Contains(out, "\t>\n\t\t{ x } %\n\t</p>") {
		t.Errorf("children glued to the tag:\n%s", out)
	}
}

func TestKeepsFileWithNothingToChange(t *testing.T) {
	// Not in templ fmt's style (two spaces before "/>"): must come back untouched.
	src := "package views\n\ntempl A() {\n\t<img src=\"a\" alt=\"b\"  />\n}\n"
	if out := mustWrap(t, src, config{width: 100, tab: 4}); out != src {
		t.Errorf("changed a file with nothing to change:\n%s", out)
	}
}

func TestLeavesSingleAttribute(t *testing.T) {
	src := "package views\n\ntempl A() {\n\t<a href=\"" + strings.Repeat("x", 200) + "\">hi</a>\n}\n"
	if out := mustWrap(t, src, narrow); out != src {
		t.Errorf("stacked a tag with one attribute:\n%s", out)
	}
}

func TestIsIdempotent(t *testing.T) {
	c := config{width: 60, tab: 4, attrs: 3, sort: true}
	once := mustWrap(t, long, c)
	if twice := mustWrap(t, once, c); once != twice {
		t.Errorf("second run changed the output:\n%s\n---\n%s", once, twice)
	}
}

func TestReportsSyntaxErrors(t *testing.T) {
	if _, err := wrap("package views\n\ntempl A( {\n", narrow); err == nil {
		t.Error("no error for a broken file")
	}
}

const sortable = `package v

templ A(x string) {
	<wa-button x-on:click="a" hx-get="/x" data-k="1" size="s" class="c" type="button" id="i" aria-label="l" x-show="b"></wa-button>
	<p b="1" { x... } a="2" class="c" id="i"></p>
	<p b="1" if x != "" { z="1" } a="2" class="c"></p>
}
`

func TestSortOrdersAttributes(t *testing.T) {
	out := mustWrap(t, sortable, config{sort: true, tab: 4})
	for _, want := range []string{
		"\t<wa-button\n\t\tid=\"i\"\n\t\tclass=\"c\"\n\t\ttype=\"button\"\n\t\tsize=\"s\"\n\t\tx-on:click=\"a\"\n" +
			"\t\tx-show=\"b\"\n\t\thx-get=\"/x\"\n\t\tdata-k=\"1\"\n\t\taria-label=\"l\"\n\t></wa-button>",
		// nothing moves across a spread or a conditional attribute
		"\t<p\n\t\tb=\"1\"\n\t\t{ x... }\n\t\tid=\"i\"\n\t\tclass=\"c\"\n\t\ta=\"2\"\n\t></p>",
		"\t<p\n\t\tb=\"1\"\n\t\tif x != \"\" {\n\t\t\tz=\"1\"\n\t\t}\n\t\tclass=\"c\"\n\t\ta=\"2\"\n\t></p>",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output lacks %q:\n%s", want, out)
		}
	}
}

func TestLeavesScriptAndStyle(t *testing.T) {
	src := "package v\n\ntempl A() {\n\t<script a=\"1\" b=\"2\" c=\"3\">var x = 1;</script>\n\t<style a=\"1\" b=\"2\" c=\"3\">p { color: red }</style>\n}\n"
	if out := mustWrap(t, src, narrow); out != src {
		t.Errorf("changed script or style:\n%s", out)
	}
}

// Everything but sorting must leave the generated Go code as it is (apart from the positions
// in its error messages): the same HTML comes out. This is the check that makes it safe to
// run on a whole project.
func TestKeepsGeneratedCode(t *testing.T) {
	for name, src := range map[string]string{
		"inline":      "package v\n\ntempl A(x string) {\n\t<p a=\"1\" b=\"2\" c=\"3\"><b>a</b> <i>b</i> { x } more</p>\n}\n",
		"text":        "package v\n\ntempl A(x string) {\n\t<span a=\"1\" b=\"2\" c=\"3\">  { x } %  </span>\n}\n",
		"pre":         "package v\n\ntempl A() {\n\t<pre a=\"1\" b=\"2\" c=\"3\">  keep\n   this </pre>\n}\n",
		"textarea":    "package v\n\ntempl A() {\n\t<textarea a=\"1\" b=\"2\" c=\"3\">  keep\n   this </textarea>\n}\n",
		"spread":      "package v\n\ntempl A(m templ.Attributes) {\n\t<a { m... } a=\"1\" b=\"2\">x</a>\n}\n",
		"conditional": "package v\n\ntempl A(x bool) {\n\t<a a=\"1\" b=\"2\" if x { c=\"3\" } else { d=\"4\" }>x</a>\n}\n",
		"component":   "package v\n\ntempl B() {\n\t{ children... }\n}\n\ntempl A() {\n\t@B() {\n\t\t<i a=\"1\" b=\"2\" c=\"3\">x</i> <i a=\"1\" b=\"2\" c=\"3\">y</i>\n\t}\n}\n",
		"control":     "package v\n\ntempl A(l []string) {\n\tfor _, s := range l {\n\t\tif s != \"\" {\n\t\t\t<li a=\"1\" b=\"2\" c=\"3\">{ s }</li>\n\t\t}\n\t}\n}\n",
		"void":        "package v\n\ntempl A() {\n\t<img src=\"a\" alt=\"b\" loading=\"lazy\"/><br/>\n}\n",
	} {
		t.Run(name, func(t *testing.T) {
			out := mustWrap(t, src, narrow)
			if out == src {
				t.Fatal("test input was not changed")
			}
			if a, b := generate(t, src), generate(t, out); a != b {
				t.Errorf("generated code differs\n--- before\n%s\n--- after\n%s\n--- wrapped source\n%s", a, b, out)
			}
		})
	}
}

var position = regexp.MustCompile("FileName: `[^`]*`, Line: [0-9]+, Col: [0-9]+")

func generate(t *testing.T, src string) string {
	t.Helper()
	tf, err := parser.ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if _, err := generator.Generate(tf, &buf, generator.WithSkipCodeGeneratedComment()); err != nil {
		t.Fatal(err)
	}
	return position.ReplaceAllString(buf.String(), "POS")
}
