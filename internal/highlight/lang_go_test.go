package highlight

import (
	"reflect"
	"testing"
)

func TestLexGo(t *testing.T) {
	type tc struct {
		code string
		want [][]Token
	}
	tests := map[string]tc{
		"keyword string number comment": {
			code: `var x = 1 // c`,
			want: [][]Token{{
				{KindKeyword, "var"},
				{KindPlain, " "},
				{KindPlain, "x"},
				{KindPlain, " "},
				{KindOperator, "="},
				{KindPlain, " "},
				{KindNumber, "1"},
				{KindPlain, " "},
				{KindComment, "// c"},
			}},
		},
		"func name is KindType, param type stays plain": {
			// "Foo" is typed (follows func); "int" is lowercase, not a keyword,
			// and not followed by '(', so the heuristics leave it KindPlain.
			code: `func Foo(n int)`,
			want: [][]Token{{
				{KindKeyword, "func"},
				{KindPlain, " "},
				{KindType, "Foo"},
				{KindOperator, "("},
				{KindPlain, "n"},
				{KindPlain, " "},
				{KindPlain, "int"},
				{KindOperator, ")"},
			}},
		},
		"raw string spans lines": {
			code: "`a\nb`",
			want: [][]Token{
				{{KindString, "`a"}},
				{{KindString, "b`"}},
			},
		},
		"block comment spans lines": {
			code: "/* a\nb */x",
			want: [][]Token{
				{{KindComment, "/* a"}},
				{{KindComment, "b */"}, {KindPlain, "x"}},
			},
		},
		"literal nil": {
			code: `nil`,
			want: [][]Token{{{KindLiteral, "nil"}}},
		},
		"method receiver is not a type": {
			code: `func (r *T) M()`,
			want: [][]Token{{
				{KindKeyword, "func"},
				{KindPlain, " "},
				{KindOperator, "("},
				{KindPlain, "r"},
				{KindPlain, " "},
				{KindOperator, "*"},
				{KindType, "T"},
				{KindOperator, ")"},
				{KindPlain, " "},
				{KindType, "M"},
				{KindOperator, "("},
				{KindOperator, ")"},
			}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := Tokenize("go", tt.code)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
