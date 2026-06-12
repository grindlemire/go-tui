package formatter

import (
	"reflect"
	"testing"
)

// TestFormatControlFlowCoverage exercises control-flow printing paths:
// else-if chains (printElseBranch), let bindings to component calls,
// component expressions, elements with refs/attributes, self-closing
// elements, multi-line children, and multi-line component call args.
func TestFormatControlFlowCoverage(t *testing.T) {
	type tc struct {
		input string
		want  string
	}

	tests := map[string]tc{
		"else if without final else": {
			input: `package main

templ A(x int) {
if x == 1 {
<span>one</span>
} else if x == 2 {
<span>two</span>
}
}
`,
			want: `package main

templ A(x int) {
	if x == 1 {
		<span>one</span>
	} else if x == 2 {
		<span>two</span>
	}
}
`,
		},
		"else if chain with final else": {
			input: `package main

templ A(x int) {
if x == 1 {
<span>one</span>
} else if x == 2 {
<span>two</span>
} else if x == 3 {
<span>three</span>
} else {
<span>many</span>
}
}
`,
			want: `package main

templ A(x int) {
	if x == 1 {
		<span>one</span>
	} else if x == 2 {
		<span>two</span>
	} else if x == 3 {
		<span>three</span>
	} else {
		<span>many</span>
	}
}
`,
		},
		"else if chain without final else": {
			input: `package main

templ A(x int) {
if x == 1 {
<span>one</span>
} else if x == 2 {
<span>two</span>
} else if x == 3 {
<span>three</span>
}
}
`,
			want: `package main

templ A(x int) {
	if x == 1 {
		<span>one</span>
	} else if x == 2 {
		<span>two</span>
	} else if x == 3 {
		<span>three</span>
	}
}
`,
		},
		"else if with trailing comments": {
			input: `package main

templ A(x int) {
if x == 1 { // first
<span>one</span>
} else if x == 2 { // second
<span>two</span>
}
}
`,
			want: `package main

templ A(x int) {
	if x == 1 {  // first
		<span>one</span>
	} else if x == 2 {  // second
		<span>two</span>
	}
}
`,
		},
		"binding to component call": {
			input: `package main

templ A() {
header := @Header("hi", 1)
<div>{header}</div>
}
`,
			want: `package main

templ A() {
	header := @Header("hi", 1)
	<div>{header}</div>
}
`,
		},
		"binding to component expression": {
			input: `package main

templ (c *card) Render() {
body := @c.body
<div>{body}</div>
}
`,
			want: `package main

templ (c *card) Render() {
	body := @c.body
	<div>{body}</div>
}
`,
		},
		"binding to self-closing and ref elements": {
			input: `package main

templ A() {
rule := <hr class="w-10" />
box := <div ref={c.ref} class="p-1">hi</div>
<div>{rule}{box}</div>
}
`,
			want: `package main

templ A() {
	rule := <hr class="w-10" />
	box := <div ref={c.ref} class="p-1">hi</div>
	<div>{rule}{box}</div>
}
`,
		},
		"binding to element with multi-line children": {
			input: `package main

templ A() {
box := <div>
<span>a</span>
<span>b</span>
</div>
<div>{box}</div>
}
`,
			want: `package main

templ A() {
	box := <div>
		<span>a</span>
		<span>b</span>
	</div>
	<div>{box}</div>
}
`,
		},
		"var binding normalized to short form": {
			input: `package main

templ A() {
var box = <span>hi</span>
<div>{box}</div>
}
`,
			want: `package main

templ A() {
	box := <span>hi</span>
	<div>{box}</div>
}
`,
		},
		"multi-line component call args": {
			input: "package main\n\ntempl A() {\n@Header(\"a,b\",\n'x',\n`raw,arg`,\n[]int{1, 2},\nfmt.Sprintf(\"%d\", 3))\n}\n",
			want:  "package main\n\ntempl A() {\n\t@Header(\n\t\t\"a,b\",\n\t\t'x',\n\t\t`raw,arg`,\n\t\t[]int{1, 2},\n\t\tfmt.Sprintf(\"%d\", 3),\n\t)\n}\n",
		},
		"block comment in call args": {
			input: `package main

templ A() {
@Header(/*title*/ "hi")
}
`,
			want: `package main

templ A() {
	/* title */
	@Header("hi")
}
`,
		},
		"invalid go func preserved verbatim": {
			input: `package main

func bad() {
return +
}

templ A() {
<span>hi</span>
}
`,
			want: `package main

func bad() {
return +
}

templ A() {
	<span>hi</span>
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			fmtr := newTestFormatter()
			got, err := fmtr.Format("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Format() mismatch:\ngot:\n%s\nwant:\n%s", got, tt.want)
			}

			// Formatting must be idempotent.
			again, err := fmtr.Format("test.gsx", got)
			if err != nil {
				t.Fatalf("second Format() error = %v", err)
			}
			if again != got {
				t.Errorf("Format() not idempotent:\nfirst:\n%s\nsecond:\n%s", got, again)
			}
		})
	}
}

// TestSplitTopLevelArgs exercises the argument splitter directly, including
// string, rune, and backtick literals as well as nested bracket depth.
func TestSplitTopLevelArgs(t *testing.T) {
	type tc struct {
		args string
		want []string
	}

	tests := map[string]tc{
		"empty": {
			args: "",
			want: nil,
		},
		"single arg": {
			args: "x",
			want: []string{"x"},
		},
		"simple args without trailing comma": {
			args: "a, b",
			want: []string{"a", "b"},
		},
		"trailing comma yields no empty arg": {
			args: "a, b,",
			want: []string{"a", "b"},
		},
		"comma inside double-quoted string": {
			args: `"a,b", x`,
			want: []string{`"a,b"`, "x"},
		},
		"escaped quote inside string": {
			args: `"a\",b", x`,
			want: []string{`"a\",b"`, "x"},
		},
		"comma inside rune literal": {
			args: `',', x`,
			want: []string{`','`, "x"},
		},
		"escaped rune literal": {
			args: `'\'', x`,
			want: []string{`'\''`, "x"},
		},
		"comma inside backtick string": {
			args: "`a,b`, x",
			want: []string{"`a,b`", "x"},
		},
		"comma inside nested parens": {
			args: `fmt.Sprintf("%d,%d", a, b), y`,
			want: []string{`fmt.Sprintf("%d,%d", a, b)`, "y"},
		},
		"comma inside brackets and braces": {
			args: `[]int{1, 2}, m[k]`,
			want: []string{"[]int{1, 2}", "m[k]"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := splitTopLevelArgs(tt.args)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitTopLevelArgs(%q) = %#v, want %#v", tt.args, got, tt.want)
			}
		})
	}
}

// TestFormatGoCode exercises the gofmt wrapper, including the fallback when
// the code is not valid Go.
func TestFormatGoCode(t *testing.T) {
	type tc struct {
		code string
		want string
	}

	tests := map[string]tc{
		"valid function is formatted": {
			code: "func helper( x int ) int {\nreturn x\n}",
			want: "func helper(x int) int {\n\treturn x\n}",
		},
		"invalid code returned unchanged": {
			code: "func bad() {\nreturn +\n}",
			want: "func bad() {\nreturn +\n}",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := formatGoCode(tt.code)
			if got != tt.want {
				t.Errorf("formatGoCode(%q) = %q, want %q", tt.code, got, tt.want)
			}
		})
	}
}
