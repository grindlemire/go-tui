package gopls

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/grindlemire/go-tui/internal/lsp/log"
	"github.com/grindlemire/go-tui/internal/tuigen"
)

// stateNewStateRegex matches state declarations like: count := tui.NewState(0)
var stateNewStateRegex = regexp.MustCompile(`(\w+)\s*:=\s*tui\.NewState\((.+)\)`)

// GenerateVirtualGo generates a valid Go source file from a .gsx AST.
// It returns the generated source and a SourceMap for position translation.
//
// Position convention: tuigen positions are 1-indexed (Line and Column both start at 1).
// All TuiLine/TuiCol values in source mappings are 0-indexed. The conversion depends on
// the node type:
//   - GoExpr/RawGoExpr: Position.Column points to the '{' delimiter. Expression content
//     starts at Column+1 (1-indexed) = Column (0-indexed). So: tuiCol = Position.Column.
//   - GoCode/ForLoop/IfStmt/LetBinding: Position.Column points to the code/keyword itself.
//     So: tuiCol = Position.Column - 1.
func GenerateVirtualGo(file *tuigen.File) (string, *SourceMap) {
	g := &generator{
		sourceMap: NewSourceMap(),
	}
	return g.generate(file), g.sourceMap
}

// generator holds state during Go code generation.
type generator struct {
	buf       strings.Builder
	sourceMap *SourceMap
	goLine    int // current line in generated Go (0-indexed)
	goCol     int // current column in generated Go (0-indexed)
}

// generate produces Go source from a .gsx AST.
func (g *generator) generate(file *tuigen.File) string {
	// Package declaration
	g.writeLine(fmt.Sprintf("package %s", file.Package))
	g.writeLine("")

	// Imports
	if len(file.Imports) > 0 {
		if len(file.Imports) == 1 {
			imp := file.Imports[0]
			if imp.Alias != "" {
				g.writeLine(fmt.Sprintf("import %s %q", imp.Alias, imp.Path))
			} else {
				g.writeLine(fmt.Sprintf("import %q", imp.Path))
			}
		} else {
			g.writeLine("import (")
			for _, imp := range file.Imports {
				if imp.Alias != "" {
					g.writeLine(fmt.Sprintf("\t%s %q", imp.Alias, imp.Path))
				} else {
					g.writeLine(fmt.Sprintf("\t%q", imp.Path))
				}
			}
			g.writeLine(")")
		}
		g.writeLine("")
	}

	// Generate a dummy function for each component to hold Go expressions
	for _, comp := range file.Components {
		g.generateComponent(comp)
		g.writeLine("")
	}

	// Generate top-level functions with source mapping
	for _, fn := range file.Funcs {
		g.generateFunc(fn)
		g.writeLine("")
	}

	return g.buf.String()
}

// generateComponent generates Go code for a component.
func (g *generator) generateComponent(comp *tuigen.Component) {
	// Build parameter list and track positions
	var params []string
	for _, p := range comp.Params {
		params = append(params, fmt.Sprintf("%s %s", p.Name, p.Type))
	}

	// Function signature
	returnType := comp.ReturnType
	if returnType == "" {
		returnType = "*element.Element"
	}

	// Calculate the position where params start in the Go line
	// Format: "func Name(" then params
	sigPrefix := fmt.Sprintf("func %s(", comp.Name)
	goParamStartCol := len(sigPrefix)

	// Add mappings for each parameter
	for _, p := range comp.Params {
		if p.Position.Line > 0 && p.Position.Column > 0 {
			// Map the full parameter (name + space + type) from .gsx to .go
			// so gopls can resolve types in component signatures
			m := Mapping{
				TuiLine: p.Position.Line - 1,
				TuiCol:  p.Position.Column - 1,
				GoLine:  g.goLine,
				GoCol:   goParamStartCol,
				Length:  len(p.Name) + 1 + len(p.Type),
			}
			log.Generate("PARAM mapping: %s -> TuiLine=%d TuiCol=%d GoLine=%d GoCol=%d Len=%d (pos.Line=%d pos.Col=%d)",
				p.Name, m.TuiLine, m.TuiCol, m.GoLine, m.GoCol, m.Length, p.Position.Line, p.Position.Column)
			g.sourceMap.AddMapping(m)
		}
		// Move past this param: "name type, "
		goParamStartCol += len(p.Name) + 1 + len(p.Type) + 2 // +1 for space, +2 for ", "
	}

	g.writeLine(fmt.Sprintf("func %s(%s) %s {", comp.Name, strings.Join(params, ", "), returnType))

	// Emit state variable declarations so gopls understands state types.
	// Scans component body GoCode nodes for tui.NewState(...) patterns and
	// emits corresponding Go declarations with source mappings.
	g.emitStateVarDeclarations(comp)

	// Emit named ref variable declarations so gopls understands ref types.
	// Scans component body for elements with #Name refs and emits
	// var declarations with the appropriate type.
	g.emitNamedRefDeclarations(comp)

	// Generate dummy assignments for all Go expressions in the body
	g.generateNodes(comp.Body, 1)

	// Return nil to make the function valid
	g.writeLine("\treturn nil")
	g.writeLine("}")
}

// generateNodes generates Go code for a list of nodes.
func (g *generator) generateNodes(nodes []tuigen.Node, indent int) {
	indentStr := strings.Repeat("\t", indent)
	for _, node := range nodes {
		g.generateNode(node, indentStr)
	}
}

// generateNode generates Go code for a single node.
func (g *generator) generateNode(node tuigen.Node, indent string) {
	switch n := node.(type) {
	case *tuigen.Element:
		g.generateElement(n, indent)
	case *tuigen.GoExpr:
		g.generateGoExpr(n, indent)
	case *tuigen.ForLoop:
		g.generateForLoop(n, indent)
	case *tuigen.IfStmt:
		g.generateIfStmt(n, indent)
	case *tuigen.LetBinding:
		g.generateLetBinding(n, indent)
	case *tuigen.ComponentCall:
		g.generateComponentCall(n, indent)
	case *tuigen.TextContent:
		// Ignore plain text
	case *tuigen.GoCode:
		g.generateGoCode(n, indent)
	case *tuigen.RawGoExpr:
		g.generateRawGoExpr(n, indent)
	case *tuigen.ChildrenSlot:
		// Ignore children slot
	}
}

// generateElement generates Go code for an element.
func (g *generator) generateElement(el *tuigen.Element, indent string) {
	if el == nil {
		return
	}
	// Generate attribute expressions
	for _, attr := range el.Attributes {
		if expr, ok := attr.Value.(*tuigen.GoExpr); ok {
			g.generateGoExpr(expr, indent)
		}
	}

	// Generate children
	for _, child := range el.Children {
		g.generateNode(child, indent)
	}
}

// generateGoExpr generates a Go expression and records the position mapping.
func (g *generator) generateGoExpr(expr *tuigen.GoExpr, indent string) {
	if expr == nil {
		return
	}
	// Clean up the expression
	code := strings.TrimSpace(expr.Code)
	if code == "" {
		return
	}

	// Record position mapping before writing
	// The expression starts after "_ = " (4 characters + indent)
	tuiLine := expr.Position.Line - 1 // convert to 0-indexed
	// Position.Column is 1-indexed and points to the '{' delimiter.
	// The expression content starts at Column+1 (1-indexed) = Column (0-indexed).
	tuiCol := expr.Position.Column

	goExprStartCol := len(indent) + 4 // "_ = " is 4 chars

	m := Mapping{
		TuiLine: tuiLine,
		TuiCol:  tuiCol,
		GoLine:  g.goLine,
		GoCol:   goExprStartCol,
		Length:  len(code),
	}
	log.Generate("GOEXPR mapping: '%s' -> TuiLine=%d TuiCol=%d GoLine=%d GoCol=%d Len=%d (pos.Line=%d pos.Col=%d)",
		code, m.TuiLine, m.TuiCol, m.GoLine, m.GoCol, m.Length, expr.Position.Line, expr.Position.Column)
	g.sourceMap.AddMapping(m)

	// Write dummy assignment
	g.writeLine(fmt.Sprintf("%s_ = %s", indent, code))
}

// generateGoCode generates Go code for a GoCode node (non-tui.NewState code).
// tui.NewState declarations are handled separately by emitStateVarDeclarations.
func (g *generator) generateGoCode(gc *tuigen.GoCode, indent string) {
	if gc == nil {
		return
	}
	code := strings.TrimSpace(gc.Code)
	if code == "" {
		return
	}
	// Skip tui.NewState declarations — already handled by emitStateVarDeclarations
	if stateNewStateRegex.MatchString(code) {
		return
	}

	tuiLine := gc.Position.Line - 1
	// GoCode Position.Column is 1-indexed and points directly to the code content
	// (no delimiter), so subtract 1 to convert to 0-indexed.
	tuiCol := gc.Position.Column - 1

	g.sourceMap.AddMapping(Mapping{
		TuiLine: tuiLine,
		TuiCol:  tuiCol,
		GoLine:  g.goLine,
		GoCol:   len(indent),
		Length:  len(code),
	})

	g.writeLine(fmt.Sprintf("%s%s", indent, code))
}

// generateRawGoExpr generates a raw Go expression.
func (g *generator) generateRawGoExpr(expr *tuigen.RawGoExpr, indent string) {
	if expr == nil {
		return
	}
	code := strings.TrimSpace(expr.Code)
	if code == "" {
		return
	}

	tuiLine := expr.Position.Line - 1
	// RawGoExpr: same convention as GoExpr — Position.Column is 1-indexed and
	// points to the '{' delimiter, so Column (0-indexed) = expression start.
	tuiCol := expr.Position.Column

	goExprStartCol := len(indent) + 4

	g.sourceMap.AddMapping(Mapping{
		TuiLine: tuiLine,
		TuiCol:  tuiCol,
		GoLine:  g.goLine,
		GoCol:   goExprStartCol,
		Length:  len(code),
	})

	g.writeLine(fmt.Sprintf("%s_ = %s", indent, code))
}

// generateForLoop generates Go code for a for loop.
func (g *generator) generateForLoop(loop *tuigen.ForLoop, indent string) {
	if loop == nil {
		return
	}
	// Record mapping for the iterable expression
	// Position.Column is 1-indexed from parser, convert to 0-indexed then add offset
	tuiLine := loop.Position.Line - 1
	// The iterable starts after "@for index, value := range "
	tuiCol := loop.Position.Column - 1 + len("@for ") + len(loop.Index) + len(", ") + len(loop.Value) + len(" := range ")

	// Generate for loop header
	indexVar := loop.Index
	if indexVar == "" {
		indexVar = "_"
	}

	goExprStartCol := len(indent) + len("for ") + len(indexVar) + len(", ") + len(loop.Value) + len(" := range ")

	m := Mapping{
		TuiLine: tuiLine,
		TuiCol:  tuiCol,
		GoLine:  g.goLine,
		GoCol:   goExprStartCol,
		Length:  len(loop.Iterable),
	}
	log.Generate("FOR mapping: iterable='%s' -> TuiLine=%d TuiCol=%d GoLine=%d GoCol=%d Len=%d (pos.Line=%d pos.Col=%d)",
		loop.Iterable, m.TuiLine, m.TuiCol, m.GoLine, m.GoCol, m.Length, loop.Position.Line, loop.Position.Column)
	g.sourceMap.AddMapping(m)

	g.writeLine(fmt.Sprintf("%sfor %s, %s := range %s {", indent, indexVar, loop.Value, loop.Iterable))

	// Generate body
	g.generateNodes(loop.Body, len(indent)+1)

	g.writeLine(fmt.Sprintf("%s}", indent))
}

// generateIfStmt generates Go code for an if statement.
func (g *generator) generateIfStmt(stmt *tuigen.IfStmt, indent string) {
	if stmt == nil {
		return
	}
	// Record mapping for the condition
	// Position.Column is 1-indexed from parser, convert to 0-indexed then add offset
	tuiLine := stmt.Position.Line - 1
	tuiCol := stmt.Position.Column - 1 + len("@if ") // -1 converts to 0-indexed

	goExprStartCol := len(indent) + len("if ")

	m := Mapping{
		TuiLine: tuiLine,
		TuiCol:  tuiCol,
		GoLine:  g.goLine,
		GoCol:   goExprStartCol,
		Length:  len(stmt.Condition),
	}
	log.Generate("IF mapping: condition='%s' -> TuiLine=%d TuiCol=%d GoLine=%d GoCol=%d Len=%d (pos.Line=%d pos.Col=%d, indent='%s')",
		stmt.Condition, m.TuiLine, m.TuiCol, m.GoLine, m.GoCol, m.Length, stmt.Position.Line, stmt.Position.Column, indent)
	g.sourceMap.AddMapping(m)

	g.writeLine(fmt.Sprintf("%sif %s {", indent, stmt.Condition))

	// Generate then branch
	g.generateNodes(stmt.Then, len(indent)+1)

	if len(stmt.Else) > 0 {
		g.writeLine(fmt.Sprintf("%s} else {", indent))
		g.generateNodes(stmt.Else, len(indent)+1)
	}

	g.writeLine(fmt.Sprintf("%s}", indent))
}

// generateLetBinding generates Go code for a let binding.
func (g *generator) generateLetBinding(binding *tuigen.LetBinding, indent string) {
	if binding == nil {
		return
	}

	// Add mapping for the variable name
	// In .gsx: "@let varName = ..." - Position points to @, so varName is at Column + len("@let ")
	// In .go: "var varName interface{}" - varName is at indent + len("var ")
	tuiLine := binding.Position.Line - 1
	tuiCol := binding.Position.Column - 1 + len("@let ")
	goVarStartCol := len(indent) + len("var ")

	m := Mapping{
		TuiLine: tuiLine,
		TuiCol:  tuiCol,
		GoLine:  g.goLine,
		GoCol:   goVarStartCol,
		Length:  len(binding.Name),
	}
	log.Generate("LET mapping: '%s' -> TuiLine=%d TuiCol=%d GoLine=%d GoCol=%d Len=%d (pos.Line=%d pos.Col=%d)",
		binding.Name, m.TuiLine, m.TuiCol, m.GoLine, m.GoCol, m.Length, binding.Position.Line, binding.Position.Column)
	g.sourceMap.AddMapping(m)

	// Generate variable declaration
	g.writeLine(fmt.Sprintf("%svar %s interface{}", indent, binding.Name))

	// Generate element expressions
	if binding.Element != nil {
		g.generateElement(binding.Element, indent)
	}
}

// generateComponentCall generates Go code for a component call.
func (g *generator) generateComponentCall(call *tuigen.ComponentCall, indent string) {
	if call == nil {
		return
	}
	// Record mapping for the arguments
	// Position.Column is 1-indexed from parser, convert to 0-indexed then add offset
	if call.Args != "" {
		tuiLine := call.Position.Line - 1
		tuiCol := call.Position.Column - 1 + len("@") + len(call.Name) + 1 // -1 for 0-index, +1 for (

		goExprStartCol := len(indent) + 4 + len(call.Name) + 1 // "_ = " + name + "("

		g.sourceMap.AddMapping(Mapping{
			TuiLine: tuiLine,
			TuiCol:  tuiCol,
			GoLine:  g.goLine,
			GoCol:   goExprStartCol,
			Length:  len(call.Args),
		})
	}

	// Generate dummy call
	g.writeLine(fmt.Sprintf("%s_ = %s(%s)", indent, call.Name, call.Args))

	// Generate children
	g.generateNodes(call.Children, len(indent)+1)
}

// generateFunc generates a top-level function with source mapping.
func (g *generator) generateFunc(fn *tuigen.GoFunc) {
	code := fn.Code
	lines := strings.Split(code, "\n")

	tuiStartLine := fn.Position.Line - 1 // convert to 0-indexed
	tuiStartCol := fn.Position.Column - 1

	for i, line := range lines {
		// Add mapping for each line of the function
		// This maps the entire line from .gsx to .go
		g.sourceMap.AddMapping(Mapping{
			TuiLine: tuiStartLine + i,
			TuiCol:  tuiStartCol,
			GoLine:  g.goLine,
			GoCol:   0,
			Length:  len(line),
		})
		g.writeLine(line)
	}
}

// writeLine writes a line and updates line tracking.
func (g *generator) writeLine(line string) {
	g.buf.WriteString(line)
	g.buf.WriteString("\n")
	g.goLine++
	g.goCol = 0
}
