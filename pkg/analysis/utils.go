package analysis

import (
	"fmt"
	"regexp"

	"github.com/autokitteh/starlark-lsp/pkg/document"
	sitter "github.com/smacker/go-tree-sitter"
)

func printNodeTree(d document.Document, n *sitter.Node, indent string) string {
	nodeType := "U"
	if n.IsNamed() {
		nodeType = "N"
	}
	result := fmt.Sprintf("\n%s%s (%s): %s", indent, n.Type(), nodeType, d.Content(n))
	indent += "  "
	for i := 0; i < int(n.ChildCount()); i++ {
		child := n.Child(i)
		result += printNodeTree(d, child, indent)
	}
	return result
}

func removeBrackets(s string) string { // remove all brackets (..)
	pattern := `\([^()]*\)`
	re := regexp.MustCompile(pattern)
	for re.MatchString(s) {
		s = re.ReplaceAllString(s, "")
	}
	return s
}

func replaceKnownTypes(parts []string) {
	replacementMap := map[byte]string{'"': "String", '{': "Dict", '[': "List"}
	for i, s := range parts {
		if len(s) > 0 {
			if t, found := replacementMap[s[0]]; found {
				parts[i] = t
			}
		}
	}
}
