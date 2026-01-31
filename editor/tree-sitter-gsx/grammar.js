/**
 * Tree-sitter grammar for GSX DSL (.gsx files)
 */

module.exports = grammar({
  name: 'gsx',

  extras: $ => [
    /\s/,
    $.comment,
  ],

  word: $ => $.identifier,

  // state_declaration (name := call_expression) overlaps with _expression since
  // "identifier := call_expression" could be parsed as either. Tree-sitter resolves
  // this by preferring the longer match (state_declaration) when both apply.
  // This is intentional: inside a component body, `x := tui.NewState(0)` should
  // parse as a state_declaration, not as an expression.
  conflicts: $ => [
    [$.state_declaration, $._expression],
  ],

  rules: {
    source_file: $ => seq(
      optional($.package_clause),
      optional($.import_section),
      repeat(choice($.component_declaration, $.function_declaration)),
    ),

    // Package clause
    package_clause: $ => seq('package', field('name', $.identifier)),

    // Import section
    import_section: $ => repeat1($.import_declaration),

    import_declaration: $ => seq(
      'import',
      choice($.import_spec, $.import_spec_list),
    ),

    import_spec_list: $ => seq('(', repeat($.import_spec), ')'),

    import_spec: $ => seq(
      optional(field('alias', $.identifier)),
      field('path', $.string),
    ),

    // Component declaration: templ Name(params) { body }
    component_declaration: $ => seq(
      'templ',
      field('name', $.identifier),
      field('parameters', $.parameter_list),
      field('body', $.component_body),
    ),

    parameter_list: $ => seq(
      '(',
      optional(seq(
        $.parameter,
        repeat(seq(',', $.parameter)),
      )),
      ')',
    ),

    parameter: $ => seq(
      field('name', $.identifier),
      field('type', $.type_expression),
    ),

    type_expression: $ => choice(
      $.identifier,
      $.qualified_type,
      $.slice_type,
    ),

    qualified_type: $ => seq($.identifier, '.', $.identifier),
    slice_type: $ => seq('[', ']', $.type_expression),

    component_body: $ => seq('{', repeat($._child), '}'),

    _child: $ => choice(
      $.element,
      $.for_statement,
      $.if_statement,
      $.let_binding,
      $.state_declaration,
      $.component_call,
      $.go_expression,
    ),

    // Elements
    element: $ => choice($.self_closing_element, $.element_with_children),

    self_closing_element: $ => seq(
      '<',
      field('tag', $.identifier),
      repeat($.attribute),
      '/',
      '>',
    ),

    element_with_children: $ => seq(
      '<',
      field('tag', $.identifier),
      repeat($.attribute),
      '>',
      repeat($._element_child),
      '</',
      $.identifier,
      '>',
    ),

    _element_child: $ => choice(
      $.element,
      $.for_statement,
      $.if_statement,
      $.let_binding,
      $.component_call,
      $.go_expression,
      $.text_content,
    ),

    text_content: $ => /[^<>{}@\s][^<>{}@\n]*/,

    attribute: $ => seq(
      field('name', $.identifier),
      optional(seq('=', field('value', $._attribute_value))),
    ),

    _attribute_value: $ => choice($.string, $.go_expression, $.number),

    // Control flow
    for_statement: $ => seq(
      '@for',
      field('clause', $.for_clause),
      field('body', $.block),
    ),

    for_clause: $ => seq(
      optional(seq(field('index', $.identifier), ',')),
      field('value', $.identifier),
      ':=',
      'range',
      field('collection', $._expression),
    ),

    if_statement: $ => seq(
      '@if',
      field('condition', $._expression),
      field('consequence', $.block),
      optional(seq('@else', field('alternative', choice($.block, $.if_statement)))),
    ),

    let_binding: $ => seq(
      '@let',
      field('name', $.identifier),
      '=',
      field('value', choice($.element, $.go_expression)),
    ),

    // State declaration: name := tui.NewState(value)
    state_declaration: $ => seq(
      field('name', $.identifier),
      ':=',
      field('initializer', $.call_expression),
    ),

    component_call: $ => seq(
      '@',
      field('name', $.identifier),
      field('arguments', $.argument_list),
    ),

    block: $ => seq('{', repeat($._child), '}'),

    // Go expressions: content between balanced braces, including nested braces
    // and string literals (which may contain unbalanced braces).
    go_expression: $ => seq('{', $.expression_content, '}'),

    expression_content: $ => repeat1(choice(
      $.go_string_literal,  // Handle string literals so braces inside them don't confuse parsing
      /[^{}"'`]+/,          // Non-brace, non-quote content
      $.nested_braces,
    )),

    // String literals inside Go expressions — captures "...", '...', `...` to prevent
    // their contents (which may include braces) from being parsed as structure.
    go_string_literal: $ => choice(
      /"[^"\\]*(?:\\.[^"\\]*)*"/,   // Double-quoted string with escape handling
      /`[^`]*`/,                     // Backtick (raw) string
    ),

    nested_braces: $ => seq('{', repeat(choice(
      $.go_string_literal,
      /[^{}"'`]+/,
      $.nested_braces,
    )), '}'),

    // Expressions (simplified)
    _expression: $ => choice(
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

    binary_expression: $ => prec.left(1, seq(
      $._expression,
      choice('==', '!=', '<', '>', '<=', '>=', '+', '-', '*', '/', '&&', '||'),
      $._expression,
    )),

    call_expression: $ => prec(2, seq(
      choice($.identifier, $.selector_expression),
      $.argument_list,
    )),

    selector_expression: $ => prec(3, seq($._expression, '.', $.identifier)),

    parenthesized_expression: $ => seq('(', $._expression, ')'),

    argument_list: $ => seq(
      '(',
      optional(seq($._expression, repeat(seq(',', $._expression)))),
      ')',
    ),

    // Literals
    identifier: $ => /[a-zA-Z_][a-zA-Z0-9_]*/,
    number: $ => /\d+(\.\d+)?/,
    string: $ => /"[^"]*"/,
    true: $ => 'true',
    false: $ => 'false',

    // Function declarations (helper functions)
    function_declaration: $ => seq(
      'func',
      field('name', $.identifier),
      field('parameters', $.parameter_list),
      optional(field('return_type', $.type_expression)),
      field('body', $.function_body),
    ),

    function_body: $ => seq('{', repeat(choice($.go_string_literal, /[^{}"'`]+/, $.nested_braces)), '}'),

    // Comments (for explicit use in the AST)
    comment: $ => choice(/\/\/.*/, /\/\*[^*]*\*+([^/*][^*]*\*+)*\//),
  },
});
