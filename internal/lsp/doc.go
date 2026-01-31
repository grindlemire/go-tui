// Package lsp implements a Language Server Protocol server for .gsx files.
//
// It provides diagnostics, completion, hover, go-to-definition, references,
// semantic tokens, and formatting. The server communicates over stdio using
// JSON-RPC 2.0 and delegates Go-specific intelligence to a gopls subprocess.
package lsp
