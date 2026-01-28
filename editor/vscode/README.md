# GSX Language Support for VS Code

Syntax highlighting and language support for `.gsx` files used with the [go-tui](https://github.com/grindlemire/go-tui) framework.

## Features

- **Syntax Highlighting**: Full highlighting support for the GSX DSL
  - Component declarations: `templ Name(params) {... }`
  - Keywords: `@for`, `@if`, `@else`, `@let`
  - Element tags: `<box>`, `<text>`, `<div>`, `<span>`, etc.
  - Attributes with string, number, and expression values
  - Go expressions inside `{}`
  - Comments: `//` and `/* */`

- **Language Configuration**
  - Auto-closing brackets and quotes
  - Comment toggling
  - Code folding
  - Smart indentation

## Installation

### From Source

1. Clone the go-tui repository:

   ```bash
   git clone https://github.com/grindlemire/go-tui.git
   ```

2. Copy the extension to your VS Code extensions folder:

   ```bash
   cp -r go-tui/editor/vscode ~/.vscode/extensions/gsx-language
   ```

3. Reload VS Code

### From VSIX (Package)

1. Build the extension:

   ```bash
   cd go-tui/editor/vscode
   npm install
   npx vsce package
   ```

2. Install the `.vsix` file:
   - Open VS Code
   - Go to Extensions (Ctrl+Shift+X)
   - Click the `...` menu and select "Install from VSIX..."
   - Select the generated `.vsix` file

## Usage

Simply open any `.gsx` file and the syntax highlighting will be applied automatically.

### Example

```gsx
package main

import (
    "fmt"
)

templ Counter(count int) {
    <box border={tui.BorderSingle} padding={1}>
        <text>{fmt.Sprintf("Count: %d", count)}</text>
    </box>

    @if count > 0 {
        <text>Positive!</text>
    } @else {
        <text>Zero or negative</text>
    }
}
```

## Supported Constructs

| Construct | Example |
|-----------|---------|
| Component | `templ Name(params) { ... }` |
| For loop | `@for i, v := range items { ... }` |
| If/Else | `@if condition { ... } @else { ... }` |
| Let binding | `@let x = <element>` |
| Component call | `@ComponentName(args)` |
| Element | `<box attr={value}>children</box>` |
| Go expression | `{fmt.Sprintf(...)}` |

## LSP Support

For advanced features like go-to-definition, hover, and auto-completion, run the GSX language server:

```bash
tui lsp
```

Configure your VS Code settings to use the language server (coming soon).

## Contributing

Contributions are welcome! Please see the [go-tui repository](https://github.com/grindlemire/go-tui) for contribution guidelines.

## License

MIT License - see the [LICENSE](https://github.com/grindlemire/go-tui/blob/main/LICENSE) file for details.
