# Comments Support Specification

**Status:** Planned
**Version:** 2.0
**Last Updated:** 2025-01-25

---

## 1. Overview

### Purpose

Add support for Go-style comments (`//` single-line and `/* */` multi-line) in `.tui` files. Comments will be preserved during formatting and available for LSP features, but discarded during code generation.

### Goals

- Support `//` single-line comments anywhere in `.tui` files
- Support `/* */` multi-line comments anywhere in `.tui` files
- Preserve comments when running `tui fmt` (round-trip preservation)
- Attach comments to AST nodes using a well-defined association algorithm
- Handle floating/orphan comments that don't attach to any node
- Report unterminated block comments as errors
- Provide LSP semantic token highlighting for comments
- Maintain backwards compatibility with existing code

### Non-Goals

- Preserving comments in generated `*_tui.go` files (comments are discarded during generation)
- HTML-style comments (`<!-- -->`)
- Doc comments with special formatting (e.g., godoc-style)
- Comment directives (e.g., `//go:generate`, `//nolint`)
- Nested block comments (`/* outer /* inner */ */` is an error)

---

## 2. Architecture

### Current Flow (Comments Lost)

```
.tui source → Lexer (skips comments) → Parser → AST (no comments) → Formatter → Output (comments lost)
```

### New Flow (Comments Preserved)

```
.tui source → Lexer (collects comments) → Parser (attaches to nodes) → AST (with comments) → Formatter → Output (comments preserved)
                                                                              ↓
                                                                           LSP (semantic tokens)
```

### Directory Structure

No new files created. Modifications to existing files:

```
pkg/tuigen/
├── token.go       # Add TokenLineComment, TokenBlockComment
├── lexer.go       # Collect comments, emit via CommentsBefore()
├── ast.go         # Add Comment, CommentGroup types
├── parser.go      # Attach comments to nodes during parsing

pkg/formatter/
└── printer.go     # Print comments when formatting nodes

pkg/lsp/
└── semantic_tokens.go  # Add tokenTypeComment for syntax highlighting
```

### Component Overview

| Component | Changes |
|-----------|---------|
| `pkg/tuigen/token.go` | Add `TokenLineComment` and `TokenBlockComment` token types |
| `pkg/tuigen/lexer.go` | Collect comments instead of skipping; add `ConsumeComments()` method |
| `pkg/tuigen/ast.go` | Add `Comment`, `CommentGroup` types; add comment fields to node structs |
| `pkg/tuigen/parser.go` | Collect pending comments, attach to nodes, store orphans in containers |
| `pkg/formatter/printer.go` | Print leading/trailing/orphan comments when formatting |
| `pkg/lsp/semantic_tokens.go` | Add comment token type (index 12) for syntax highlighting |

---

## 3. Core Entities

### Comment Types

```go
// Comment represents a single comment (line or block).
type Comment struct {
    Text     string   // Raw text including delimiters (// or /* */)
    Position Position // Start position
    EndLine  int      // End line (for multi-line block comments)
    EndCol   int      // End column
    IsBlock  bool     // true for /* */ comments, false for // comments
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
```

### Node Comment Fields (Composition Approach)

Rather than modifying the `Node` interface (breaking change), we add optional comment fields to each struct that can have comments. This maintains full backwards compatibility.

```go
// Element now includes comment fields
type Element struct {
    Tag             string
    Attributes      []*Attribute
    Children        []Node
    SelfClose       bool
    Position        Position
    // New comment fields
    LeadingComments  *CommentGroup // Comments immediately before this element
    TrailingComments *CommentGroup // Comments on same line after this element
}

// Component now includes comment fields
type Component struct {
    Name            string
    Params          []*Param
    ReturnType      string
    Body            []Node
    AcceptsChildren bool
    Position        Position
    // New comment fields
    LeadingComments  *CommentGroup // Doc comments before @component
    OrphanComments   []*CommentGroup // Comments in body not attached to any node
}

// File now includes comment fields
type File struct {
    Package    string
    Imports    []Import
    Components []*Component
    Funcs      []*GoFunc
    Position   Position
    // New comment fields
    LeadingComments *CommentGroup   // Comments before package declaration
    OrphanComments  []*CommentGroup // Comments not attached to any node
}

// Similar pattern for: IfStmt, ForLoop, LetBinding, ComponentCall, GoFunc, Import, Attribute
```

### Token Types

```go
const (
    // ... existing tokens ...

    TokenLineComment  TokenType = iota // // comment
    TokenBlockComment                  // /* comment */
)
```

---

## 4. Comment Association Algorithm

### Overview

Comments are associated with nodes using proximity and position rules:

1. **Trailing comments**: On the same line as a node, after the node
2. **Leading comments**: On previous lines, before the node, with no blank line separation
3. **Orphan comments**: Not associated with any node (stored in container)

### Detailed Algorithm

```
ALGORITHM: AttachComments(pending_comments, current_node, next_node)

INPUT:
  - pending_comments: List of unattached comments collected since last node
  - current_node: The node just parsed (or nil at start)
  - next_node: The node about to be parsed (or nil at end)

FOR each comment C in pending_comments:

  1. TRAILING CHECK (same line as previous node):
     IF current_node != nil AND C.Line == current_node.EndLine:
       Attach C to current_node.TrailingComments
       CONTINUE

  2. LEADING CHECK (immediately before next node):
     IF next_node != nil:
       blank_lines = C.EndLine - (previous_comment.EndLine or current_node.EndLine)
       IF blank_lines <= 1 AND C.EndLine == next_node.Line - 1:
         Attach C to next_node.LeadingComments
         CONTINUE

  3. ORPHAN (not attached to any node):
     Add C to container.OrphanComments
     (container = File, Component, or Element depending on scope)

GROUPING:
  - Adjacent comments (no blank line between) form a CommentGroup
  - A blank line (2+ newlines) starts a new group
```

### Examples

```tui
// Comment A (orphan - before package, stored in File.LeadingComments)

package example

// Comment B (leading for Import group)
import (
    "fmt"  // Comment C (trailing on import)
)

// Comment D - doc comment
// Comment D continued (same group, leading for Header)
@component Header(title string) {  // Comment E (trailing on @component line)
    // Comment F (orphan in component body - stored in Component.OrphanComments)

    // Comment G (leading for div)
    <div>  // Comment H (trailing on div open)
        // Comment I (orphan in element children)
        <span>{title}</span>  // Comment J (trailing on span)
    </div>  // Comment K (trailing on div close)
}
```

### Edge Cases

| Case | Behavior |
|------|----------|
| `<span /* comment */ class="x">` | Error: comments not allowed inside element tags |
| `@if /* comment */ condition` | Comment becomes trailing on `@if` keyword |
| `// comment at end of file` | Orphan, stored in File.OrphanComments |
| Comment-only file | Valid, stored as File.OrphanComments |
| `{children...} // comment` | Trailing on ChildrenSlot |
| `} // comment between } @else {` | Trailing on the `}` of the if-then body |
| `@if condition { // only comment }` | Orphan, stored in IfStmt.OrphanComments |
| `@let x = // comment here` followed by newline and `<div />` | Comment is trailing on `@let x =` partial, error during parse |

---

## 5. Go Code Blocks

Comments inside raw Go code blocks are **preserved as-is** since they're captured as raw strings:

```go
// These methods capture raw Go and preserve embedded comments:
// - ReadBalancedBraces()
// - ReadBalancedBracesFrom()
// - ReadUntilBrace()
// - parseGoFunc() - captures entire function body including comments
// - parseGoStatement() - captures statement including inline comments

// Examples that work:
{foo /* comment */ + bar}  // Preserved in GoExpr.Code
func helper() { // comment inside }  // Preserved in GoFunc.Code
x := 1 // inline comment  // Preserved in GoCode.Code
```

Comments in Go code blocks are the responsibility of the Go compiler, not the .tui parser.

---

## 6. Lexer Changes

### Current Behavior

```go
// skipWhitespaceAndComments skips spaces, tabs, and comments (but not newlines).
func (l *Lexer) skipWhitespaceAndComments() {
    for {
        switch l.ch {
        case ' ', '\t', '\r':
            l.readChar()
        case '/':
            if l.peekChar() == '/' {
                l.skipLineComment()  // Comments discarded
            } else if l.peekChar() == '*' {
                l.skipBlockComment() // Comments discarded
            } else {
                return
            }
        default:
            return
        }
    }
}
```

### New Behavior

```go
// Lexer now has a pending comments buffer
type Lexer struct {
    // ... existing fields ...
    pendingComments []*Comment // Comments collected since last ConsumeComments()
}

// skipWhitespace now only skips spaces/tabs, NOT comments
func (l *Lexer) skipWhitespace() {
    for l.ch == ' ' || l.ch == '\t' || l.ch == '\r' {
        l.readChar()
    }
}

// collectComment reads a comment and adds it to pending buffer
func (l *Lexer) collectComment() {
    l.startToken()

    if l.peekChar() == '/' {
        // Line comment
        startPos := l.pos
        for l.ch != '\n' && l.ch != 0 {
            l.readChar()
        }
        comment := &Comment{
            Text:     l.source[startPos:l.pos],
            Position: l.position(),
            EndLine:  l.line,
            EndCol:   l.column,
            IsBlock:  false,
        }
        l.pendingComments = append(l.pendingComments, comment)
    } else if l.peekChar() == '*' {
        // Block comment
        startPos := l.pos
        startLine := l.line
        l.readChar() // skip /
        l.readChar() // skip *

        for {
            if l.ch == 0 {
                l.errors.AddError(l.position(), "unterminated block comment")
                return
            }
            if l.ch == '*' && l.peekChar() == '/' {
                l.readChar() // skip *
                l.readChar() // skip /
                break
            }
            l.readChar()
        }

        comment := &Comment{
            Text:     l.source[startPos:l.pos],
            Position: Position{File: l.filename, Line: startLine, Column: l.tokenColumn},
            EndLine:  l.line,
            EndCol:   l.column,
            IsBlock:  true,
        }
        l.pendingComments = append(l.pendingComments, comment)
    }
}

// ConsumeComments returns and clears pending comments.
// Called by parser after each node is parsed.
func (l *Lexer) ConsumeComments() []*Comment {
    comments := l.pendingComments
    l.pendingComments = nil
    return comments
}

// Next now collects comments instead of skipping them
func (l *Lexer) Next() Token {
    for {
        l.skipWhitespace()

        // Check for comments
        if l.ch == '/' {
            if l.peekChar() == '/' || l.peekChar() == '*' {
                l.collectComment()
                continue // Keep looking for the actual token
            }
        }

        break
    }

    // ... rest of existing Next() logic ...
}
```

### Parser Integration

```go
// Parser calls ConsumeComments() at key points
func (p *Parser) parseElement() *Element {
    // Get any pending comments before this element
    leadingComments := p.collectLeadingComments()

    elem := &Element{
        LeadingComments: leadingComments,
        // ... parse element ...
    }

    // Check for trailing comment on same line
    if comments := p.lexer.ConsumeComments(); len(comments) > 0 {
        if comments[0].Position.Line == elem.Position.Line {
            elem.TrailingComments = &CommentGroup{List: comments[:1]}
            // Remaining comments go to next node or orphans
            p.pendingComments = append(p.pendingComments, comments[1:]...)
        } else {
            p.pendingComments = append(p.pendingComments, comments...)
        }
    }

    return elem
}
```

---

## 7. User Experience

### Comment Syntax

```tui
// Package-level comment
package mypackage

// Import comment
import (
    "fmt"  // inline comment
)

// Header component renders a page header.
// Multi-line doc comments are grouped.
@component Header(title string) {
    /* Block comment explaining
       the layout structure */
    <div class="header">
        // Comment inside body
        <span>{title}</span>  // trailing comment
    </div>
}

// Helper function
func formatTitle(s string) string {
    return fmt.Sprintf("[%s]", s)
}
```

### Formatting Behavior

Comments are preserved in their relative position:

**Input:**
```tui
// Leading comment
<div class="box">  // Trailing comment
    // Inner comment
    <span>Hello</span>
</div>
```

**After `tui fmt`:**
```tui
// Leading comment
<div class="box">  // Trailing comment
    // Inner comment
    <span>Hello</span>
</div>
```

### Error Messages

```
example.tui:5:1: unterminated block comment
example.tui:10:15: comments not allowed inside element tags
```

---

## 8. LSP Integration

### Semantic Token Type

Add comment token type to the semantic tokens legend:

```go
// semantic_tokens.go

const (
    // ... existing token types 0-11 ...
    tokenTypeComment = 12 // comments
)
```

### Token Collection

The LSP server will emit semantic tokens for comments by:

1. Using the AST's comment fields (LeadingComments, TrailingComments, OrphanComments)
2. Iterating through all comments and emitting tokens with `tokenTypeComment`

```go
// In collectSemanticTokens():
func (s *Server) collectCommentTokens(file *tuigen.File, tokens *[]semanticToken) {
    // Collect from File
    s.collectCommentGroupTokens(file.LeadingComments, tokens)
    for _, cg := range file.OrphanComments {
        s.collectCommentGroupTokens(cg, tokens)
    }

    // Collect from Components
    for _, comp := range file.Components {
        s.collectCommentGroupTokens(comp.LeadingComments, tokens)
        for _, cg := range comp.OrphanComments {
            s.collectCommentGroupTokens(cg, tokens)
        }
        // ... recursively collect from body nodes ...
    }
}

func (s *Server) collectCommentGroupTokens(cg *CommentGroup, tokens *[]semanticToken) {
    if cg == nil {
        return
    }
    for _, c := range cg.List {
        *tokens = append(*tokens, semanticToken{
            line:      c.Position.Line - 1,
            startChar: c.Position.Column - 1,
            length:    len(c.Text),
            tokenType: tokenTypeComment,
            modifiers: 0,
        })
    }
}
```

### Capabilities Registration

Update server initialization to include comment token type:

```go
SemanticTokensLegend: SemanticTokensLegend{
    TokenTypes: []string{
        "namespace", "type", "class", "function", "parameter",
        "variable", "property", "keyword", "string", "number",
        "operator", "decorator", "comment", // Added
    },
    // ...
}
```

---

## 9. Blank Line Preservation

### Rules

1. **Within a CommentGroup**: No blank lines between comments
2. **Between CommentGroups**: Blank lines are preserved via separate groups
3. **Between comment and node**: One blank line max keeps association; 2+ blank lines makes it orphan

### Example

```tui
// Comment group 1
// Still group 1 (no blank line)

// Comment group 2 (blank line before = new group)

<div>  // Both groups become leading comments for <div>
```

After formatting:
```tui
// Comment group 1
// Still group 1 (no blank line)

// Comment group 2 (blank line before = new group)
<div>
```

---

## 10. Migration & Backwards Compatibility

### API Compatibility

- **Node interface**: Unchanged (no new methods required)
- **Existing code**: Continues to work; comment fields default to nil
- **Parser output**: Same AST structure, just with additional optional fields populated

### Test Migration

Existing tests continue to pass because:
1. `Next()` still returns the same tokens (comments are collected, not emitted as tokens)
2. AST structure is unchanged
3. Formatter produces valid output (just now with comments)

New tests needed:
1. Lexer comment collection tests
2. Parser comment attachment tests
3. Formatter round-trip tests with comments
4. LSP semantic token tests for comments

### Round-Trip Property

The formatter must satisfy:
```
format(format(source)) == format(source)
```

For any valid `.tui` source with comments.

---

## 11. Complexity Assessment

| Factor | Assessment |
|--------|------------|
| Files modified | 6 (token.go, lexer.go, ast.go, parser.go, printer.go, semantic_tokens.go) |
| New types | 2 (Comment, CommentGroup) |
| Core logic changes | Lexer collection, parser attachment, formatter printing, LSP tokens |
| Breaking changes | None (additive only) |
| Testing scope | Unit tests for each component + integration round-trip tests |

**Assessed Size:** Medium
**Recommended Phases:** 4
**Rationale:** The feature touches multiple components with complex interaction patterns (comment association algorithm). Adding LSP support increases scope. Four phases allow for:
1. Token/lexer changes (foundation)
2. AST types + parser attachment
3. Formatter integration
4. LSP semantic tokens

> **IMPORTANT:** User must approve the complexity assessment before proceeding to implementation plan.

---

## 12. Success Criteria

1. `//` comments are collected by the lexer and stored in pending buffer
2. `/* */` comments are collected with correct start/end positions
3. Unterminated block comments produce clear error messages
4. Comments are attached to AST nodes following the association algorithm
5. Orphan comments are stored in container nodes (File, Component, etc.)
6. `tui fmt` preserves all comments in their relative positions
7. Round-trip formatting produces identical output
8. Comments work in all positions: top-level, components, elements, attributes
9. LSP provides syntax highlighting for comments
10. Existing tests continue to pass (backwards compatible)

---

## 13. Open Questions (Resolved)

| Question | Resolution |
|----------|------------|
| Comment syntax? | Go-style (`//` and `/* */`) |
| Preserve in generated code? | No, discarded during generation |
| Where allowed? | Everywhere except inside element tags |
| How to associate? | Attach to nearest node (leading/trailing) |
| Floating comments? | Store in container's OrphanComments |
| Breaking interface change? | No, use composition with optional fields |
| LSP support? | Yes, add tokenTypeComment (index 12) |
| Nested block comments? | Error (not supported, like Go) |
