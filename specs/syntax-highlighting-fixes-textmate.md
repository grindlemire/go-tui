# TextMate Grammar Fixes for VS Code

## Overview

The VS Code extension uses a TextMate grammar (`editor/vscode/syntaxes/gsx.tmLanguage.json`)
to provide immediate syntax highlighting for `.gsx` files before the LSP starts. Several
patterns are broken or incomplete, causing parts of the syntax to appear as plain unhighlighted
text.

## File to Edit

`editor/vscode/syntaxes/gsx.tmLanguage.json`

---

## Fix 1: Component parameters have no highlighting (Critical)

### Problem

The `component-declaration` rule's `begin` regex consumes the entire parameter list as a single
capture group but never assigns it a TextMate scope. The result is that every component's
parameters appear as plain text with no highlighting.

Example — in `templ Header(title string)`, the text `title string` has no color.

### Root Cause

The `begin` pattern on line 74:

```json
"begin": "^\\s*(templ)\\s+([a-zA-Z_][a-zA-Z0-9_]*)\\s*\\(([^)]*)\\)\\s*\\{"
```

Capture group 3 `([^)]*)` grabs the parameter text, but `beginCaptures` only assigns scopes to
groups 1 and 2:

```json
"beginCaptures": {
    "1": { "name": "keyword.function.gsx" },
    "2": { "name": "entity.name.function.component.gsx" }
}
```

Group 3 has no scope. And since the regex consumes everything through `{`, no child pattern
(like `#parameter-list`) can ever match the parameters — they're already consumed.

### Fix

Change the `component-declaration` to NOT consume the parameter list in the `begin` regex.
Instead, stop at the opening `(` and let child patterns handle parameters, similar to how
`function-declaration` works.

Replace the current `component-declaration`:

```json
"component-declaration": {
    "name": "meta.component.gsx",
    "begin": "^\\s*(templ)\\s+([a-zA-Z_][a-zA-Z0-9_]*)\\s*\\(",
    "end": "(?<=\\})",
    "beginCaptures": {
        "1": { "name": "keyword.function.gsx" },
        "2": { "name": "entity.name.function.component.gsx" }
    },
    "patterns": [
        { "include": "#parameter-list" },
        { "include": "#component-body" }
    ]
}
```

This makes it structurally identical to `function-declaration`: the `begin` stops at `(`, the
`#parameter-list` rule (which uses a lookbehind `(?<=\\()`) matches the parameters, and
`#component-body` matches the `{ ... }` block.

The existing `#parameter-list` rule (lines 84-101) already handles `name type` pairs with the
correct scopes (`variable.parameter.gsx` and `entity.name.type.gsx`).

### Verification

Open any `.gsx` file and confirm that in `templ Foo(name string, count int)`:
- `templ` is highlighted as a keyword
- `Foo` is highlighted as a function/component name
- `name` and `count` are highlighted as parameters
- `string` and `int` are highlighted as types

---

## Fix 2: Missing function call patterns in inline Go expressions (Medium)

### Problem

The `#go-expression-inline` pattern is used inside `@for` and `@if` conditions (e.g.,
`@if len(items) > 0`). It includes basic patterns like keywords, constants, operators, and
identifiers, but is **missing** function call patterns. This means function calls in conditions
are highlighted as plain variables.

Example — in `@if len(items) > 0`, `len` shows as a variable instead of a builtin function.
In `@for` iterables referencing `count.Get()`, `Get` shows as a variable instead of a function.

### Root Cause

The `#go-expression-inline` rule (lines 320-328) only includes:

```json
"go-expression-inline": {
    "patterns": [
        { "include": "#go-keywords" },
        { "include": "#go-constants" },
        { "include": "#go-operators" },
        { "include": "#go-identifiers" },
        { "include": "#strings" },
        { "include": "#numbers" }
    ]
}
```

It is missing `#go-builtin-functions`, `#go-function-call`, and `#state-method-call`.

### Fix

Add the missing includes. The order matters — more specific patterns must come before the
generic `#go-identifiers` catch-all:

```json
"go-expression-inline": {
    "patterns": [
        { "include": "#go-keywords" },
        { "include": "#go-constants" },
        { "include": "#go-builtin-functions" },
        { "include": "#state-declaration" },
        { "include": "#state-method-call" },
        { "include": "#go-operators" },
        { "include": "#go-function-call" },
        { "include": "#go-identifiers" },
        { "include": "#strings" },
        { "include": "#numbers" }
    ]
}
```

### Verification

Open a `.gsx` file with conditions and confirm:
- `@if len(items) > 0` — `len` is highlighted as a builtin function
- `@if count.Get() > 0` — `Get` is highlighted as a function
- `@for i, item := range items` — `range` is a keyword, `i` and `item` are identifiers

---

## Fix 3: Ref declarations not highlighted (Medium)

### Problem

The `#state-declaration-line` pattern only matches `tui.NewState(...)` declarations. It does
not match `tui.NewRef()`, `tui.NewRefList()`, or `tui.NewRefMap[...]()`. These lines appear
as completely unhighlighted plain text.

Example — `container := tui.NewRef()` has no highlighting at all.

### Root Cause

The regex on line 128 is hardcoded to `NewState`:

```json
"match": "([a-zA-Z_][a-zA-Z0-9_]*)\\s*(:=)\\s*(tui)(\\.)(NewState)\\s*\\(([^)]*)\\)"
```

### Fix

Broaden the function name match to include all `New*` constructor functions from the tui
package. Replace `(NewState)` with a group that matches all relevant constructors:

```json
"state-declaration-line": {
    "comment": "State/ref variable declaration at body level: name := tui.NewState(value) or tui.NewRef() etc.",
    "name": "meta.state-declaration.gsx",
    "match": "([a-zA-Z_][a-zA-Z0-9_]*)\\s*(:=)\\s*(tui)(\\.)(NewState|NewRef|NewRefList|NewRefMap)\\s*(?:\\[[^\\]]*\\])?\\s*\\(([^)]*)\\)",
    "captures": {
        "1": { "name": "variable.other.state.gsx" },
        "2": { "name": "keyword.operator.go.gsx" },
        "3": { "name": "variable.other.go.gsx" },
        "4": { "name": "punctuation.delimiter.go.gsx" },
        "5": { "name": "entity.name.function.go.gsx" },
        "6": { "name": "meta.embedded.block.go" }
    }
}
```

Key changes:
- `(NewState)` → `(NewState|NewRef|NewRefList|NewRefMap)` to match all constructors
- Added optional `(?:\\[[^\\]]*\\])?` after the function name to handle generic type parameters
  like `NewRefMap[string]`

### Verification

Open `editor/vscode/test/complex.gsx` (or `examples/refs-demo/refs.gsx`) and confirm:
- `container := tui.NewRef()` — `container` is a state variable, `tui.NewRef` is highlighted
- `itemRefs := tui.NewRefList()` — same treatment
- `userRefs := tui.NewRefMap[string]()` — same treatment, generic param not breaking the match
- `count := tui.NewState(0)` — still works as before (regression check)

---

## Fix 4: `@let` equals sign not highlighted (Low)

### Problem

The `@let` pattern captures the keyword and variable name but the `=` sign between the variable
and its value is not highlighted as an operator.

### Root Cause

The `@let` pattern on line 168:

```json
"begin": "(@let)\\s+([a-zA-Z_][a-zA-Z0-9_]*)\\s*=",
"end": "(?=<|\\{|$)",
"beginCaptures": {
    "1": { "name": "keyword.control.let.gsx" },
    "2": { "name": "variable.other.gsx" }
}
```

The `=` is part of the `begin` regex but has no capture group or scope.

### Fix

Add a capture for the `=`:

```json
"begin": "(@let)\\s+([a-zA-Z_][a-zA-Z0-9_]*)\\s*(=)",
"beginCaptures": {
    "1": { "name": "keyword.control.let.gsx" },
    "2": { "name": "variable.other.gsx" },
    "3": { "name": "keyword.operator.assignment.gsx" }
}
```

### Verification

In `@let label = fmt.Sprintf(...)`, confirm the `=` is highlighted as an operator.
