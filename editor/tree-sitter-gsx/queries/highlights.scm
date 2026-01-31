; Tree-sitter highlight queries for GSX DSL
; This file maps tree-sitter node types to highlight groups

; ====================
; Keywords
; ====================

; Package and import keywords
"package" @keyword
"import" @keyword.import

; DSL keywords
"templ" @keyword.function
"@for" @keyword.repeat
"@if" @keyword.conditional
"@else" @keyword.conditional
"@let" @keyword

; Function keyword
"func" @keyword.function

; Range keyword
"range" @keyword.repeat

; ====================
; Components
; ====================

; Component declaration name
(component_declaration
  name: (identifier) @function.definition)

; Component call
(component_call
  "@" @punctuation.special
  name: (identifier) @function.call)

; ====================
; Elements (Tags)
; ====================

; Tag names in elements (using identifier directly in simplified grammar)
(self_closing_element
  tag: (identifier) @tag)

(element_with_children
  tag: (identifier) @tag)

; Closing tags
(element_with_children
  (identifier) @tag)

; Tag delimiters
"<" @tag.delimiter
">" @tag.delimiter
"</" @tag.delimiter
"/" @tag.delimiter

; ====================
; Attributes
; ====================

(attribute
  name: (identifier) @property)

; ====================
; State Declarations
; ====================

; State variable name in declaration
(state_declaration
  name: (identifier) @variable.definition)

; ====================
; Functions
; ====================

; Function declaration name
(function_declaration
  name: (identifier) @function.definition)

; Function calls in expressions
(call_expression
  (identifier) @function.call)

(call_expression
  (selector_expression
    (identifier) @function.method.call))

; ====================
; Types
; ====================

(type_expression
  (identifier) @type)

(qualified_type
  (identifier) @module
  (identifier) @type)

(slice_type
  "[" @punctuation.bracket
  "]" @punctuation.bracket)

; ====================
; Variables & Parameters
; ====================

; Parameter names
(parameter
  name: (identifier) @variable.parameter)

; For loop variables
(for_clause
  index: (identifier) @variable)

(for_clause
  value: (identifier) @variable)

; Let binding name
(let_binding
  name: (identifier) @variable)

; ====================
; Identifiers
; ====================

; Selector expressions
(selector_expression
  (identifier) @variable
  (identifier) @property)

; Generic identifiers
(identifier) @variable

; ====================
; Literals
; ====================

; Strings
(string) @string

; Numbers
(number) @number

; Booleans
(true) @boolean
(false) @boolean

; ====================
; Operators
; ====================

":=" @operator
"=" @operator
"==" @operator
"!=" @operator
"<" @operator
"<=" @operator
">" @operator
">=" @operator
"+" @operator
"-" @operator
"*" @operator
"/" @operator
"&&" @operator
"||" @operator

; ====================
; Punctuation
; ====================

"(" @punctuation.bracket
")" @punctuation.bracket
"[" @punctuation.bracket
"]" @punctuation.bracket
"{" @punctuation.bracket
"}" @punctuation.bracket

"." @punctuation.delimiter
"," @punctuation.delimiter

; ====================
; Go Expressions
; ====================

; Expression content inside braces
(go_expression
  "{" @punctuation.special
  "}" @punctuation.special)

(expression_content) @embedded

; ====================
; Imports
; ====================

(import_spec
  alias: (identifier)? @module
  path: (string) @string.special)

; ====================
; Comments
; ====================

(comment) @comment
