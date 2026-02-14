package tuigen

import (
	"bytes"
	"fmt"
	"go/format"

	"golang.org/x/tools/imports"
)

// Generator transforms a validated AST into Go source code.
type Generator struct {
	buf        bytes.Buffer
	indent     int
	varCounter int
	sourceFile string // original .gsx filename for header comment

	// Refs tracking for current component
	refs []RefInfo

	// Component calls with watchers that need aggregation
	componentVars []string

	// State tracking for current component (for reactive bindings)
	stateVars     []StateVar
	stateBindings []StateBinding

	// Events tracking for current component (for BindApp generation)
	eventsVars []EventsVar

	// Conditional counter for reactive @if wrapper elements (__cond_0, __cond_1, etc.)
	condCounter int

	// Loop counter for reactive @for wrapper elements (__loop_0, __loop_1, etc.)
	loopCounter int

	// Mount index counter for struct component @Component() calls in method templs.
	// Reset per method component. Assigns position indices to app.Mount() calls.
	mountIndex int

	// currentReceiver is the receiver variable name for the current method templ.
	// Set during generateMethodComponent, used by generateComponentCallWithRefs
	// to emit app.Mount(receiverVar, index, factory).
	currentReceiver string

	// loopIndexStack tracks loop index variable names for nested @for loops.
	// Used by generateStructMount to generate unique mount indices per iteration.
	loopIndexStack []string

	// fileDecls stores the current file's GoDecl nodes for struct lookup.
	// Used by generateUpdateProps to find struct definitions for method components.
	fileDecls []*GoDecl

	// fileFuncs stores the current file's top-level Go functions/methods.
	// Used to detect user-defined BindApp methods and avoid duplicate generation.
	fileFuncs []*GoFunc

	// SkipImports uses format.Source instead of imports.Process (faster for tests)
	SkipImports bool

	// Source map tracking
	sourceMap   *SourceMap
	currentLine int // current line in generated output (0-indexed)
}

// NewGenerator creates a new code generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate produces Go source code from a parsed and analyzed AST.
// Returns the generated code as a byte slice, or an error if generation fails.
func (g *Generator) Generate(file *File, sourceFile string) ([]byte, error) {
	g.buf.Reset()
	g.varCounter = 0
	g.sourceFile = sourceFile
	g.sourceMap = NewSourceMap(sourceFile)
	g.currentLine = 0

	// Generate header
	g.generateHeader()

	// Generate package
	g.generatePackage(file.Package)

	// Generate imports
	g.generateImports(file.Imports)

	// Track where content after imports starts (for source map adjustment)
	firstContentLine := g.currentLine

	// Store file decls for struct lookup in generateUpdateProps
	g.fileDecls = file.Decls
	g.fileFuncs = file.Funcs

	// Generate top-level Go declarations (type, const, var)
	for _, decl := range file.Decls {
		g.generateGoDecl(decl)
	}

	// Generate top-level Go functions
	for _, fn := range file.Funcs {
		g.generateGoFunc(fn)
	}

	// Generate components
	for _, comp := range file.Components {
		g.generateComponent(comp)
	}

	// For tests: just format without import processing (much faster)
	if g.SkipImports {
		return format.Source(g.buf.Bytes())
	}

	// For production: format and fix imports with goimports
	preOutput := g.buf.Bytes()
	postOutput, err := imports.Process(g.sourceFile, preOutput, nil)
	if err != nil {
		return nil, err
	}

	// Adjust source map for line shifts caused by goimports.
	// Goimports only modifies the import section, so we calculate the shift
	// by comparing line counts before and after.
	g.adjustSourceMapForGoimports(preOutput, postOutput, firstContentLine)

	return postOutput, nil
}

// adjustSourceMapForGoimports adjusts source map line numbers after goimports
// modifies the import section. We find where imports end in both versions
// and calculate the shift from that, since content after imports shifts
// by the difference in import section sizes.
func (g *Generator) adjustSourceMapForGoimports(pre, post []byte, firstContentLine int) {
	// Find where content starts in the post-goimports output.
	// This is the first non-blank line after the import block.
	postContentStart := findFirstContentLineAfterImports(post)

	// The shift is how much the content start moved
	lineShift := postContentStart - firstContentLine

	if lineShift == 0 {
		return
	}

	// Adjust all source map entries for lines at or after where content starts
	for i := range g.sourceMap.Mappings {
		if g.sourceMap.Mappings[i].GoLine >= firstContentLine {
			g.sourceMap.Mappings[i].GoLine += lineShift
		}
	}
}

// findFirstContentLineAfterImports finds the first non-blank line after the import block.
func findFirstContentLineAfterImports(code []byte) int {
	lines := bytes.Split(code, []byte("\n"))
	inImportBlock := false
	importBlockEnded := false

	for i, line := range lines {
		trimmed := bytes.TrimSpace(line)

		// Track import block
		if bytes.HasPrefix(trimmed, []byte("import (")) {
			inImportBlock = true
			continue
		}
		if inImportBlock && len(trimmed) == 1 && trimmed[0] == ')' {
			inImportBlock = false
			importBlockEnded = true
			continue
		}
		// Handle single-line imports: import "path"
		if bytes.HasPrefix(trimmed, []byte("import ")) && !bytes.HasPrefix(trimmed, []byte("import (")) {
			importBlockEnded = true
			continue
		}

		// After imports end, find first non-blank line
		if importBlockEnded && len(trimmed) > 0 {
			return i
		}
	}

	return len(lines) // Fallback if no content found
}

// GetSourceMap returns the source map generated during code generation.
// Must be called after Generate().
func (g *Generator) GetSourceMap() *SourceMap {
	return g.sourceMap
}

// generateHeader writes the "DO NOT EDIT" comment.
func (g *Generator) generateHeader() {
	g.writeln("// Code generated by tui generate. DO NOT EDIT.")
	if g.sourceFile != "" {
		g.writef("// Source: %s\n", g.sourceFile)
	}
	g.writeln("")
}

// generatePackage writes the package declaration.
func (g *Generator) generatePackage(pkg string) {
	g.writef("package %s\n\n", pkg)
}

// generateImports writes the import block.
func (g *Generator) generateImports(imports []Import) {
	if len(imports) == 0 {
		// Always include root tui import for generated code
		g.writeln("import (")
		g.indent++
		g.writeln(`tui "github.com/grindlemire/go-tui"`)
		g.indent--
		g.writeln(")")
		g.writeln("")
		return
	}

	// Check if root tui package is already imported
	hasTui := false
	for _, imp := range imports {
		if imp.Path == "github.com/grindlemire/go-tui" {
			hasTui = true
		}
	}

	g.writeln("import (")
	g.indent++

	for _, imp := range imports {
		if imp.Alias != "" {
			g.writef("%s %q\n", imp.Alias, imp.Path)
		} else {
			g.writef("%q\n", imp.Path)
		}
	}

	// Add required import if not present
	if !hasTui {
		g.writeln("")
		g.writeln(`tui "github.com/grindlemire/go-tui"`)
	}

	g.indent--
	g.writeln(")")
	g.writeln("")
}

// nextVar returns the next unique variable name.
func (g *Generator) nextVar() string {
	name := fmt.Sprintf("__tui_%d", g.varCounter)
	g.varCounter++
	return name
}

// nextCondVar returns the next unique conditional wrapper variable name.
func (g *Generator) nextCondVar() string {
	name := fmt.Sprintf("__cond_%d", g.condCounter)
	g.condCounter++
	return name
}

// nextLoopVar returns the next unique loop wrapper variable name.
func (g *Generator) nextLoopVar() string {
	name := fmt.Sprintf("__loop_%d", g.loopCounter)
	g.loopCounter++
	return name
}

// pushLoopIndex adds a loop index variable to the stack and returns the variable name to use.
// If the loop has a usable index variable (not "" or "_"), uses it directly.
// Otherwise, generates a synthetic index variable name.
func (g *Generator) pushLoopIndex(loop *ForLoop) string {
	var idxVar string
	if loop.Index != "" && loop.Index != "_" {
		idxVar = loop.Index
	} else {
		idxVar = fmt.Sprintf("__idx_%d", len(g.loopIndexStack))
	}
	g.loopIndexStack = append(g.loopIndexStack, idxVar)
	return idxVar
}

// popLoopIndex removes the most recent loop index variable from the stack.
func (g *Generator) popLoopIndex() {
	if len(g.loopIndexStack) > 0 {
		g.loopIndexStack = g.loopIndexStack[:len(g.loopIndexStack)-1]
	}
}

// loopIndexExpr returns a Go expression for computing a unique mount index
// based on the current loop context. Returns empty string if not in a loop.
// The expression combines a static base index with runtime loop indices.
func (g *Generator) loopIndexExpr(baseIndex int) string {
	if len(g.loopIndexStack) == 0 {
		return ""
	}
	// For a single loop level: baseIndex*1000000 + loopIdx
	// For nested loops: combine all indices with large multipliers
	// This ensures unique keys for each (component call site, loop iteration) combination
	expr := fmt.Sprintf("%d", baseIndex)
	multiplier := 1000000
	for _, idxVar := range g.loopIndexStack {
		expr = fmt.Sprintf("%s*%d+%s", expr, multiplier, idxVar)
	}
	return expr
}

// stateNameSet returns a set of state variable names for quick lookup.
func (g *Generator) stateNameSet() map[string]bool {
	m := make(map[string]bool)
	for _, sv := range g.stateVars {
		m[sv.Name] = true
	}
	return m
}

// write writes a string without indentation and tracks line numbers.
func (g *Generator) write(s string) {
	g.buf.WriteString(s)
	// Count newlines in the output
	for _, c := range s {
		if c == '\n' {
			g.currentLine++
		}
	}
}

// writef writes a formatted string with indentation and tracks line numbers.
func (g *Generator) writef(format string, args ...interface{}) {
	g.writeIndent()
	s := fmt.Sprintf(format, args...)
	g.buf.WriteString(s)
	// Count newlines in the output
	for _, c := range s {
		if c == '\n' {
			g.currentLine++
		}
	}
}

// writeln writes a line with indentation and tracks line numbers.
func (g *Generator) writeln(s string) {
	if s == "" {
		g.buf.WriteByte('\n')
		g.currentLine++
		return
	}
	g.writeIndent()
	g.buf.WriteString(s)
	g.buf.WriteByte('\n')
	g.currentLine++
}

// writeIndent writes the current indentation.
func (g *Generator) writeIndent() {
	for i := 0; i < g.indent; i++ {
		g.buf.WriteByte('\t')
	}
}

// GenerateString is a convenience method that returns the generated code as a string.
func (g *Generator) GenerateString(file *File, sourceFile string) (string, error) {
	data, err := g.Generate(file, sourceFile)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParseAndGenerate parses source code and generates Go code in one step.
// This is a convenience function for simple use cases.
func ParseAndGenerate(filename, source string) ([]byte, error) {
	return parseAndGenerate(filename, source, false)
}

// parseAndGenerateSkipImports is like ParseAndGenerate but uses format.Source
// instead of imports.Process. This is much faster for tests.
func parseAndGenerateSkipImports(filename, source string) ([]byte, error) {
	return parseAndGenerate(filename, source, true)
}

func parseAndGenerate(filename, source string, skipImports bool) ([]byte, error) {
	lexer := NewLexer(filename, source)
	parser := NewParser(lexer)

	file, err := parser.ParseFile()
	if err != nil {
		return nil, err
	}

	gen := NewGenerator()
	gen.SkipImports = skipImports
	return gen.Generate(file, filename)
}

// GenerateToBuffer generates code and writes it to the buffer.
// This avoids an extra allocation compared to Generate().
func (g *Generator) GenerateToBuffer(buf *bytes.Buffer, file *File, sourceFile string) error {
	data, err := g.Generate(file, sourceFile)
	if err != nil {
		return err
	}
	buf.Write(data)
	return nil
}

// generateStateBindings generates Bind() calls for reactive state bindings.
// This is called after all elements are created so the element variables exist.
func (g *Generator) generateStateBindings() {
	if len(g.stateBindings) == 0 {
		return
	}

	g.writeln("")
	g.writeln("// State bindings")

	// Build a map of state variable names to their types
	stateTypes := make(map[string]string)
	for _, sv := range g.stateVars {
		stateTypes[sv.Name] = sv.Type
	}

	for _, binding := range g.stateBindings {
		g.generateBinding(binding, stateTypes)
	}
}

// generateBinding generates a Bind() call for a single state binding.
func (g *Generator) generateBinding(b StateBinding, stateTypes map[string]string) {
	if len(b.StateVars) == 0 {
		return
	}

	// Determine the setter method based on attribute
	setter := g.getSetterForAttribute(b.Attribute)
	if setter == "" {
		return
	}

	if len(b.StateVars) == 1 {
		// Single state variable - direct binding
		stateName := b.StateVars[0]
		stateType := stateTypes[stateName]
		g.writef("%s.Bind(func(_ %s) {\n", stateName, stateType)
		g.indent++
		g.writef("%s.%s(%s)\n", b.ElementName, setter, b.Expr)
		g.indent--
		g.writeln("})")
	} else {
		// Multiple state variables - shared update function
		updateFn := fmt.Sprintf("__update_%s", b.ElementName)
		g.writef("%s := func() { %s.%s(%s) }\n", updateFn, b.ElementName, setter, b.Expr)
		for _, stateName := range b.StateVars {
			stateType := stateTypes[stateName]
			g.writef("%s.Bind(func(_ %s) { %s() })\n", stateName, stateType, updateFn)
		}
	}
}

// getSetterForAttribute returns the element setter method for a given attribute.
func (g *Generator) getSetterForAttribute(attr string) string {
	switch attr {
	case "text":
		return "SetText"
	case "class":
		// Note: class attribute bindings would need SetClass or similar
		// For now, we don't support dynamic class bindings since element
		// doesn't have a SetClass method. This is a future enhancement.
		return ""
	default:
		return ""
	}
}
