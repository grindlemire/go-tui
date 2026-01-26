# Comments Support Implementation Plan

Implementation phases for comment support in .tui files. Each phase builds on the previous and has clear acceptance criteria.

---

## Phase 1: Lexer Comment Collection

**Reference:** [comments-design.md §6](./comments-design.md#6-lexer-changes)

**Completed in commit:** (pending)

- [ ] Modify `pkg/tuigen/token.go`
  - Add `TokenLineComment` and `TokenBlockComment` constants to TokenType enum
  - Add entries to `tokenNames` map for debugging

- [ ] Add `Comment` type to `pkg/tuigen/ast.go`
  - Define `Comment` struct with `Text`, `Position`, `EndLine`, `EndCol`, `IsBlock` fields
  - Define `CommentGroup` struct with `List []*Comment`
  - Add `Text()` method to `CommentGroup` that strips comment markers
  - See [comments-design.md §3](./comments-design.md#3-core-entities)

- [ ] Modify `pkg/tuigen/lexer.go`
  - Add `pendingComments []*Comment` field to `Lexer` struct
  - Rename `skipWhitespaceAndComments()` to `skipWhitespace()` (remove comment handling)
  - Add `collectComment()` method that reads comment and appends to `pendingComments`
  - Add `ConsumeComments() []*Comment` method that returns and clears pending comments
  - Modify `Next()` to call `collectComment()` when encountering `//` or `/*`
  - Ensure `collectComment()` correctly captures start/end positions for block comments
  - Handle unterminated block comment error

- [ ] Create `pkg/tuigen/lexer_comment_test.go`
  - Test line comment collection (`// comment`)
  - Test block comment collection (`/* comment */`)
  - Test multi-line block comment with correct EndLine/EndCol
  - Test unterminated block comment error
  - Test comments between tokens are collected
  - Test `ConsumeComments()` clears the buffer

**Tests:** Run `go test ./pkg/tuigen/... -run Comment` once at phase end

---

## Phase 2: AST Types and Parser Attachment

**Reference:** [comments-design.md §3](./comments-design.md#3-core-entities), [comments-design.md §4](./comments-design.md#4-comment-association-algorithm)

**Completed in commit:** (pending)

- [ ] Add comment fields to AST node structs in `pkg/tuigen/ast.go`
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

- [ ] Add comment association helpers to `pkg/tuigen/parser.go`
  - Add `pendingComments []*Comment` field to `Parser` struct
  - Add `collectPendingComments()` helper that calls `lexer.ConsumeComments()` and groups adjacent comments
  - Add `groupComments(comments []*Comment) []*CommentGroup` helper that groups by blank lines
  - Add `attachLeadingComments(node, comments)` helper
  - Add `attachTrailingComment(node, line int)` helper that checks if comment is on same line

- [ ] Integrate comment attachment into parser methods
  - `ParseFile()`: Collect leading comments before package, attach to File; collect orphans after last declaration
  - `parseComponent()`: Attach leading comments before `@component`; collect body orphans
  - `parseElement()`: Attach leading comments; check for trailing after `>`
  - `parseIf()`: Attach leading comments; collect orphans in empty then/else blocks
  - `parseFor()`: Attach leading comments; collect orphans in empty body
  - `parseLet()`: Attach leading comments
  - `parseComponentCall()`: Attach leading comments
  - `parseGoFunc()`: Attach leading comments
  - `parseGoStatement()`: Attach leading comments (for GoCode nodes)
  - `parseGoExprOrChildrenSlot()`: Attach leading/trailing comments

- [ ] Create `pkg/tuigen/parser_comment_test.go`
  - Test leading comment attachment to component
  - Test trailing comment attachment on same line
  - Test orphan comment storage in component body
  - Test orphan comment storage in file
  - Test comment grouping (adjacent vs blank-line separated)
  - Test comments in empty @if/@for bodies
  - Test comments at end of file

**Tests:** Run `go test ./pkg/tuigen/... -run Comment` once at phase end

---

## Phase 3: Formatter Comment Printing

**Reference:** [comments-design.md §7](./comments-design.md#7-user-experience)

**Completed in commit:** (pending)

- [ ] Add comment printing helpers to `pkg/formatter/printer.go`
  - Add `printCommentGroup(cg *tuigen.CommentGroup)` method
  - Add `printLeadingComments(cg *tuigen.CommentGroup)` method (prints with trailing newline)
  - Add `printTrailingComment(cg *tuigen.CommentGroup)` method (prints with leading spaces, no newline)
  - Add `printOrphanComments(groups []*tuigen.CommentGroup)` method

- [ ] Integrate comment printing into existing print methods
  - `PrintFile()`: Print File.LeadingComments before package; print File.OrphanComments at end
  - `printComponent()`: Print leading comments; print trailing after `{`; print orphans before `}`
  - `printElement()`: Print leading comments; print trailing after `>`
  - `printForLoop()`: Print leading comments; print orphans in body
  - `printIfStmt()`: Print leading comments; print orphans in then/else
  - `printLetBinding()`: Print leading comments
  - `printComponentCall()`: Print leading comments
  - `printGoFunc()`: Print leading comments
  - `printNode()`: Handle GoCode, GoExpr, ChildrenSlot leading/trailing comments

- [ ] Create `pkg/formatter/formatter_comment_test.go`
  - Test round-trip preservation of leading comments
  - Test round-trip preservation of trailing comments
  - Test round-trip preservation of orphan comments
  - Test round-trip with multiple comment groups (blank line preservation)
  - Test format idempotency: `format(format(src)) == format(src)`
  - Test complex file with comments at all positions

**Tests:** Run `go test ./pkg/formatter/... -run Comment` once at phase end

---

## Phase 4: LSP Semantic Tokens

**Reference:** [comments-design.md §8](./comments-design.md#8-lsp-integration)

**Completed in commit:** (pending)

- [ ] Modify `pkg/lsp/semantic_tokens.go`
  - Add `tokenTypeComment = 12` constant
  - Add `collectCommentGroupTokens(cg *tuigen.CommentGroup, tokens *[]semanticToken)` helper
  - Add `collectNodeCommentTokens(node tuigen.Node, tokens *[]semanticToken)` helper that extracts comments from any node
  - Add `collectAllCommentTokens(file *tuigen.File, tokens *[]semanticToken)` that walks AST collecting all comments

- [ ] Integrate comment tokens into `collectSemanticTokens()`
  - Call `collectAllCommentTokens()` to add comment tokens to result
  - Ensure tokens are sorted by position (existing sort handles this)

- [ ] Update LSP server capabilities in `pkg/lsp/server.go`
  - Add "comment" to `SemanticTokensLegend.TokenTypes` array (index 12)

- [ ] Create `pkg/lsp/semantic_tokens_comment_test.go`
  - Test line comment emits correct token
  - Test block comment emits correct token with correct length
  - Test multi-line block comment position/length
  - Test comments at various positions (leading, trailing, orphan) all emit tokens

**Tests:** Run `go test ./pkg/lsp/... -run Comment` once at phase end

---

## Phase Summary

| Phase | Description | Status |
|-------|-------------|--------|
| 1 | Lexer comment collection (tokens, Comment type, collection logic) | Pending |
| 2 | AST types and parser attachment (comment fields, association algorithm) | Pending |
| 3 | Formatter comment printing (round-trip preservation) | Pending |
| 4 | LSP semantic tokens (syntax highlighting for comments) | Pending |

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
