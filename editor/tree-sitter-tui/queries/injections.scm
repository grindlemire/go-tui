; Tree-sitter injection queries for TUI DSL
; This file defines language injections for embedded code

; ====================
; Go Expression Injection
; ====================

; Inject Go language into expression content inside {braces}
; This enables Go syntax highlighting and intelligence within expressions
((go_expression
  (expression_content) @injection.content)
  (#set! injection.language "go")
  (#set! injection.include-children))

; ====================
; Import Path Injection
; ====================

; Import paths get string highlighting (already handled by Go if using full injection)
; But we mark them for semantic purposes
((import_spec
  path: (string) @injection.content)
  (#set! injection.language "go"))
