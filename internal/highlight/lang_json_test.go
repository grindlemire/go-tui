package highlight

import (
	"reflect"
	"testing"
)

func TestLexJSON(t *testing.T) {
	type tc struct {
		code string
		want [][]Token
	}
	tests := map[string]tc{
		"key vs string value": {
			code: `{"a": "b"}`,
			want: [][]Token{{
				{KindOperator, "{"},
				{KindKey, `"a"`},
				{KindOperator, ":"},
				{KindPlain, " "},
				{KindString, `"b"`},
				{KindOperator, "}"},
			}},
		},
		"number and literal": {
			code: `{"n": -1, "ok": true}`,
			want: [][]Token{{
				{KindOperator, "{"},
				{KindKey, `"n"`},
				{KindOperator, ":"},
				{KindPlain, " "},
				{KindNumber, "-1"},
				{KindOperator, ","},
				{KindPlain, " "},
				{KindKey, `"ok"`},
				{KindOperator, ":"},
				{KindPlain, " "},
				{KindLiteral, "true"},
				{KindOperator, "}"},
			}},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := Tokenize("json", tt.code)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %#v, want %#v", got, tt.want)
			}
		})
	}
}
