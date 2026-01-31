// Package tuigen compiles .gsx template files into type-safe Go code.
//
// The pipeline consists of:
//   - [Lexer]: tokenizes .gsx source into a token stream
//   - [Parser]: builds an AST from the token stream
//   - [Analyzer]: performs semantic analysis (imports, refs, state bindings)
//   - [Generator]: emits Go source code from the analyzed AST
//
// Additionally, [ParseTailwindClasses] translates Tailwind-style class strings
// into tui element options.
package tuigen
