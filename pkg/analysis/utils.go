package analysis

import (
	"fmt"

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
