package tuigen

// Node is the interface implemented by all AST nodes.
type Node interface {
	node()        // marker method to ensure type safety
	Pos() Position // returns the source position of the node
}

// File represents a complete .tui source file.
type File struct {
	Package    string
	Imports    []Import
	Components []*Component
	Funcs      []*GoFunc // top-level Go functions
	Position   Position
}

func (f *File) node()        {}
func (f *File) Pos() Position { return f.Position }

// Import represents a Go import statement.
type Import struct {
	Alias    string // optional alias (empty if none)
	Path     string // import path
	Position Position
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
	Tag        string
	Attributes []*Attribute
	Children   []Node // Elements, GoExpr, TextContent, ForLoop, IfStmt, LetBinding
	SelfClose  bool
	Position   Position
}

func (e *Element) node()        {}
func (e *Element) Pos() Position { return e.Position }

// Attribute represents a tag attribute: name=value or name={expr}
type Attribute struct {
	Name     string
	Value    Node // StringLit, IntLit, FloatLit, GoExpr, or BoolLit
	Position Position
}

func (a *Attribute) node()        {}
func (a *Attribute) Pos() Position { return a.Position }

// GoExpr represents a Go expression embedded in {braces}.
type GoExpr struct {
	Code     string
	Position Position
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
	Text     string
	Position Position
}

func (t *TextContent) node()        {}
func (t *TextContent) Pos() Position { return t.Position }

// LetBinding represents @let name = <element>.
type LetBinding struct {
	Name     string
	Element  *Element
	Position Position
}

func (l *LetBinding) node()        {}
func (l *LetBinding) Pos() Position { return l.Position }

// ForLoop represents @for i, v := range items { ... }
type ForLoop struct {
	Index    string // loop index variable (may be "_" or empty)
	Value    string // loop value variable
	Iterable string // Go expression for the iterable
	Body     []Node // Elements and other nodes
	Position Position
}

func (f *ForLoop) node()        {}
func (f *ForLoop) Pos() Position { return f.Position }

// IfStmt represents @if condition { ... } @else { ... }
type IfStmt struct {
	Condition string // Go expression for the condition
	Then      []Node
	Else      []Node // optional else branch
	Position  Position
}

func (i *IfStmt) node()        {}
func (i *IfStmt) Pos() Position { return i.Position }

// GoCode represents a block of embedded Go code.
type GoCode struct {
	Code     string
	Position Position
}

func (g *GoCode) node()        {}
func (g *GoCode) Pos() Position { return g.Position }

// GoFunc represents a top-level Go function definition in a .tui file.
type GoFunc struct {
	Code     string // the entire function definition
	Position Position
}

func (g *GoFunc) node()        {}
func (g *GoFunc) Pos() Position { return g.Position }

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
	Name     string // component name (e.g., "Card", "Header")
	Args     string // raw Go expression for arguments
	Children []Node // child elements (may be empty if no children block)
	Position Position
}

func (c *ComponentCall) node()        {}
func (c *ComponentCall) Pos() Position { return c.Position }

// ChildrenSlot represents {children...} placeholder in a component body
type ChildrenSlot struct {
	Position Position
}

func (c *ChildrenSlot) node()        {}
func (c *ChildrenSlot) Pos() Position { return c.Position }
