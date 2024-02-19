package analysis

import (
	"fmt"
	"strings"

	"go.lsp.dev/protocol"
	"go.uber.org/zap"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/autokitteh/starlark-lsp/pkg/document"
	"github.com/autokitteh/starlark-lsp/pkg/query"
)

var gAvailableSymbols []query.Symbol // cache computed available symbols

func SymbolMatching(symbols []query.Symbol, name string) query.Symbol {
	if name == "" {
		return query.Symbol{}
	}
	for _, sym := range symbols {
		if sym.Name == name {
			return sym
		}
	}
	return query.Symbol{}
}

// ak: deal with binded symbols (sym.Name -> sym.Detail) ---------------------
func akIsBindedSymbol(sym query.Symbol) bool {
	if len(sym.Tags) > 0 {
		for _, tag := range sym.Tags {
			if tag == query.Binded {
				return true
			}
		}
	}
	return false
}

// find and return either symbol or resolved binded symbol
func akSymbolMatching(symbols []query.Symbol, name string) query.Symbol {
	if name == "" {
		return query.Symbol{}
	}
	sym := SymbolMatching(symbols, name)
	if akIsBindedSymbol(sym) {
		return SymbolMatching(symbols, sym.Detail)
	}
	return sym
} // -------------------------------------------------------------------------

func SymbolsStartingWith(symbols []query.Symbol, prefix string) []query.Symbol {
	if prefix == "" {
		return symbols
	}
	result := []query.Symbol{}
	for _, sym := range symbols {
		if strings.HasPrefix(sym.Name, prefix) {
			result = append(result, sym)
		}
	}
	return result
}

func ToCompletionItemKind(k protocol.SymbolKind) protocol.CompletionItemKind {
	switch k {
	case protocol.SymbolKindField:
		return protocol.CompletionItemKindField
	case protocol.SymbolKindFunction:
		return protocol.CompletionItemKindFunction
	case protocol.SymbolKindMethod:
		return protocol.CompletionItemKindMethod
	default:
		return protocol.CompletionItemKindVariable
	}
}

func (a *Analyzer) Completion(doc document.Document, pos protocol.Position) *protocol.CompletionList {
	pt := query.PositionToPoint(pos)
	nodes, ok := a.nodesAtPointForCompletion(doc, pt)
	symbols := []query.Symbol{}

	if ok {
		symbols = a.completeExpression(doc, nodes, pt)
	}

	completionList := &protocol.CompletionList{
		Items: make([]protocol.CompletionItem, len(symbols)),
	}

	names := make([]string, len(symbols))
	for i, sym := range symbols {
		names[i] = sym.Name
		var sortText string
		if strings.HasSuffix(sym.Name, "=") {
			sortText = fmt.Sprintf("0%s", sym.Name)
		} else {
			sortText = fmt.Sprintf("1%s", sym.Name)
		}
		firstDetailLine := strings.SplitN(sym.Detail, "\n", 2)[0]
		completionList.Items[i] = protocol.CompletionItem{
			Label:    sym.Name,
			Detail:   firstDetailLine,
			Kind:     ToCompletionItemKind(sym.Kind),
			SortText: sortText,
		}
	}

	if len(names) > 0 {
		a.logger.Debug("completion result", zap.Strings("symbols", names))
	}
	return completionList
}

func (a *Analyzer) completeExpression(doc document.Document, nodes []*sitter.Node, pt sitter.Point) []query.Symbol {
	var nodeAtPoint *sitter.Node
	if len(nodes) > 0 {
		nodeAtPoint = nodes[len(nodes)-1]
	}

	symbols := a.availableSymbols(doc, nodeAtPoint, pt)
	gAvailableSymbols = symbols
	identifiers := query.ExtractIdentifiers(doc, nodes, &pt)

	a.logger.Debug("completion attempt",
		zap.String("code", document.NodesToContent(doc, nodes)),
		zap.Strings("identifiers", identifiers),
		zap.Strings("nodes", Transform(nodes, func(n *sitter.Node) string { return n.String() })),
	)

	if len(identifiers) == 0 {
		return []query.Symbol{}
	}

	if nodeAtPoint == nil || (len(identifiers) == 1 && identifiers[0] == "") {
		return symbols
	}

	if len(identifiers) == 1 { // REVIEW: could we miss any case?
		symbols = SymbolsStartingWith(gAvailableSymbols, identifiers[0])
		if len(symbols) > 0 {
			return symbols
		}
	}

	sym := a.resolveNode(doc, nodes[0])
	if sym.Type != "" {
		identifiers = append([]string{sym.GetType()}, identifiers[1:]...)
	}
	symbols, _ = a.builtinsCompletion(doc, identifiers)
	return symbols
}

func (a *Analyzer) builtinsCompletion(doc document.Document, identifiers []string) (symbols []query.Symbol, ok bool) {
	symbols = a.builtins.Symbols // we need builtins and doc-rebinded ones
	for _, sym := range doc.Symbols() {
		if akIsBindedSymbol(sym) {
			symbols = append(symbols, sym)
		}
	}

	ids := identifiers
	if symbols, ids = a.builtinFuncCompletion(symbols, ids); len(ids) == 0 && len(symbols) > 0 {
		return symbols, true
	}
	if symbols, ids = a.builtinTypeNestedCompletion(ids); len(ids) == 0 && len(symbols) > 0 {
		return symbols, true
	}
	if len(identifiers) != len(ids) {
		a.logger.Debug("incomplete via builtins", zap.Strings("orig ids", identifiers), zap.Strings("new ids", ids))
		return symbols, true // something was completed. Don't proceed
	}
	return symbols, false
}

func (a *Analyzer) builtinFuncCompletion(symbols []query.Symbol, identifiers []string) ([]query.Symbol, []string) {
	if len(identifiers) == 0 || identifiers[0] == "" {
		return []query.Symbol{}, identifiers
	}
	a.logger.Debug("completion attempt builtin func", zap.Strings("ids", identifiers))

	// start with all builtins and narrow down. dir.file.func builtin imported as a tree,
	// where dir is uplevel symbol, file is a child, and func is a child of file

	var i int
	var id string
	var sym query.Symbol
	for i, id = range identifiers {
		sym = akSymbolMatching(symbols, id)
		if i == 0 && sym.Name != "" {
			if sym.Name != id { // ak: -----------------------------------------------
				identifiers[0] = sym.Name // replace remapped top level builtin symbol
			} // ---------------------------------------------------------------------
		}
		if i == len(identifiers)-1 || sym.Children == nil { // last id, no exact sym or no children
			symbols = SymbolsStartingWith(symbols, id)
			break
		}

		symbols = sym.Children // narrow down
		a.logger.Debug("children", zap.String("id", id), zap.Strings("names", query.SymbolNames(symbols)))
	}

	if i == len(identifiers)-1 && len(symbols) > 0 {
		return symbols, []string{}
	}

	if sym.Name != "" {
		identifiers = identifiers[i+1:]
	}

	if len(symbols) == 1 && sym.Name != "" { // exact, but uncomplete resolving, update type
		if t := query.SymbolKindToBuiltinType(sym.Kind); t != "" && i == 0 { // String, Dict, List
			identifiers = append([]string{t}, identifiers...)
		} else { // method/function
			identifiers = append([]string{sym.Type}, identifiers...) // method/function ret type
		}
	}

	return symbols, identifiers
}

func (a *Analyzer) builtinTypeNestedCompletion(identifiers []string) ([]query.Symbol, []string) {
	a.logger.Debug("completion attempt builtin types/members", zap.Strings("ids", identifiers))

	symbols := []query.Symbol{}
	if len(identifiers) == 0 {
		return symbols, identifiers
	}

	t := identifiers[0]

	if len(identifiers) == 1 { // just check presence and wrap type as symbol
		if _, found := a.builtins.Types[t]; found {
			identifiers = identifiers[1:]
			symbols = []query.Symbol{{Name: t, Kind: protocol.SymbolKindClass, Type: t}}
		}
		return symbols, identifiers
	}

	found := false
	for i, id := range identifiers {
		if i > 0 { // narrow
			symbols = SymbolsStartingWith(symbols, id)
			if len(symbols) == 0 {
				break
			}
			if len(symbols) > 1 || len(identifiers) == 1 { // cannot narrow more or last identifier.
				identifiers = identifiers[1:] // mark this identifier as used
				break
			}
			t = symbols[0].GetType()
			if symbols[0].Name != id { // not exact match, cannot narrow more
				break
			}
		}

		if symbols, found = a.builtinTypeMembers(t); !found {
			break
		} else {
			identifiers = identifiers[1:]
		}
	}

	return symbols, identifiers
}

func docAndScopeSymbols(doc document.Document, node *sitter.Node) (symbols []query.Symbol) {
	if node != nil {
		symbols = query.SymbolsInScope(doc, node)
	}
	// REVIEW: will local symbols override global ones, since they appear in list earlier?
	// should we exclude same named symbols similarly to how it's done in availableSymbols?
	symbols = append(symbols, doc.Symbols()...)
	return symbols
}

// Returns a list of available symbols for completion as follows:
//   - If in a function argument list, include keyword args for that function
//   - Add symbols in scope for the node at point, excluding symbols at the module
//     level (document symbols), because the document already has those computed
//   - Add document symbols
//   - Add builtins
func (a *Analyzer) availableSymbols(doc document.Document, nodeAtPoint *sitter.Node, pt sitter.Point) []query.Symbol {
	symbols := []query.Symbol{}
	if nodeAtPoint != nil {
		if args := keywordArgContext(doc, nodeAtPoint, pt); args.fnName != "" {
			if fn, ok := a.signatureInformation(doc, nodeAtPoint, args); ok {
				symbols = append(symbols, a.keywordArgSymbols(fn, args)...)
			}
		}
		symbols = append(symbols, query.SymbolsInScope(doc, nodeAtPoint)...)
	}
	docAndBuiltin := append(doc.Symbols(), a.builtins.Symbols...)
	for _, sym := range docAndBuiltin {
		found := false
		for _, s := range symbols {
			if sym.Name == s.Name {
				found = true
				break
			}
		}
		if !found {
			symbols = append(symbols, sym)
		}
	}

	return symbols
}

func (a *Analyzer) nodesAtPointForCompletion(doc document.Document, pt sitter.Point) ([]*sitter.Node, bool) {
	node, ok := query.NodeAtPoint(doc, pt)
	if !ok {
		return []*sitter.Node{}, false
	}
	a.logger.Debug("node at point", zap.String("node", node.Type()), zap.String("content", doc.Content(node)))
	return a.nodesForCompletion(doc, node, pt)
}

// Zoom in or out from the node to include adjacent attribute expressions, so we can
// complete starting from the top-most attribute expression.
//
// TODO: rewrite findObjectExpression or nodesForCompletion to handle uniformily Erorr, Identifier and Attribute
// Right now nodesForCompletion returns either identifier and dot nodes or single attribue node (which should be flatten)
func (a *Analyzer) nodesForCompletion(doc document.Document, node *sitter.Node, pt sitter.Point) ([]*sitter.Node, bool) {
	nodes := []*sitter.Node{}
	switch node.Type() {
	case query.NodeTypeString, query.NodeTypeComment:
		if query.PointCovered(pt, node) {
			// No completion inside a string or comment
			return nodes, false
		}
	case query.NodeTypeModule, query.NodeTypeBlock:
		// Sometimes the top-level module is the most granular node due to
		// location of the point being between children, in this case, advance
		// to the first child node that appears after the point
		if node.NamedChildCount() > 0 {
			for node = node.NamedChild(0); node != nil && query.PointBefore(node.StartPoint(), pt); {
				next := node.NextNamedSibling()
				if next == nil {
					break
				}
				node = next
			}
			return a.nodesForCompletion(doc, node, pt)
		}

	case query.NodeTypeIfStatement,
		query.NodeTypeExpressionStatement,
		query.NodeTypeForStatement,
		query.NodeTypeAssignment:
		if node.NamedChildCount() == 1 {
			return a.nodesForCompletion(doc, node.NamedChild(0), pt)
		}

		for i := 0; i < int(node.NamedChildCount()); i++ {
			child := node.NamedChild(i)
			if query.PointBefore(child.EndPoint(), pt) {
				return a.leafNodesForCompletion(doc, child, pt)
			}
		}

	case query.NodeTypeAttribute, query.NodeTypeIdentifier:
		// If inside an attribute expression, capture the larger expression for completion.
		if node.Parent().Type() == query.NodeTypeAttribute {
			nodes, _ = a.nodesForCompletion(doc, node.Parent(), pt)
		} // TODO: flatten attribute to identifiers and dots to unify completion nodes

	case query.NodeTypeERROR, query.NodeTypeCall:
		leafNodes, ok := a.leafNodesForCompletion2(doc, node, pt)
		if len(leafNodes) > 0 {
			return leafNodes, ok
		}
		node = node.Child(int(node.ChildCount()) - 1)

	case query.NodeTypeArgList:
		leafNodes, ok := a.leafNodesForCompletion(doc, node, pt)
		if len(leafNodes) > 0 {
			return leafNodes, ok
		}
		node = node.Child(int(node.ChildCount()) - 1)
	}

	if len(nodes) == 0 {
		nodes = append(nodes, node)
	}
	return nodes, true
}

// Look at all leaf nodes for the node and its previous sibling in a
// flattened slice, in order of appearance. Take all consecutive trailing
// identifiers or '.' as the attribute expression to complete.
func (a *Analyzer) leafNodesForCompletion(doc document.Document, node *sitter.Node, pt sitter.Point) ([]*sitter.Node, bool) {
	leafNodes := []*sitter.Node{}

	if node.PrevNamedSibling() != nil {
		leafNodes = append(leafNodes, query.LeafNodes(node.PrevNamedSibling())...)
	}

	leafNodes = append(leafNodes, query.LeafNodes(node)...)

	// count number of trailing id/'.' nodes, if any
	trailingCount := 0
	leafCount := len(leafNodes)
	for i := 0; i < leafCount && i == trailingCount; i++ {
		switch leafNodes[leafCount-1-i].Type() {
		case query.NodeTypeIdentifier, ".":
			trailingCount++
		}
	}
	nodes := make([]*sitter.Node, trailingCount)
	for j := 0; j < len(nodes); j++ {
		nodes[j] = leafNodes[leafCount-trailingCount+j]
	}

	return nodes, true
}

// FIXME(s):
// - unify with leafNodesForCompletion. Right now just cut-n-paste with small changes.
// - do it with a query.
// - do we need prevNamedSibling? do we need argument_list?
// - then use it to extract identifiers and dot nodes from attributes as well
func (a *Analyzer) leafNodesForCompletion2(doc document.Document, node *sitter.Node, pt sitter.Point) ([]*sitter.Node, bool) {
	leafNodes := []*sitter.Node{}

	skipMap := map[string]string{"(": ")", "[": "]", "{": "}", "\"": "\""}

	leafNodes = append(leafNodes, query.LeafNodes(node)...)

	leafNodesFiltered := []*sitter.Node{}
	skip, found := false, false
	skipEnd := ""
	for _, n := range leafNodes {
		if !skip {
			if skipEnd, found = skipMap[n.Type()]; found {
				skip = true
				continue
			}
			leafNodesFiltered = append(leafNodesFiltered, n)
		}
		if n.Type() == skipEnd {
			skip = false
		}
	}

	// count number of trailing id/'.' nodes, if any
	countIdAndDots := func(nodes []*sitter.Node) int {
		leafCount := len(nodes)
		trailingCount := 0
		for i := 0; i < leafCount && i == trailingCount; i++ {
			switch nodes[leafCount-1-i].Type() {
			case query.NodeTypeIdentifier, query.NodeTypeDot,
				query.NodeTypeList, query.NodeTypeDictionary, query.NodeTypeString:
				trailingCount++
			}
		}
		return trailingCount
	}
	trailingCount := countIdAndDots(leafNodes)
	trailingCountF := countIdAndDots(leafNodesFiltered)
	if trailingCountF > trailingCount {
		leafNodes = leafNodesFiltered
		trailingCount = trailingCountF
	}

	nodes := make([]*sitter.Node, trailingCount)
	for j := 0; j < len(nodes); j++ {
		nodes[j] = leafNodes[len(leafNodes)-trailingCount+j]
	}

	return nodes, true
}

func (a *Analyzer) keywordArgSymbols(fn query.Signature, args callWithArguments) []query.Symbol {
	symbols := []query.Symbol{}
	for i, param := range fn.Params {
		if i < int(args.positional) {
			continue
		}
		kwarg := param.Name
		if used := args.keywords[kwarg]; !used {
			symbols = append(symbols, query.Symbol{
				Name:   kwarg + "=",
				Detail: param.Content,
				Kind:   protocol.SymbolKindVariable,
			})
		}
	}
	return symbols
}

// Find the object part of an expression that has a dot '.' immediately before the given point.
func (a *Analyzer) findObjectExpression(nodes []*sitter.Node, pt sitter.Point) *sitter.Node {
	// There could be multiple cases:
	// 1. attribute node with 3 kids (identifier, dot, identifier), e.g. `bar.foo()` via hoover
	// 2. two nodes (identifier, dot) w/o kids, with common parent, e.g. `bar.` via completion
	// 3. two nodes (identifier, dot) w/o kids, with different parents, e.g. `baz(bar.)`` via completion

	if pt.Column == 0 {
		return nil
	}

	var dot, expr, parentNode *sitter.Node
	searchRange := sitter.Range{StartPoint: sitter.Point{Row: pt.Row, Column: pt.Column - 1}, EndPoint: pt}
	nodeComparisonFunc := func(n *sitter.Node) int {
		if query.PointBeforeOrEqual(n.EndPoint(), searchRange.StartPoint) {
			return -1
		}
		if n.StartPoint() == searchRange.StartPoint &&
			n.EndPoint() == searchRange.EndPoint &&
			n.Type() == "." {
			return 0
		}
		if query.PointBeforeOrEqual(n.StartPoint(), searchRange.StartPoint) &&
			query.PointAfterOrEqual(n.EndPoint(), searchRange.EndPoint) {
			return 0
		}
		return 1
	}

	// first search in children. For example attribute node with 3 kids (identifier, dot, identifier)
	for i := len(nodes) - 1; i >= 0; i-- {
		parentNode = nodes[i]
		if dot = query.FindChildNode(parentNode, nodeComparisonFunc); dot != nil {
			break
		}
	}

	if dot != nil {
		expr = parentNode.PrevSibling()
		for n := dot; n != parentNode; n = n.Parent() {
			if n.PrevSibling() != nil {
				expr = n.PrevSibling()
				break
			}
		}
	}

	// if not found, check nodes themselves
	if dot == nil && expr == nil {
		for i := len(nodes) - 1; i > 0; i-- {
			if nodeComparisonFunc(nodes[i]) == 0 {
				dot = nodes[i]
				expr = nodes[i-1]
			}
		}
	}

	if expr != nil {
		a.logger.Debug("dot completion",
			zap.String("dot", dot.String()),
			zap.String("expr", expr.String()))
	}
	return expr
}

func (a *Analyzer) resolveNodeWithSymbol(doc document.Document, node *sitter.Node, symbol string) query.Symbol {
	if sym, found := a.FindDefinition(doc, node, symbol); found {
		// TODO: handle better symbols. we already compute symbols in findDefinition
		if sym.Kind == protocol.SymbolKindVariable {
			symbols := gAvailableSymbols
			if len(symbols) == 0 { // no cache
				symbols = docAndScopeSymbols(doc, node) // to resolve variable we need only doc and scope
			}
			identifiers := a.resolveSymbolIdentifiers(symbols, sym)
			a.logger.Debug("resolved node identifiers", zap.String("node", doc.Content(node)), zap.Strings("identifiers", identifiers))
			if symbols, ok := a.builtinsCompletion(doc, identifiers); ok && len(symbols) == 1 {
				return symbols[0]
			}
		} else if sym.Kind == protocol.SymbolKindFunction {
			if symbols, ok := a.builtinsCompletion(doc, []string{symbol}); ok && len(symbols) == 1 {
				return symbols[0]
			}
		} else if t := query.SymbolKindToBuiltinType(sym.Kind); t != "" {
			return query.Symbol{Name: symbol, Type: t}
		}
	}
	return query.Symbol{}
}

func (a *Analyzer) resolveNode(doc document.Document, node *sitter.Node) query.Symbol {
	parts := strings.Split(doc.Content(node), ".") // extract first part, e.g. from attribute node
	switch node.Type() {
	case query.NodeTypeString, query.NodeTypeDictionary, query.NodeTypeList:
		return query.Symbol{Name: parts[0], Type: query.SymbolKindToBuiltinType(query.StrToSymbolKind(node.Type()))}

	case query.NodeTypeIdentifier, query.NodeTypeAttribute:
		return a.resolveNodeWithSymbol(doc, node, parts[0])
	}
	return query.Symbol{}
}

// Perform some rudimentary type analysis to determine the Starlark type of the node
func (a *Analyzer) analyzeType(doc document.Document, node *sitter.Node) string {
	if node == nil {
		return ""
	}
	nodeT := node.Type()
	switch nodeT {
	case query.NodeTypeString, query.NodeTypeDictionary, query.NodeTypeList:
		return query.SymbolKindToBuiltinType(query.StrToSymbolKind(nodeT))

	case query.NodeTypeIdentifier:
		return a.resolveNodeWithSymbol(doc, node, doc.Content(node)).GetType()

	case query.NodeTypeAttribute:
		return a.resolveNodeWithSymbol(doc, node, strings.Split(doc.Content(node), ".")[0]).GetType()

	case query.NodeTypeCall:
		fnName := doc.Content(node.ChildByFieldName("function"))
		args := node.ChildByFieldName("arguments")
		sig, _ := a.signatureInformation(doc, node, callWithArguments{fnName: fnName, argsNode: args})
		_, t := query.StrToSymbolKindAndType(sig.ReturnType)
		return t
	}
	return ""
}

// traverse provided symbols and all resolve given symbol iteratively to the final list of identifiers
// e.g. r1 = foo().bar r2=r1.baz().q.w.e r=r1.r.t.y, should be resolved to [foo, bar, baz, q, w, e, t, y]
func (a *Analyzer) resolveSymbolIdentifiers(symbols []query.Symbol, sym query.Symbol) []string {
	resolvedType := sym.Name
	origSymName := sym.Name

	maxResolveSteps := 5 // just to limit
	for i := 0; i < maxResolveSteps; i++ {

		t := sym.GetType()
		if t == "" {
			break
		}

		resolvedType = strings.TrimPrefix(resolvedType, sym.Name)
		resolvedType = t + resolvedType

		if _, knownKind := query.KnownSymbolKinds[sym.Kind]; knownKind {
			break
		}

		parts := strings.SplitN(t, ".", 2)
		t = parts[0] // take first part and use it as a new type
		sym = query.Symbol{}
		for _, s := range symbols {
			if s.Name == t {
				sym = s
				break
			}
		}
	}

	identifiers := strings.Split(removeBrackets(resolvedType), ".")
	replaceKnownTypes(identifiers)
	a.logger.Debug("resolve symbol identifiers", zap.String("sym", origSymName), zap.Strings("identifiers", identifiers))
	return identifiers
}

// returns available members for a given type
func (a *Analyzer) builtinTypeMembers(t string) ([]query.Symbol, bool) {
	if t != "" {
		if class, found := a.builtins.Types[t]; found {
			return class.Members, true
		}
		switch t {
		case "None", "bool", "int", "float":
			return []query.Symbol{}, true
		}
		return []query.Symbol{}, false
	} else {
		return a.builtins.Members, false // REVIEW: return everything or nothing?
	}
}

func (a *Analyzer) availableMembersForNode(doc document.Document, node *sitter.Node) []query.Symbol {
	t := a.analyzeType(doc, node)
	members, _ := a.builtinTypeMembers(t)
	return members
}

func (a *Analyzer) FindDefinition(doc document.Document, node *sitter.Node, name string) (query.Symbol, bool) {
	if node != nil {
		for _, sym := range query.SymbolsInScope(doc, node) {
			if sym.Name == name {
				return sym, true
			}
		}
	}
	for _, sym := range doc.Symbols() {
		if sym.Name == name {
			return sym, true
		}
	}
	for _, sym := range a.builtins.Symbols {
		if sym.Name == name {
			return sym, true
		}
	}
	return query.Symbol{}, false
}

func keywordArgContext(doc document.Document, node *sitter.Node, pt sitter.Point) callWithArguments {
	if node.Type() == "=" ||
		query.HasAncestor(node, func(anc *sitter.Node) bool {
			return anc.Type() == query.NodeTypeKeywordArgument
		}) {
		return callWithArguments{}
	}
	return possibleCallInfo(doc, node, pt)
}
