package analysis

import (
	"fmt"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/document"
)

// Functions finds all function definitions that are direct children of the provided sitter.Node.
func Functions(doc document.Document, node *sitter.Node) map[string]protocol.SignatureInformation {
	signatures := make(map[string]protocol.SignatureInformation)

	// N.B. we don't use a query here for a couple reasons:
	// 	(1) Tree-sitter doesn't support bounding the depth, and we only want
	//		direct descendants (to avoid matching on functions in nested scopes)
	//		See https://github.com/tree-sitter/tree-sitter/issues/1212.
	//	(2) function_definition nodes have named fields for what we care about,
	//		which makes it easy to get the data without using a query to help
	//		massage/standardize it (for example, we do this for params since
	//		there are multiple type of param values)
	for n := node.NamedChild(0); n != nil; n = n.NextNamedSibling() {
		if n.Type() != NodeTypeFunctionDef {
			continue
		}
		fnName, sig := extractSignatureInformation(doc, n)
		signatures[fnName] = sig
	}

	return signatures
}

// Function finds a function definition for the given function name that is a direct child of the provided sitter.Node.
func Function(doc document.Document, node *sitter.Node, fnName string) (protocol.SignatureInformation, bool) {
	for n := node.NamedChild(0); n != nil; n = n.NextNamedSibling() {
		if n.Type() != NodeTypeFunctionDef {
			continue
		}
		curFuncName := n.ChildByFieldName(FieldName).Content(doc.Contents)
		if curFuncName == fnName {
			_, sig := extractSignatureInformation(doc, n)
			return sig, true
		}
	}
	return protocol.SignatureInformation{}, false
}

func extractSignatureInformation(doc document.Document, n *sitter.Node) (string, protocol.SignatureInformation) {
	if n.Type() != NodeTypeFunctionDef {
		panic(fmt.Errorf("invalid node type: %s", n.Type()))
	}

	fnName := n.ChildByFieldName(FieldName).Content(doc.Contents)
	// params might be empty but a node for `()` will still exist
	params := extractParameters(doc, n.ChildByFieldName(FieldParameters))
	// unlike name + params, returnType is optional
	var returnType string
	if rtNode := n.ChildByFieldName(FieldReturnType); rtNode != nil {
		returnType = rtNode.Content(doc.Contents)
	}

	sig := protocol.SignatureInformation{
		Label:      signatureLabel(params, returnType),
		Parameters: params,
	}

	return fnName, sig
}

// signatureLabel produces a human-readable label for a function signature.
//
// It's modeled to behave similarly to VSCode Python signature labels.
func signatureLabel(params []protocol.ParameterInformation, returnType string) string {
	if returnType == "" {
		returnType = "None"
	}

	var sb strings.Builder
	sb.WriteRune('(')
	for i := range params {
		sb.WriteString(params[i].Label)
		if i != len(params)-1 {
			sb.WriteString(", ")
		}
	}
	sb.WriteString(") -> ")
	sb.WriteString(returnType)
	return sb.String()
}
