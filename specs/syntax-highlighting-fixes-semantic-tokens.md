# Semantic Token Provider Fixes (LSP)

## Overview

The LSP's semantic token provider generates tokens that VS Code uses for rich, AST-aware
syntax highlighting. These tokens override or supplement the TextMate grammar. Several node
types are missing token generation, causing them to either fall back to TextMate highlighting
or have no highlighting at all.

## Files to Edit

- `internal/lsp/provider/semantic_nodes.go` — AST node token collection
- `internal/lsp/provider/semantic_gocode.go` — Go code tokenization

---

## Fix 1: No semantic tokens for element tag names (Medium)

### Problem

When the LSP provides semantic tokens, VS Code uses them and may suppress TextMate highlighting
in token-covered regions. The semantic token provider walks the AST and generates tokens for
keywords, attributes, variables, etc. — but it **never generates tokens for element tag names**.

This means `<div>`, `<span>`, `<button>`, `<input>`, etc. have no LSP-provided highlighting.
They rely entirely on the TextMate grammar's `entity.name.tag.gsx` scope. In most cases the
TextMate fallback works, but if there are any token position overlaps or if VS Code's semantic
token integration suppresses nearby TextMate scopes, tag names can lose their color.

### Root Cause

In `internal/lsp/provider/semantic_nodes.go`, the `collectTokensFromNode` method's
`*tuigen.Element` case (starting at line 30) handles:

1. `ref={}` attribute tokens
2. Regular attribute name/value tokens
3. Recursive child processing

But it never emits a token for the element's tag name. The tag name, opening `<`/`>`, and
closing `</tag>` are all skipped.

### Fix

Add tag name token emission at the start of the `*tuigen.Element` case in
`collectTokensFromNode`. The element's `Position` points to the `<` character. The tag name
starts at `Position.Column` (after the `<`). The element's tag name is stored in `n.Tag`.

Add this at the beginning of the `case *tuigen.Element:` block (after the nil check), before
the ref handling:

```go
case *tuigen.Element:
    if n == nil {
        return
    }

    // Emit token for the opening tag name
    // Element position points to '<', tag name starts one character after
    tagLine := n.Position.Line - 1
    tagCol := n.Position.Column // Column is 1-indexed, and tag starts after '<'
    if n.Tag != "" {
        *tokens = append(*tokens, SemanticToken{
            Line:      tagLine,
            StartChar: tagCol,
            Length:    len(n.Tag),
            TokenType: TokenTypeProperty, // Use "property" type (index 6) for tag names
            Modifiers: 0,
        })
    }

    // ... rest of the existing code (ref handling, attributes, children)
```

Note: Using `TokenTypeProperty` (index 6, "property") for tag names. This maps to tag-like
coloring in most VS Code themes. An alternative would be to add a dedicated tag token type,
but that would require updating the legend in `internal/lsp/handler.go` (the `TokenTypes`
array in `handleInitialize`), which could break existing clients. Using "property" is a safe
choice that provides visible differentiation.

However, the element tag must be found in the source. The `n.Position` in the AST points to
the `<` character (column is 1-indexed). The tag name starts at column `n.Position.Column`
(0-indexed: `n.Position.Column - 1 + 1` = `n.Position.Column`). Verify this by checking how
the parser sets element positions — look at `internal/tuigen/parser_element.go` to confirm
that `Position` points to `<`.

Also emit a token for the **closing tag name** if the element has children (meaning it's not
self-closing). The closing tag position isn't stored in the AST directly, so you'll need to
scan the document content. Use `s.currentContent` (available on the provider struct) to find
`</tagname>` after the element's children. A simpler approach: search from the last child's
position forward for `</` + tag name.

If finding the closing tag position is too complex, it's acceptable to only emit the opening
tag token and leave the closing tag to TextMate.

### Verification

Open any `.gsx` file and confirm:
- `<div>`, `<span>`, `<button>` tag names are colored (not plain text)
- Self-closing tags like `<hr />` and `<br />` also have colored tag names
- The color is consistent between opening and (if implemented) closing tags

---

## Fix 2: No semantic tokens for `ChildrenSlot` (Low)

### Problem

The `{children...}` slot placeholder used in components that accept children has no semantic
tokens. The `children` keyword and `...` operator are unhighlighted by the LSP.

### Root Cause

The `collectTokensFromNode` switch in `internal/lsp/provider/semantic_nodes.go` (lines 29-262)
has cases for `Element`, `GoExpr`, `RawGoExpr`, `ForLoop`, `IfStmt`, `LetBinding`, `GoCode`,
and `ComponentCall`, but no case for `*tuigen.ChildrenSlot`.

### Fix

Add a case for `*tuigen.ChildrenSlot` in the switch statement. The position points to the `{`
character. Emit tokens for:

1. `children` as a keyword (7 characters, starts at column + 1 for the `{`)
2. `...` as an operator (3 characters, starts after `children`)

```go
case *tuigen.ChildrenSlot:
    if n == nil {
        return
    }
    // {children...} — emit "children" as keyword
    *tokens = append(*tokens, SemanticToken{
        Line:      n.Position.Line - 1,
        StartChar: n.Position.Column, // after the '{'
        Length:    len("children"),
        TokenType: TokenTypeKeyword,
        Modifiers: 0,
    })
    // Emit "..." as operator
    *tokens = append(*tokens, SemanticToken{
        Line:      n.Position.Line - 1,
        StartChar: n.Position.Column + len("children"),
        Length:    3,
        TokenType: TokenTypeOperator,
        Modifiers: 0,
    })
```

Verify the position math by checking how the parser stores `ChildrenSlot` positions. Look at
`internal/tuigen/parser_element.go` or `parser_control.go` for `ChildrenSlot` construction to
confirm whether `Position` points to `{` or to `children`.

### Verification

In a component that uses `{children...}`, confirm:
- `children` is highlighted as a keyword
- `...` is highlighted as an operator
- The surrounding `{` and `}` may or may not be highlighted (they're punctuation)

---

## Fix 3: Multi-line Go code token positions incorrect (Low, edge case)

### Problem

The `collectTokensInGoCode` function in `internal/lsp/provider/semantic_gocode.go` (line 21)
processes a Go code string and emits semantic tokens. It uses `pos.Line - 1` as the line number
for ALL emitted tokens. If the code string contains newline characters (multi-line Go code), all
tokens after the first line will have incorrect line positions — they'll all be placed on the
first line.

This is mostly a theoretical issue because Go expressions in `.gsx` files are almost always
single-line (`{fmt.Sprintf(...)}`). But it could affect:
- Multi-line function bodies (though `collectTokensFromFuncBody` handles these line-by-line)
- Unusual multi-line Go expressions

### Root Cause

In `collectTokensInGoCode` (line 21-284), the main loop skips newlines:

```go
if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
    i++
    continue
}
```

But it never increments a line counter. Every token is emitted with `Line: pos.Line - 1`.

### Fix

Track the current line within the code string. When a `\n` is encountered, increment the line
counter and reset the column tracking:

```go
func (s *semanticTokensProvider) collectTokensInGoCode(code string, pos tuigen.Position, startOffset int, paramNames map[string]bool, localVars map[string]bool, tokens *[]SemanticToken) {
    currentLine := pos.Line - 1     // 0-indexed document line
    lineStartIdx := -startOffset    // index in code where the current line starts
    i := 0
    for i < len(code) {
        ch := code[i]

        if ch == '\n' {
            i++
            currentLine++
            lineStartIdx = i
            continue
        }
        if ch == ' ' || ch == '\t' || ch == '\r' {
            i++
            continue
        }

        // ... rest of tokenization, but replace:
        //   Line: pos.Line - 1
        //   StartChar: pos.Column - 1 + startOffset + start
        // with:
        //   Line: currentLine
        //   StartChar: if currentLine == pos.Line-1 { pos.Column - 1 + startOffset + start } else { start - lineStartIdx }
```

This is a larger refactor. A simpler alternative: when `code` contains `\n`, split it into
lines and call `collectTokensInGoCode` per-line (similar to what `collectTokensFromFuncBody`
already does). This avoids changing the core function's contract.

### Verification

Create a `.gsx` test file with a multi-line Go expression (if the parser supports it) and
confirm tokens appear on the correct lines. If multi-line Go expressions aren't supported by
the parser, this fix is truly low priority and can be deferred.
