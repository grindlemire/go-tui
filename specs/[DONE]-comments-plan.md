# Comments Support Implementation Plan

Implementation phases for comment support in .tui files. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Lexer Comment Collection

**Reference:** [comments-design.md §6](./comments-design.md#6-lexer-changes)

**Status:** COMPLETE

- [x] Modify `pkg/tuigen/token.go`
  - Add `TokenLineComment` and `TokenBlockComment` constants to TokenType enum
  - Add entries to `tokenNames` map for debugging

- [x] Add `Comment` type to `pkg/tuigen/ast.go`
  - Define `Comment` struct with `Text`, `Position`, `EndLine`, `EndCol`, `IsBlock` fields
  - Define `CommentGroup` struct with `List []*Comment`
  - Add `Text()` method to `CommentGroup` that strips comment markers
  - See [comments-design.md §3](./comments-design.md#3-core-entities)

- [x] Modify `pkg/tuigen/lexer.go`
  - Add `pendingComments []*Comment` field to `Lexer` struct
  - Rename `skipWhitespaceAndComments()` to `skipWhitespaceAndCollectComments()` (now collects comments)
  - Add `collectLineComment()` method that reads line comment and appends to `pendingComments`
  - Add `collectBlockComment()` method that reads block comment and appends to `pendingComments`
  - Add `ConsumeComments() []*Comment` method that returns and clears pending comments
  - Modify `Next()` to call `skipWhitespaceAndCollectComments()` which collects comments
  - Ensure `collectBlockComment()` correctly captures start/end positions for block comments
  - Handle unterminated block comment error

- [x] Create `pkg/tuigen/lexer_comment_test.go`
  - Test line comment collection (`// comment`)
  - Test block comment collection (`/* comment */`)
  - Test multi-line block comment with correct EndLine/EndCol
  - Test unterminated block comment error
  - Test comments between tokens are collected
  - Test `ConsumeComments()` clears the buffer

**Tests:** All tests pass

---

## Phase 2: AST Types and Parser Attachment

**Reference:** [comments-design.md §3](./comments-design.md#3-core-entities), [comments-design.md §4](./comments-design.md#4-comment-association-algorithm)

**Status:** COMPLETE

- [x] Add comment fields to AST node structs in `pkg/tuigen/ast.go`
  - `File`: Add `LeadingComments *CommentGroup`, `OrphanComments []*CommentGroup`
  - `Component`: Add `LeadingComments *CommentGroup`, `TrailingComments *CommentGroup`, `OrphanComments []*CommentGroup`
  - `Element`: Add `LeadingComments *CommentGroup`, `TrailingComments *CommentGroup`
  - `IfStmt`: Add `LeadingComments *CommentGroup`, `TrailingComments *CommentGroup`, `OrphanComments []*CommentGroup`
  - `ForLoop`: Add `LeadingComments *CommentGroup`, `TrailingComments *CommentGroup`, `OrphanComments []*CommentGroup`
  - `LetBinding`: Add `LeadingComments *CommentGroup`, `TrailingComments *CommentGroup`
  - `ComponentCall`: Add `LeadingComments *CommentGroup`, `TrailingComments *CommentGroup`
  - `GoFunc`: Add `LeadingComments *CommentGroup`, `TrailingComments *CommentGroup`
  - `GoCode`: Add `LeadingComments *CommentGroup`, `TrailingComments *CommentGroup`
  - `GoExpr`: Add `LeadingComments *CommentGroup`, `TrailingComments *CommentGroup`
  - `ChildrenSlot`: Add `LeadingComments *CommentGroup`, `TrailingComments *CommentGroup`
  - `Import`: Add `TrailingComments *CommentGroup` (inline comment on import line)

- [x] Add comment association helpers to `pkg/tuigen/parser.go`
  - Add `pendingComments []*Comment` field to `Parser` struct
  - Add `collectPendingComments()` helper that calls `lexer.ConsumeComments()` and groups adjacent comments
  - Add `groupComments(comments []*Comment) []*CommentGroup` helper that groups by blank lines
  - Add `attachLeadingComments(node, comments)` helper
  - Add `getTrailingCommentOnLine(line int)` helper that checks if comment is on same line

- [x] Integrate comment attachment into parser methods
  - `ParseFile()`: Collect leading comments before package, attach to File; collect orphans after last declaration
  - `parseComponent()`: Attach leading comments before `@component`; collect body orphans
  - `parseElement()`: Attach leading comments; check for trailing after `>`
  - `parseIf()`: Attach leading comments; collect orphans in empty then/else blocks
  - `parseFor()`: Attach leading comments; collect orphans in empty body
  - `parseChildren()`: Attach leading comments to child nodes
  - `parseComponentBodyWithOrphans()`: Handle orphan comments in body

- [x] Create `pkg/tuigen/parser_comment_test.go`
  - Test leading comment attachment to component
  - Test trailing comment attachment on same line
  - Test orphan comment storage in component body
  - Test orphan comment storage in file
  - Test comment grouping (adjacent vs blank-line separated)
  - Test comments in empty @if/@for bodies
  - Test comments at end of file

**Tests:** All tests pass

---

## Phase 3: Formatter Comment Printing

**Reference:** [comments-design.md §7](./comments-design.md#7-user-experience)

**Status:** COMPLETE

- [x] Add comment printing helpers to `pkg/formatter/printer.go`
  - Add `printCommentGroup(cg *tuigen.CommentGroup)` method
  - Add `printLeadingComments(cg *tuigen.CommentGroup)` method (prints with trailing newline)
  - Add `printTrailingComment(cg *tuigen.CommentGroup)` method (prints with leading spaces, no newline)
  - Add `printOrphanComments(groups []*tuigen.CommentGroup)` method

- [x] Integrate comment printing into existing print methods
  - `PrintFile()`: Print File.LeadingComments before package; print File.OrphanComments at end
  - `printComponent()`: Print leading comments; print trailing after `{`; print orphans before `}`
  - `printElement()`: Print leading comments; print trailing after `>`
  - `printForLoop()`: Print leading comments; print orphans in body
  - `printIfStmt()`: Print leading comments; print orphans in then/else
  - `printLetBinding()`: Print leading comments
  - `printComponentCall()`: Print leading comments
  - `printGoFunc()`: Print leading comments
  - `printNode()`: Handle GoCode, GoExpr, ChildrenSlot leading/trailing comments

- [x] Create `pkg/formatter/formatter_comment_test.go`
  - Test round-trip preservation of leading comments
  - Test round-trip preservation of trailing comments
  - Test round-trip preservation of orphan comments
  - Test round-trip with multiple comment groups (blank line preservation)
  - Test format idempotency: `format(format(src)) == format(src)`
  - Test complex file with comments at all positions

**Tests:** All tests pass

---

## Phase 4: LSP Semantic Tokens

**Reference:** [comments-design.md §8](./comments-design.md#8-lsp-integration)

**Status:** COMPLETE

- [x] Modify `pkg/lsp/semantic_tokens.go`
  - Add `tokenTypeComment = 13` constant (index 13 since regexp is 12)
  - Add `collectCommentGroupTokens(cg *tuigen.CommentGroup, tokens *[]semanticToken)` helper
  - Add `collectCommentToken(c *tuigen.Comment, tokens *[]semanticToken)` helper for single comments
  - Add `collectNodeCommentTokens(node tuigen.Node, tokens *[]semanticToken)` helper that extracts comments from any node
  - Add `collectAllCommentTokens(file *tuigen.File, tokens *[]semanticToken)` that walks AST collecting all comments
  - Add `collectComponentCommentTokens(comp *tuigen.Component, tokens *[]semanticToken)` helper

- [x] Integrate comment tokens into `collectSemanticTokens()`
  - Call `collectAllCommentTokens()` to add comment tokens to result
  - Ensure tokens are sorted by position (existing sort handles this)

- [x] Update LSP server capabilities in `pkg/lsp/handler.go`
  - Add "comment" to `SemanticTokensLegend.TokenTypes` array (index 13)

- [x] Create `pkg/lsp/semantic_tokens_comment_test.go`
  - Test line comment emits correct token
  - Test block comment emits correct token with correct length
  - Test multi-line block comment position/length
  - Test comments at various positions (leading, trailing, orphan) all emit tokens

**Tests:** All tests pass

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Lexer comment collection (tokens, Comment type, collection logic) | **COMPLETE** |
| 2 | AST types and parser attachment (comment fields, association algorithm) | **COMPLETE** |
| 3 | Formatter comment printing (round-trip preservation) | **COMPLETE** |
| 4 | LSP semantic tokens (syntax highlighting for comments) | **COMPLETE** |

## Files to Create

```
pkg/tuigen/
└── lexer_comment_test.go
└── parser_comment_test.go

pkg/formatter/
└── formatter_comment_test.go

pkg/lsp/
└── semantic_tokens_comment_test.go
```

## Files to Modify

| File | Changes |
|------|---------|
| `pkg/tuigen/token.go` | Add TokenLineComment, TokenBlockComment |
| `pkg/tuigen/ast.go` | Add Comment, CommentGroup types; add comment fields to all node structs |
| `pkg/tuigen/lexer.go` | Add pendingComments, collectComment(), ConsumeComments(); modify Next() |
| `pkg/tuigen/parser.go` | Add comment attachment logic to all parse methods |
| `pkg/formatter/printer.go` | Add comment printing methods; integrate into all print methods |
| `pkg/lsp/semantic_tokens.go` | Add tokenTypeComment; add comment token collection |
| `pkg/lsp/server.go` | Add "comment" to SemanticTokensLegend |
