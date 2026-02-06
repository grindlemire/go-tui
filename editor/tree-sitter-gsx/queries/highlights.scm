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

; Go declaration keywords
"type" @keyword
"var" @keyword
"const" @keyword

; ====================
; Components
; ====================

; Component declaration name
(component_declaration
  name: (identifier) @function.definition)

; Method component receiver
(component_declaration
  receiver: (receiver
    name: (identifier) @variable.parameter
    type: (type_expression) @type))

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

; Tag delimiters (context-specific to avoid operator conflicts)
(self_closing_element "<" @tag.delimiter)
(self_closing_element "/" @tag.delimiter)
(self_closing_element ">" @tag.delimiter)

(element_with_children "<" @tag.delimiter)
(element_with_children ">" @tag.delimiter)
(element_with_children "</" @tag.delimiter)

; ====================
; Attributes
; ====================

(attribute
  name: (identifier) @property)

; ====================
; Struct Declarations
; ====================

; Struct type name
(type_struct_declaration
  name: (identifier) @type.definition)

; Struct keyword
"struct" @keyword

; Struct field name and type
(struct_field
  name: (identifier) @property.definition)

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

; Method function receiver
(function_declaration
  receiver: (receiver
    name: (identifier) @variable.parameter
    type: (type_expression) @type))

; Function calls in expressions
(call_expression
  (identifier) @function.call)

(call_expression
  (selector_expression
    (_)
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

(pointer_type
  "*" @operator)

(generic_type
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

; Package name
(package_clause
  name: (identifier) @module)

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

; These can remain bare (no conflicts)
":=" @operator
"=" @operator

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
; Children Slot
; ====================

; {children...} placeholder
(children_slot
  "{" @punctuation.special
  "children" @keyword
  "..." @punctuation.special
  "}" @punctuation.special)

; ====================
; Go Expressions
; ====================

; Expression content inside braces
(go_expression
  "{" @punctuation.special
  "}" @punctuation.special)

(expression_content) @embedded

; Go code content inside function/brace bodies
(go_code_content) @embedded

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
