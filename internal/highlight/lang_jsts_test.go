package highlight

import (
	"reflect"
	"testing"
)

func TestLexJS(t *testing.T) {
	type tc struct {
		code string
		want [][]Token
	}
	tests := map[string]tc{
		"const string comment": {
			code: `const s = "x" // c`,
			want: [][]Token{{
				{KindKeyword, "const"},
				{KindPlain, " "},
				{KindPlain, "s"},
				{KindPlain, " "},
				{KindOperator, "="},
				{KindPlain, " "},
				{KindString, `"x"`},
				{KindPlain, " "},
				{KindComment, "// c"},
			}},
		},
		"function name and class name are KindType": {
			code: `function foo() class Bar`,
			want: [][]Token{{
				{KindKeyword, "function"},
				{KindPlain, " "},
				{KindType, "foo"},
				{KindOperator, "("},
				{KindOperator, ")"},
				{KindPlain, " "},
				{KindKeyword, "class"},
				{KindPlain, " "},
				{KindType, "Bar"},
			}},
		},
		"template literal spans lines": {
			code: "`a\nb`",
			want: [][]Token{
				{{KindString, "`a"}},
				{{KindString, "b`"}},
			},
		},
		"literal null": {
			code: `null`,
			want: [][]Token{{{KindLiteral, "null"}}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := Tokenize("js", tt.code)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
