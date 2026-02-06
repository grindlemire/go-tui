# VS Code Extension Configuration Fix

## Overview

The VS Code extension (`editor/vscode/`) registers the `.gsx` language, starts the LSP, and
configures how VS Code handles the language. One configuration improvement ensures semantic
highlighting is reliably enabled.

## File to Edit

- `editor/vscode/package.json`

---

## Fix 1: Explicitly enable semantic highlighting for gsx (Low)

### Problem

VS Code uses semantic tokens from the LSP to provide rich, AST-aware highlighting that
supplements the TextMate grammar. For built-in languages, VS Code enables this automatically.
For custom languages contributed by extensions, the behavior can be inconsistent — some VS Code
versions or configurations may not enable semantic highlighting by default, causing users to
see only the (more limited) TextMate grammar highlighting.

The extension's `package.json` does not explicitly enable semantic highlighting for the `gsx`
language.

### Root Cause

The `configurationDefaults` section (lines 74-80) currently only sets a custom color for
`typeParameter`:

```json
"configurationDefaults": {
    "editor.semanticTokenColorCustomizations": {
        "rules": {
            "typeParameter:gsx": "#C586C0"
        }
    }
}
```

It doesn't set `editor.semanticHighlighting.enabled` for the `gsx` language.

### Fix

Add an explicit language-scoped configuration to enable semantic highlighting. Update the
`configurationDefaults` section:

```json
"configurationDefaults": {
    "[gsx]": {
        "editor.semanticHighlighting.enabled": true
    },
    "editor.semanticTokenColorCustomizations": {
        "rules": {
            "typeParameter:gsx": "#C586C0"
        }
    }
}
```

The `[gsx]` scoped setting ensures that semantic highlighting is always active for `.gsx` files,
regardless of the user's global setting.

### Verification

1. Install the extension (or run in Extension Development Host)
2. Open a `.gsx` file
3. Open VS Code command palette → "Developer: Inspect Editor Tokens and Scopes"
4. Click on different tokens and confirm both "textmate scopes" and "semantic token type" are
   shown — the presence of "semantic token type" confirms semantic highlighting is active
5. Confirm tokens like `templ` show a semantic token type of `keyword`, component names show
   `class`, parameters show `parameter`, etc.
