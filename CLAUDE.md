# Project Guidelines

## Git Commits IMPORTANT

Use `gcommit -m ""` for all commits to ensure proper signing.

ONLY EVER COMMIT USING THIS APPROACH

## Testing

Use table-driven tests for all unit tests with the following pattern:

```go
type tc struct {
    // test case fields
}

tests := map[string]tc{
    "test name": {
        // test case values
    },
}

for name, tt := range tests {
    t.Run(name, func(t *testing.T) {
        // test logic
    })
}
```

Always define the `tc` struct separately before the test map.
