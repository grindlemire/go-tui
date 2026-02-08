package tuigen

import "strings"

// Node is the interface implemented by all AST nodes.
type Node interface {
	node()         // marker method to ensure type safety
	Pos() Position // returns the source position of the node
}

// Comment represents a single comment (line or block).
type Comment struct {
	Text            string   // Raw text including delimiters (// or /* */)
	Position        Position // Start position
	EndLine         int      // End line (for multi-line block comments)
	EndCol          int      // End column
	IsBlock         bool     // true for /* */ comments, false for // comments
	BlankLineBefore bool     // true if there was a blank line before this comment
}

// CommentGroup represents a sequence of comments with no blank lines between them.
// Adjacent line comments or a single block comment form a group.
type CommentGroup struct {
	List []*Comment
}

// Text returns the text of the comment group, with comment markers removed
// and lines joined with newlines.
func (g *CommentGroup) Text() string {
	if g == nil || len(g.List) == 0 {
		return ""
	}
	var lines []string
	for _, c := range g.List {
		text := c.Text
		if c.IsBlock {
			// Remove /* and */
			text = strings.TrimPrefix(text, "/*")
			text = strings.TrimSuffix(text, "*/")
			text = strings.TrimSpace(text)
		} else {
			// Remove //
			text = strings.TrimPrefix(text, "//")
			text = strings.TrimSpace(text)
		}
		lines = append(lines, text)
	}
	return strings.Join(lines, "\n")
}

// File represents a complete .tui source file.
type File struct {
	Package    string
	Imports    []Import
	Decls      []*GoDecl    // top-level Go declarations (type, const, var)
	Components []*Component
	Funcs      []*GoFunc // top-level Go functions
	Position   Position
	// Comment fields
	LeadingComments *CommentGroup   // Comments before package declaration
	OrphanComments  []*CommentGroup // Comments not attached to any node
}

func (f *File) node()        {}
func (f *File) Pos() Position { return f.Position }

// Import represents a Go import statement.
type Import struct {
	Alias    string // optional alias (empty if none)
	Path     string // import path
	Position Position
	// Comment fields
	TrailingComments *CommentGroup // Inline comment on import line
}

func (i *Import) node()        {}
func (i *Import) Pos() Position { return i.Position }

// Component represents a @component definition.
type Component struct {
	Name            string
	Params          []*Param
	ReturnType      string // defaults to "*element.Element"
	Body            []Node // Element, GoCode, LetBinding, ForLoop, IfStmt
	AcceptsChildren bool   // true if body contains {children...}
	Position        Position
	// Method receiver fields (for templ (recv) Render() syntax)
	Receiver     string // Full receiver text, e.g. "s *sidebar" (empty for function components)
	ReceiverName string // Receiver variable name, e.g. "s" (empty for function components)
	ReceiverType string // Receiver type, e.g. "*sidebar" (empty for function components)
	// Comment fields
	LeadingComments  *CommentGroup   // Doc comments before @component
	TrailingComments *CommentGroup   // Comments on same line after opening {
	OrphanComments   []*CommentGroup // Comments in body not attached to any node
}

func (c *Component) node()        {}
func (c *Component) Pos() Position { return c.Position }

// Param represents a function parameter.
type Param struct {
	Name     string
	Type     string
	Position Position
}

func (p *Param) node()        {}
func (p *Param) Pos() Position { return p.Position }

// Element represents an XML-like element: <tag attrs>children</tag> or <tag />
type Element struct {
	Tag     string
	RefExpr *GoExpr // Expression from ref={expr} attribute (e.g., ref={content})
	RefKey  *GoExpr // Key expression for map-based refs (e.g., key={item.ID})
	Attributes []*Attribute
	Children   []Node // Elements, GoExpr, TextContent, ForLoop, IfStmt, LetBinding
	SelfClose  bool
	Position   Position
	// Layout hints (detected from source positions during parsing)
	MultiLineAttrs        bool // attrs span multiple source lines
	ClosingBracketNewLine bool // > or /> is on its own line (after last attr)
	InlineChildren        bool // children are on same line as opening/closing tags
	BlankLineBefore        bool // blank line before this node in source
	// Comment fields
	LeadingComments  *CommentGroup   // Comments immediately before this element
	TrailingComments *CommentGroup   // Comments on same line after this element
	OrphanComments   []*CommentGroup // Comments in children not attached to any node
}

func (e *Element) node()        {}
func (e *Element) Pos() Position { return e.Position }

// Attribute represents a tag attribute: name=value or name={expr}
type Attribute struct {
	Name          string
	Value         Node     // StringLit, IntLit, FloatLit, GoExpr, or BoolLit
	Position      Position // Position of the attribute name
	ValuePosition Position // Position of the attribute value (start of value, after '=')
}

func (a *Attribute) node()        {}
func (a *Attribute) Pos() Position { return a.Position }

// GoExpr represents a Go expression embedded in {braces}.
type GoExpr struct {
	Code            string
	Position        Position
	BlankLineBefore bool // blank line before this node in source
	// Comment fields
	LeadingComments  *CommentGroup // Comments immediately before this expression
	TrailingComments *CommentGroup // Comments on same line after this expression
}

func (g *GoExpr) node()        {}
func (g *GoExpr) Pos() Position { return g.Position }

// StringLit represents a string literal "...".
type StringLit struct {
	Value    string
	Position Position
}

func (s *StringLit) node()        {}
func (s *StringLit) Pos() Position { return s.Position }

// IntLit represents an integer literal.
type IntLit struct {
	Value    int64
	Position Position
}

func (i *IntLit) node()        {}
func (i *IntLit) Pos() Position { return i.Position }

// FloatLit represents a floating-point literal.
type FloatLit struct {
	Value    float64
	Position Position
}

func (f *FloatLit) node()        {}
func (f *FloatLit) Pos() Position { return f.Position }

// BoolLit represents a boolean literal (true/false).
type BoolLit struct {
	Value    bool
	Position Position
}

func (b *BoolLit) node()        {}
func (b *BoolLit) Pos() Position { return b.Position }

// TextContent represents literal text content inside an element.
type TextContent struct {
	Text            string
	Position        Position
	BlankLineBefore bool // blank line before this node in source
}

func (t *TextContent) node()        {}
func (t *TextContent) Pos() Position { return t.Position }

// LetBinding represents @let name = <element>.
type LetBinding struct {
	Name            string
	Element         *Element
	Position        Position
	BlankLineBefore bool // blank line before this node in source
	// Comment fields
	LeadingComments  *CommentGroup // Comments immediately before @let
	TrailingComments *CommentGroup // Comments on same line after element
}

func (l *LetBinding) node()        {}
func (l *LetBinding) Pos() Position { return l.Position }

// ForLoop represents @for i, v := range items { ... }
type ForLoop struct {
	Index           string // loop index variable (may be "_" or empty)
	Value           string // loop value variable
	Iterable        string // Go expression for the iterable
	Body            []Node // Elements and other nodes
	Position        Position
	BlankLineBefore bool // blank line before this node in source
	// Comment fields
	LeadingComments  *CommentGroup   // Comments immediately before @for
	TrailingComments *CommentGroup   // Comments on same line after opening {
	OrphanComments   []*CommentGroup // Comments in body not attached to any node
}

func (f *ForLoop) node()        {}
func (f *ForLoop) Pos() Position { return f.Position }

// IfStmt represents @if condition { ... } @else { ... }
type IfStmt struct {
	Condition       string // Go expression for the condition
	Then            []Node
	Else            []Node // optional else branch
	Position        Position
	BlankLineBefore bool // blank line before this node in source
	// Comment fields
	LeadingComments  *CommentGroup   // Comments immediately before @if
	TrailingComments *CommentGroup   // Comments on same line after opening {
	OrphanComments   []*CommentGroup // Comments in body not attached to any node
}

func (i *IfStmt) node()        {}
func (i *IfStmt) Pos() Position { return i.Position }

// GoCode represents a block of embedded Go code.
type GoCode struct {
	Code     string
	Position Position
	// Comment fields
	LeadingComments  *CommentGroup // Comments immediately before Go code
	TrailingComments *CommentGroup // Comments on same line after Go code
}

func (g *GoCode) node()        {}
func (g *GoCode) Pos() Position { return g.Position }

// GoFunc represents a top-level Go function definition in a .tui file.
type GoFunc struct {
	Code     string // the entire function definition
	Position Position
	// Comment fields
	LeadingComments  *CommentGroup // Comments immediately before func
	TrailingComments *CommentGroup // Comments on same line after closing }
}

func (g *GoFunc) node()        {}
func (g *GoFunc) Pos() Position { return g.Position }

// GoDecl represents a top-level Go declaration (type, const, var) in a .gsx file.
type GoDecl struct {
	Kind     string // "type", "const", or "var"
	Code     string // the entire declaration
	Position Position
	// Comment fields
	LeadingComments  *CommentGroup // Comments immediately before declaration
	TrailingComments *CommentGroup // Comments on same line after declaration
}

func (g *GoDecl) node()        {}
func (g *GoDecl) Pos() Position { return g.Position }

// RawGoExpr represents a raw Go expression that should be emitted as-is
// (used for element references captured via @let)
type RawGoExpr struct {
	Code     string
	Position Position
}

func (r *RawGoExpr) node()        {}
func (r *RawGoExpr) Pos() Position { return r.Position }

// ComponentCall represents @ComponentName(args) { children }
type ComponentCall struct {
	Name            string // component name (e.g., "Card", "Header")
	Args            string // raw Go expression for arguments
	Children        []Node // child elements (may be empty if no children block)
	IsStructMount   bool   // true when inside a method templ (generates tui.Mount())
	Position        Position
	BlankLineBefore bool // blank line before this node in source
	// Comment fields
	LeadingComments  *CommentGroup // Comments immediately before @ComponentName
	TrailingComments *CommentGroup // Comments on same line after )
}

func (c *ComponentCall) node()        {}
func (c *ComponentCall) Pos() Position { return c.Position }

// ComponentExpr represents @expr where expr is a Component field/variable.
// The expression's .Render() method is called to get the element.
type ComponentExpr struct {
	Expr            string // expression (e.g., "c.textarea", "myComponent")
	Position        Position
	BlankLineBefore bool // blank line before this node in source
	// Comment fields
	LeadingComments  *CommentGroup // Comments immediately before @expr
	TrailingComments *CommentGroup // Comments on same line after expr
}

func (c *ComponentExpr) node()        {}
func (c *ComponentExpr) Pos() Position { return c.Position }

// ChildrenSlot represents {children...} placeholder in a component body
type ChildrenSlot struct {
	Position        Position
	BlankLineBefore bool // blank line before this node in source
	// Comment fields
	LeadingComments  *CommentGroup // Comments immediately before {children...}
	TrailingComments *CommentGroup // Comments on same line after {children...}
}

func (c *ChildrenSlot) node()        {}
func (c *ChildrenSlot) Pos() Position { return c.Position }
