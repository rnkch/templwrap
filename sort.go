package main

import (
	"slices"
	"strings"

	"github.com/a-h/templ/parser/v2"
)

// sortAttributes puts attributes in a fixed order and reports whether that moved any.
// Spreads, conditional attributes and attributes with a computed name are fixed points:
// what follows them can override what precedes them, so nothing moves across one. The sort
// is stable, so attributes of one rank keep their order.
func sortAttributes(attrs []parser.Attribute) bool {
	moved := false
	start := 0
	for i := 0; i <= len(attrs); i++ {
		if i < len(attrs) {
			if _, ok := attrName(attrs[i]); ok {
				continue
			}
		}
		seg := attrs[start:i]
		sorted := slices.Clone(seg)
		slices.SortStableFunc(sorted, func(a, b parser.Attribute) int {
			an, _ := attrName(a)
			bn, _ := attrName(b)
			return rank(an) - rank(bn)
		})
		if !slices.Equal(seg, sorted) {
			copy(seg, sorted)
			moved = true
		}
		start = i + 1
	}
	return moved
}

// attrName is the attribute's name; ok is false for attributes that must stay where they are.
func attrName(a parser.Attribute) (name string, ok bool) {
	var key parser.AttributeKey
	switch a := a.(type) {
	case *parser.ConstantAttribute:
		key = a.Key
	case *parser.BoolConstantAttribute:
		key = a.Key
	case *parser.ExpressionAttribute:
		key = a.Key
	case *parser.BoolExpressionAttribute:
		key = a.Key
	}
	k, isConst := key.(parser.ConstantAttributeKey)
	return k.Name, isConst
}

func rank(name string) int {
	switch {
	case name == "id":
		return 0
	case name == "class":
		return 1
	case name == "name" || name == "type" || name == "slot":
		return 2
	case strings.HasPrefix(name, "x-") || strings.HasPrefix(name, "@") || strings.HasPrefix(name, ":"):
		return 4
	case strings.HasPrefix(name, "hx-"):
		return 5
	case strings.HasPrefix(name, "data-") || strings.HasPrefix(name, "aria-"):
		return 6
	}
	return 3
}
