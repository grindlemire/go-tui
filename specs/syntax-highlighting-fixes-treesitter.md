# Tree-sitter Grammar & Highlights Fixes (Zed / Neovim)

## Overview

The tree-sitter grammar (`editor/tree-sitter-gsx/grammar.js`) defines the parse tree for
`.gsx` files. The highlight queries (`editor/tree-sitter-gsx/queries/highlights.scm`) map
tree nodes to highlight groups. The same queries are duplicated for the Zed extension at
`editor/zed-gsx/languages/gsx/highlights.scm`.

Several highlight patterns have conflicts or are missing context, causing incorrect
highlighting in tree-sitter-based editors (Zed, Neovim with tree-sitter).

## Files to Edit

- `editor/tree-sitter-gsx/queries/highlights.scm` — Primary highlight queries
- `editor/zed-gsx/languages/gsx/highlights.scm` — Zed copy (keep in sync)

---

## Fix 1: `<`, `>`, `*`, `/` tag delimiter vs operator conflicts (Critical)

### Problem

The highlights.scm file defines bare (context-free) patterns for symbols that appear in
multiple grammatical contexts. Tree-sitter resolves conflicts between bare patterns by
giving priority to whichever pattern is defined **later** in the file. This causes ALL
instances of a symbol to get the same highlight, regardless of context.

Current patterns:

```scheme
; Lines 65-68 — Tag delimiters section (defined FIRST)
"<" @tag.delimiter
">" @tag.delimiter
"</" @tag.delimiter
"/" @tag.delimiter

; Lines 193-206 — Operators section (defined LATER, wins all conflicts)
"<" @operator
">" @operator
"*" @operator
"/" @operator
```

Because the operator patterns appear later, they win. This means:
- `<div>` — the `<` and `>` are highlighted as **operators** (wrong, should be tag delimiters)
- `<hr />` — the `/` is highlighted as an **operator** (wrong, should be tag delimiter)
- `i < 5` — the `<` is highlighted as an operator (correct, but only by accident)

### Root Cause

Bare anonymous node patterns like `"<" @tag.delimiter` match ALL anonymous `<` nodes in the
entire parse tree, regardless of what parent node they appear under. When two bare patterns
match the same node, the one defined later in the file takes precedence.

### Fix

Replace bare patterns with **context-specific** patterns that reference the parent node. This
way each `<` or `>` gets the correct highlight based on where it appears in the parse tree.

Remove the bare tag delimiter lines (lines 65-68) and the conflicting bare operator lines for
`<`, `>`, `*`, `/`. Replace them with context-specific patterns:

**Tag delimiters** — add in the Elements section:

```scheme
; Tag delimiters (context-specific to avoid operator conflicts)
(self_closing_element "<" @tag.delimiter)
(self_closing_element "/" @tag.delimiter)
(self_closing_element ">" @tag.delimiter)

(element_with_children "<" @tag.delimiter)
(element_with_children ">" @tag.delimiter)
(element_with_children "</" @tag.delimiter)
```

**Operators** — keep in the Operators section but make `<`, `>`, `*`, `/` context-specific:

```scheme
; Comparison and arithmetic operators (context-specific)
(binary_expression "<" @operator)
(binary_expression ">" @operator)
(binary_expression "<=" @operator)
(binary_expression ">=" @operator)
(binary_expression "*" @operator)
(binary_expression "/" @operator)
(binary_expression "+" @operator)
(binary_expression "-" @operator)
(binary_expression "==" @operator)
(binary_expression "!=" @operator)
(binary_expression "&&" @operator)
(binary_expression "||" @operator)

; These can remain bare (no conflicts):
":=" @operator
"=" @operator
```

**Pointer star** — already handled in the Types section:

```scheme
(pointer_type "*" @operator)
```

This is already context-specific and correct.

### Verification

Parse a `.gsx` file containing both elements and comparisons:

```gsx
templ Foo(count int) {
    @if count > 0 {
        <div class="flex">
            <span>{count}</span>
        </div>
    }
}
```

Confirm:
- `<div>` and `</div>` — `<`, `>`, `</` are tag delimiters (not operators)
- `<span>` and `</span>` — same
- `count > 0` — `>` is an operator (not a tag delimiter)
- `<hr />` — `<`, `/`, `>` are tag delimiters

Run `npx tree-sitter highlight` on a test file to verify:
```bash
cd editor/tree-sitter-gsx
npx tree-sitter highlight ../../editor/vscode/test/complex.gsx
```

---

## Fix 2: `selector_expression` in `call_expression` captures wrong identifier (Medium)

### Problem

The current pattern for method calls inside call expressions:

```scheme
(call_expression
  (selector_expression
    (identifier) @function.method.call))
```

This captures ALL `identifier` children of the `selector_expression`. A selector expression
`fmt.Sprintf` has two identifiers: `fmt` (package) and `Sprintf` (method). Both get captured
as `@function.method.call`, which is wrong — `fmt` should be a namespace/module, not a
function.

### Root Cause

The pattern `(selector_expression (identifier) @capture)` captures every `identifier` node
that is a child of the selector expression. It doesn't distinguish between the first child
(the object/package) and the second child (the method/property).

### Fix

Use a more specific pattern that only captures the **second** identifier (the method name).
A `selector_expression` rule in the grammar is `seq($._expression, ".", $.identifier)`, so
the method name is the `identifier` field that comes after the `.`. Use the tree-sitter field
syntax or positional matching:

Replace the existing pattern (lines 118-120):

```scheme
; OLD — captures both identifiers:
; (call_expression
;   (selector_expression
;     (identifier) @function.method.call))

; NEW — captures only the called method name (after the dot):
(call_expression
  (selector_expression
    (_)
    (identifier) @function.method.call))
```

The `(_)` wildcard matches the first child (the object/package expression) without capturing
it. The `(identifier) @function.method.call` then captures only the second child — the method
name.

Additionally, ensure the package/object part gets the right highlight. The general
`selector_expression` pattern (lines 168-170) already handles this:

```scheme
(selector_expression
  (identifier) @variable
  (identifier) @property)
```

This captures the first identifier as `@variable` and the second as `@property`. The more
specific `call_expression > selector_expression` pattern will override the `@property` capture
with `@function.method.call` for the method name when it's being called. The first identifier
keeps `@variable`.

### Verification

In `fmt.Sprintf("hello")`:
- `fmt` should be highlighted as a variable/namespace (not as a function)
- `Sprintf` should be highlighted as a function/method call

In `count.Get()`:
- `count` should be a variable
- `Get` should be a function/method call

In `tui.NewState(0)`:
- `tui` should be a variable/namespace
- `NewState` should be a function call

---

## Fix 3: Package name falls through to generic variable (Low)

### Problem

The `package main` declaration's package name (`main`) has no specific highlight pattern. It
falls through to the generic catch-all `(identifier) @variable` at line 173, making it appear
as a regular variable instead of a namespace/module.

### Root Cause

There's no `(package_clause name: ...)` pattern in highlights.scm.

### Fix

Add a pattern in the Keywords section (after `"package" @keyword`):

```scheme
; Package name
(package_clause
  name: (identifier) @module)
```

Place this BEFORE the generic `(identifier) @variable` pattern so it takes priority.

### Verification

In `package main`, confirm `main` is highlighted as a module/namespace (not a variable).

---

## Keeping Zed Extension in Sync

The Zed extension at `editor/zed-gsx/languages/gsx/highlights.scm` is a copy of the
tree-sitter queries. After making the above changes to
`editor/tree-sitter-gsx/queries/highlights.scm`, copy the updated file to the Zed location:

```bash
cp editor/tree-sitter-gsx/queries/highlights.scm editor/zed-gsx/languages/gsx/highlights.scm
```

Also check `editor/zed-gsx/languages/gsx/injections.scm` — it should match
`editor/tree-sitter-gsx/queries/injections.scm`. Currently they are identical and no changes
are needed for injections.
