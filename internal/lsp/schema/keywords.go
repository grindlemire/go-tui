package schema

// KeywordDef describes a GSX language keyword.
type KeywordDef struct {
	Name          string
	Description   string
	Syntax        string
	Documentation string // Markdown documentation
}

// Canonical keyword definitions. Bare forms ("for", "if", etc.) point to
// the same *KeywordDef as the prefixed forms ("@for", "@if", etc.) to
// avoid duplicating multi-line documentation strings.
var (
	kwTempl = &KeywordDef{
		Name:        "templ",
		Description: "Define a reusable TUI component",
		Syntax:      "templ Name(params) { ... }",
		Documentation: `## templ

Defines a reusable TUI component. Supports two forms:

### Function Component
` + "```gsx" + `
templ Name(param1 Type1, param2 Type2) {
    <div>...</div>
}
` + "```" + `

Compiles to a Go function returning ` + "`*tui.Element`" + `.

### Method Component (Struct)
` + "```gsx" + `
templ (s *MyStruct) Render() {
    <div>...</div>
}
` + "```" + `

Used with Go struct types that implement ` + "`tui.Component`" + `.
The struct, constructor, and methods are defined as regular Go code
in the same .gsx file:

` + "```gsx" + `
type sidebar struct {
    expanded *tui.State[bool]
}

func Sidebar() *sidebar {
    return &sidebar{expanded: tui.NewState(true)}
}

func (s *sidebar) KeyMap() tui.KeyMap {
    return tui.KeyMap{...}
}

templ (s *sidebar) Render() {
    <div>...</div>
}
` + "```" + `

### Children Slot

Components can accept children using ` + "`{children...}`" + `:

` + "```gsx" + `
templ Card(title string) {
    <div class="border-rounded p-1">
        <span>{title}</span>
        {children...}
    </div>
}
` + "```" + `

Callers pass children with a block:

` + "```gsx" + `
@Card("Title") {
    <span>Content</span>
}
` + "```",
	}

	kwFor = &KeywordDef{
		Name:        "@for",
		Description: "Loop over items",
		Syntax:      "@for index, item := range collection { ... }",
		Documentation: `## @for

Iterates over a collection, rendering elements for each item.

**Syntax:**
` + "```gsx" + `
@for index, item := range collection {
    <element>...</element>
}
` + "```" + `

**Example:**
` + "```gsx" + `
@for i, name := range names {
    <li>{fmt.Sprintf("%d. %s", i+1, name)}</li>
}
` + "```" + `

- Supports standard Go range syntax
- Variables are scoped to the loop body
- Can iterate over slices, arrays, maps, and channels`,
	}

	kwIf = &KeywordDef{
		Name:        "@if",
		Description: "Conditional rendering",
		Syntax:      "@if condition { ... } @else { ... }",
		Documentation: `## @if

Conditionally renders elements based on a boolean expression.

**Syntax:**
` + "```gsx" + `
@if condition {
    <element>...</element>
} @else {
    <element>...</element>
}
` + "```" + `

**Example:**
` + "```gsx" + `
@if user.IsAdmin {
    <span class="text-green">Admin</span>
} @else {
    <span class="text-dim">User</span>
}
` + "```" + `

- Condition must be a Go boolean expression
- @else clause is optional`,
	}

	kwElse = &KeywordDef{
		Name:        "@else",
		Description: "Else branch of conditional",
		Syntax:      "} @else { ... }",
		Documentation: `## @else

The else branch of a conditional @if statement.

**Syntax:**
` + "```gsx" + `
@if condition {
    <element>...</element>
} @else {
    <element>...</element>
}
` + "```" + `

- Must follow an @if block
- Contains elements to render when condition is false`,
	}

	kwLet = &KeywordDef{
		Name:        "@let",
		Description: "Bind element to variable",
		Syntax:      "@let varName = <element>",
		Documentation: `## @let

Creates a local binding within a component body.

**Syntax:**
` + "```gsx" + `
@let varName = <element>
` + "```" + `

**Examples:**
` + "```gsx" + `
@let header = <div class="font-bold">{title}</div>
{header}
` + "```" + `

- Binds element trees to variables
- Variables are scoped to the component body
- Useful for reusable sub-elements`,
	}

	kwPackage = &KeywordDef{
		Name:        "package",
		Description: "Go package declaration",
		Syntax:      "package mypackage",
		Documentation: `## package

Declares the Go package for the generated code.

**Syntax:**
` + "```gsx" + `
package mypackage
` + "```" + `

- Must be the first declaration in the file
- Generated Go code will have this package name
- Should match the directory name (Go convention)`,
	}

	kwImport = &KeywordDef{
		Name:        "import",
		Description: "Import Go packages",
		Syntax:      "import ( ... )",
		Documentation: `## import

Imports Go packages for use in expressions.

**Syntax:**
` + "```gsx" + `
import (
    "fmt"
    "strings"

    "mymodule/pkg/utils"
)
` + "```" + `

- Standard Go import syntax
- Imported packages can be used in {} expressions
- The tui package is automatically available`,
	}

	kwFunc = &KeywordDef{
		Name:        "func",
		Description: "Define a helper function or method",
		Syntax:      "func name(params) returnType { ... }",
		Documentation: `## func

Defines a Go function or method at the file level. Unlike ` + "`templ`" + `,
these are plain Go code compiled as-is into the generated file.

### Helper Function
` + "```gsx" + `
func formatLabel(s string) string {
    return fmt.Sprintf("[%s]", s)
}
` + "```" + `

### Method (for struct components)
` + "```gsx" + `
func (s *sidebar) KeyMap() tui.KeyMap {
    return tui.KeyMap{
        tui.OnKey(tui.KeyCtrlB, s.toggle),
    }
}
` + "```" + `

- Regular Go function and method syntax
- Cannot contain GSX element literals
- Methods are used with struct components for ` + "`KeyMap()`" + `, ` + "`Init()`" + `, callbacks, etc.
- Compiled as-is into the generated Go file`,
	}
)

// Keywords maps keyword names to their definitions.
// Both the bare form ("for") and the prefixed form ("@for") are included
// so lookups work regardless of how the cursor word is extracted.
// Bare forms point to the same *KeywordDef as their prefixed counterparts.
var Keywords = map[string]*KeywordDef{
	"templ":   kwTempl,
	"@for":    kwFor,
	"for":     kwFor,
	"@if":     kwIf,
	"if":      kwIf,
	"@else":   kwElse,
	"else":    kwElse,
	"@let":    kwLet,
	"let":     kwLet,
	"package": kwPackage,
	"import":  kwImport,
	"func":    kwFunc,
}

// GetKeyword returns the keyword definition, or nil if unknown.
func GetKeyword(name string) *KeywordDef {
	return Keywords[name]
}
