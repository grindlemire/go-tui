/**
 * Tree-sitter grammar for GSX DSL (.gsx files)
 */

module.exports = grammar({
  name: "gsx",

  extras: ($) => [/\s/, $.comment],

  word: ($) => $.identifier,

  // state_declaration (name := call_expression) overlaps with _expression since
  // "identifier := call_expression" could be parsed as either. Tree-sitter resolves
  // this by preferring the longer match (state_declaration) when both apply.
  // This is intentional: inside a component body, `x := tui.NewState(0)` should
  // parse as a state_declaration, not as an expression.
  conflicts: ($) => [[$.state_declaration, $._expression]],

  rules: {
    source_file: ($) =>
      seq(
        optional($.package_clause),
        optional($.import_section),
        repeat(
          choice(
            $.component_declaration,
            $.function_declaration,
            $.type_struct_declaration,
            $.go_declaration,
          ),
        ),
      ),

    // Package clause
    package_clause: ($) => seq("package", field("name", $.identifier)),

    // Import section
    import_section: ($) => repeat1($.import_declaration),

    import_declaration: ($) =>
      seq("import", choice($.import_spec, $.import_spec_list)),

    import_spec_list: ($) => seq("(", repeat($.import_spec), ")"),

    import_spec: ($) =>
      seq(optional(field("alias", $.identifier)), field("path", $.string)),

    // Component declaration: supports both function and method forms
    //   templ Name(params) { body }
    //   templ (s *Type) Render() { body }
    component_declaration: ($) =>
      choice(
        seq(
          "templ",
          field("name", $.identifier),
          field("parameters", $.parameter_list),
          field("body", $.component_body),
        ),
        seq(
          "templ",
          field("receiver", $.receiver),
          field("name", $.identifier),
          "(",
          ")",
          field("body", $.component_body),
        ),
      ),

    // Receiver for method declarations: (name *Type) or (name Type)
    receiver: ($) =>
      seq(
        "(",
        field("name", $.identifier),
        field("type", $.type_expression),
        ")",
      ),

    parameter_list: ($) =>
      seq("(", optional(seq($.parameter, repeat(seq(",", $.parameter)))), ")"),

    parameter: ($) =>
      seq(field("name", $.identifier), field("type", $.type_expression)),

    type_expression: ($) =>
      choice(
        $.identifier,
        $.qualified_type,
        $.slice_type,
        $.pointer_type,
        $.generic_type,
      ),

    qualified_type: ($) => seq($.identifier, ".", $.identifier),
    slice_type: ($) => seq("[", "]", $.type_expression),
    pointer_type: ($) =>
      seq("*", choice($.identifier, $.qualified_type, $.generic_type)),
    // Generic type: Type[T] or pkg.Type[T]
    generic_type: ($) =>
      seq(
        choice($.identifier, $.qualified_type),
        "[",
        $.type_expression,
        repeat(seq(",", $.type_expression)),
        "]",
      ),

    // Struct type declaration: type Name struct { fields }
    // Uses prec(1) to take priority over the opaque go_declaration for struct types.
    type_struct_declaration: ($) =>
      prec(
        1,
        seq(
          "type",
          field("name", $.identifier),
          "struct",
          field("body", $.struct_body),
        ),
      ),

    struct_body: ($) => seq("{", repeat($.struct_field), "}"),

    struct_field: ($) =>
      seq(field("name", $.identifier), field("type", $.type_expression)),

    component_body: ($) => seq("{", repeat($._child), "}"),

    _child: ($) =>
      choice(
        $.element,
        $.for_statement,
        $.if_statement,
        $.let_binding,
        $.state_declaration,
        $.component_call,
        $.children_slot,
        $.go_expression,
      ),

    // Elements
    element: ($) => choice($.self_closing_element, $.element_with_children),

    self_closing_element: ($) =>
      seq("<", field("tag", $.identifier), repeat($.attribute), "/", ">"),

    element_with_children: ($) =>
      seq(
        "<",
        field("tag", $.identifier),
        repeat($.attribute),
        ">",
        repeat($._element_child),
        "</",
        $.identifier,
        ">",
      ),

    _element_child: ($) =>
      choice(
        $.element,
        $.for_statement,
        $.if_statement,
        $.let_binding,
        $.component_call,
        $.children_slot,
        $.go_expression,
        $.text_content,
      ),

    text_content: ($) => /[^<>{}@\s][^<>{}@\n]*/,

    attribute: ($) =>
      seq(
        field("name", $.identifier),
        optional(seq("=", field("value", $._attribute_value))),
      ),

    _attribute_value: ($) => choice($.string, $.go_expression, $.number),

    // Control flow
    for_statement: ($) =>
      seq("@for", field("clause", $.for_clause), field("body", $.block)),

    for_clause: ($) =>
      seq(
        optional(seq(field("index", $.identifier), ",")),
        field("value", $.identifier),
        ":=",
        "range",
        field("collection", $._expression),
      ),

    if_statement: ($) =>
      seq(
        "@if",
        field("condition", $._expression),
        field("consequence", $.block),
        optional(
          seq("@else", field("alternative", choice($.block, $.if_statement))),
        ),
      ),

    let_binding: ($) =>
      seq(
        "@let",
        field("name", $.identifier),
        "=",
        field("value", choice($.element, $.go_expression)),
      ),

    // State declaration: name := tui.NewState(value)
    state_declaration: ($) =>
      seq(
        field("name", $.identifier),
        ":=",
        field("initializer", $.call_expression),
      ),

    // Component call: @Name(args) or @Name(args) { children }
    // prec(1) ensures that a '{' after the argument list is parsed as a
    // children block rather than a separate go_expression sibling.
    component_call: ($) =>
      prec.right(
        1,
        seq(
          "@",
          field("name", $.identifier),
          field("arguments", $.argument_list),
          optional(field("children", $.block)),
        ),
      ),

    // Children slot: {children...}
    children_slot: ($) => seq("{", "children", "...", "}"),

    block: ($) => seq("{", repeat($._child), "}"),

    // Go expressions: content between balanced braces, including nested braces
    // and string literals (which may contain unbalanced braces).
    go_expression: ($) => seq("{", $.expression_content, "}"),

    expression_content: ($) =>
      repeat1(
        choice(
          $.go_string_literal, // Handle string literals so braces inside them don't confuse parsing
          /[^{}"'`]+/, // Non-brace, non-quote content
          $.nested_braces,
        ),
      ),

    // String and rune literals inside Go expressions — captures "...", '...', `...`
    // to prevent their contents (which may include braces) from being parsed as structure.
    go_string_literal: ($) =>
      choice(
        /"[^"\\]*(?:\\.[^"\\]*)*"/, // Double-quoted string with escape handling
        /`[^`]*`/, // Backtick (raw) string
        /'[^'\\]'/, // Simple rune literal: 'a', '/', etc.
        /'\\.'/, // Escaped rune literal: '\n', '\\', '\'', etc.
      ),

    nested_braces: ($) =>
      seq(
        "{",
        repeat(choice($.go_string_literal, /[^{}"'`]+/, $.nested_braces)),
        "}",
      ),

    // Expressions (simplified)
    _expression: ($) =>
      choice(
        $.identifier,
        $.number,
        $.string,
        $.true,
        $.false,
        $.binary_expression,
        $.call_expression,
        $.selector_expression,
        $.parenthesized_expression,
      ),

    binary_expression: ($) =>
      prec.left(
        1,
        seq(
          $._expression,
          choice(
            "==",
            "!=",
            "<",
            ">",
            "<=",
            ">=",
            "+",
            "-",
            "*",
            "/",
            "&&",
            "||",
          ),
          $._expression,
        ),
      ),

    call_expression: ($) =>
      prec(
        2,
        seq(choice($.identifier, $.selector_expression), $.argument_list),
      ),

    selector_expression: ($) => prec(3, seq($._expression, ".", $.identifier)),

    parenthesized_expression: ($) => seq("(", $._expression, ")"),

    argument_list: ($) =>
      seq(
        "(",
        optional(seq($._expression, repeat(seq(",", $._expression)))),
        ")",
      ),

    // Literals
    identifier: ($) => /[a-zA-Z_][a-zA-Z0-9_]*/,
    number: ($) => /\d+(\.\d+)?/,
    string: ($) => /"[^"]*"/,
    true: ($) => "true",
    false: ($) => "false",

    // Function declarations: supports both regular and method forms
    //   func name(params) returnType { body }
    //   func (recv) name(params) returnType { body }
    function_declaration: ($) =>
      choice(
        seq(
          "func",
          field("name", $.identifier),
          field("parameters", $.parameter_list),
          optional(field("return_type", $.type_expression)),
          field("body", $.function_body),
        ),
        seq(
          "func",
          field("receiver", $.receiver),
          field("name", $.identifier),
          field("parameters", $.parameter_list),
          optional(field("return_type", $.type_expression)),
          field("body", $.function_body),
        ),
      ),

    // Go declaration: type, var, or const blocks (captured as opaque Go code)
    // Handles non-struct type declarations, grouped declarations, and interface checks.
    // Note: struct declarations are handled by type_struct_declaration above.
    // The preamble uses choice(identifier, punctuation_regex) instead of a single greedy
    // regex so that Go keywords like "struct" are tokenized separately and not consumed
    // as part of the opaque preamble text.
    go_declaration: ($) =>
      choice(
        // Braced form: type Name interface { ... }, var/const ... { ... }
        seq(
          field("keyword", choice("type", "var", "const")),
          field(
            "preamble",
            repeat1(choice($.identifier, /[^\sa-zA-Z_0-9{()"'`\n]+/)),
          ),
          field("body", $.go_brace_body),
        ),
        // Parenthesized form: var ( ... ) or const ( ... )
        seq(
          field("keyword", choice("type", "var", "const")),
          field("body", $.go_paren_body),
        ),
        // Single-line form: var _ tui.Component = (*foo)(nil)
        seq(
          field("keyword", choice("type", "var", "const")),
          field(
            "preamble",
            repeat1(choice($.identifier, /[^\sa-zA-Z_0-9{()"'`\n]+/)),
          ),
          // Parenthesized expressions in single-line form
          optional(seq("(", /[^)]*/, ")")),
          optional(seq("(", /[^)]*/, ")")),
        ),
      ),

    // Brace-delimited body (for interface, etc. — struct handled by type_struct_declaration)
    go_brace_body: ($) => seq("{", optional($.go_code_content), "}"),

    // Parenthesis-delimited body (for grouped var/const/type)
    go_paren_body: ($) =>
      seq(
        "(",
        repeat(choice($.go_string_literal, /[^()"'`]+/, $.nested_parens)),
        ")",
      ),

    nested_parens: ($) =>
      seq(
        "(",
        repeat(choice($.go_string_literal, /[^()"'`]+/, $.nested_parens)),
        ")",
      ),

    function_body: ($) => seq("{", optional($.go_code_content), "}"),

    // Named node for Go code content inside function/brace bodies.
    // Having a named node enables Go language injection for syntax highlighting.
    go_code_content: ($) =>
      repeat1(choice($.go_string_literal, /[^{}"'`]+/, $.nested_braces)),

    // Comments (for explicit use in the AST)
    comment: ($) => choice(/\/\/.*/, /\/\*[^*]*\*+([^/*][^*]*\*+)*\//),
  },
});
