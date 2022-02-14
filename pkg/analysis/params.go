package analysis

import (
	"fmt"

	sitter "github.com/smacker/go-tree-sitter"
	"go.lsp.dev/protocol"

	"github.com/tilt-dev/starlark-lsp/pkg/docstring"
	"github.com/tilt-dev/starlark-lsp/pkg/document"
	"github.com/tilt-dev/starlark-lsp/pkg/query"
)

type parameter struct {
	name         string
	typeHint     string
	defaultValue string
	content      string
}

func (p parameter) paramInfo(fnDocs docstring.Parsed) protocol.ParameterInformation {
	var docs string
	for _, fieldsBlock := range fnDocs.Fields {
		if fieldsBlock.Title != "Args" {
			continue
		}
		for _, f := range fieldsBlock.Fields {
			if f.Name == p.name {
				docs = f.Desc
			}
		}
	}

	// TODO(milas): revisit labels - with type hints this can make signatures
	// 	really long; it might make sense to only include param name and default
	// 	value (if any)
	return protocol.ParameterInformation{
		Label: p.content,
		Documentation: protocol.MarkupContent{
			Kind:  protocol.Markdown,
			Value: docs,
		},
	}
}

func extractParameters(doc document.Document, fnDocs docstring.Parsed,
	node *sitter.Node) []protocol.ParameterInformation {
	if node.Type() != NodeTypeParameters {
		// A query is used here because there's several different node types
		// for parameter values, and the query handles normalization gracefully
		// for us.
		//
		// Technically, the query will execute regardless of passed in node
		// type, but since Tree-sitter doesn't currently support bounding query
		// depth, that could result in finding parameters from funcs in nested
		// scopes if a block was passed, so this ensures that the actual
		// `parameters` node is passed in here.
		//
		// See https://github.com/tree-sitter/tree-sitter/issues/1212
		panic(fmt.Errorf("invalid node type: %v", node.Type()))
	}

	var params []protocol.ParameterInformation
	Query(node, query.FunctionParameters, func(q *sitter.Query, match *sitter.QueryMatch) bool {
		var param parameter

		for _, c := range match.Captures {
			content := c.Node.Content(doc.Contents)
			switch q.CaptureNameForId(c.Index) {
			case "name":
				param.name = content
			case "type":
				param.typeHint = content
			case "value":
				param.defaultValue = content
			case "param":
				param.content = content
			}
		}

		params = append(params, param.paramInfo(fnDocs))
		return true
	})
	return params
}