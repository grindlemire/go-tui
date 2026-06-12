package tuigen

import (
	"strings"
	"testing"
)

func TestGenerator_TagSpecificOptions(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"hr renders WithHR": {
			input: `package x
templ App() {
	<div>
		<hr/>
	</div>
}`,
			wantContains: []string{"tui.WithHR()"},
		},
		"br renders zero width one height": {
			input: `package x
templ App() {
	<div>
		<br/>
	</div>
}`,
			wantContains: []string{
				"tui.WithWidth(0)",
				"tui.WithHeight(1)",
			},
		},
		"table row and cells get tags and direction": {
			input: `package x
templ App() {
	<table>
		<tr>
			<th>Name</th>
			<td>Bob</td>
		</tr>
	</table>
}`,
			wantContains: []string{
				`tui.WithTag("table")`,
				"tui.WithDisplay(tui.DisplayFlex)",
				"tui.WithDirection(tui.Column)",
				`tui.WithTag("tr")`,
				"tui.WithDirection(tui.Row)",
				`tui.WithTag("th")`,
				`tui.WithText("Name")`,
				`tui.WithTag("td")`,
				`tui.WithText("Bob")`,
			},
		},
		"p with expression child uses WithText": {
			input: `package x
templ App(msg string) {
	<p>{msg}</p>
}`,
			wantContains: []string{"tui.WithText(msg)"},
		},
		"span with numeric expression wraps in Sprint": {
			input: `package x
templ App() {
	<span>{42}</span>
}`,
			wantContains: []string{"tui.WithText(fmt.Sprint(42))"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := parseAndGenerateSkipImports("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}
			code := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, code)
				}
			}
		})
	}
}

func TestGenerator_HandlerAttributes(t *testing.T) {
	input := `package x
templ App(onF func(), onB func(), onA func()) {
	<div focusable={true} onFocus={onF} onBlur={onB} onActivate={onA}></div>
}`

	output, err := parseAndGenerateSkipImports("test.gsx", input)
	if err != nil {
		t.Fatalf("generation failed: %v", err)
	}
	code := string(output)
	for _, want := range []string{
		"tui.WithFocusable(true)",
		"tui.WithOnFocus(onF)",
		"tui.WithOnBlur(onB)",
		"tui.WithOnActivate(onA)",
	} {
		if !strings.Contains(code, want) {
			t.Errorf("output missing %q\nGot:\n%s", want, code)
		}
	}
}

func TestGenerator_AttributeValueForms(t *testing.T) {
	type tc struct {
		input           string
		wantContains    []string
		wantNotContains []string
	}

	tests := map[string]tc{
		"width percent string": {
			input: `package x
templ App() {
	<div width="50%"></div>
}`,
			wantContains: []string{"tui.WithWidthPercent(50)"},
		},
		"height percent string": {
			input: `package x
templ App() {
	<div height="25%"></div>
}`,
			wantContains: []string{"tui.WithHeightPercent(25)"},
		},
		"int literal attribute": {
			input: `package x
templ App() {
	<div gap=2></div>
}`,
			wantContains: []string{"tui.WithGap(2)"},
		},
		"float literal attribute": {
			input: `package x
templ App() {
	<div flexGrow=1.5></div>
}`,
			wantContains: []string{"tui.WithFlexGrow(1.5)"},
		},
		"bool literal attribute": {
			input: `package x
templ App() {
	<div scrollable=true hideScrollbar=false></div>
}`,
			wantContains: []string{
				"tui.WithScrollable(true)",
				"tui.WithScrollbarHidden(false)",
			},
		},
		"string literal attribute": {
			input: `package x
templ App() {
	<div text="hello" borderTitle="Title"></div>
}`,
			wantContains: []string{
				`tui.WithText("hello")`,
				`tui.WithBorderTitle("Title")`,
			},
		},
		"go expr attribute": {
			input: `package x
templ App(w int) {
	<div width={w * 2}></div>
}`,
			wantContains: []string{"tui.WithWidth(w * 2)"},
		},
		"unknown attribute is skipped": {
			input: `package x
templ App() {
	<div bogus="nope"></div>
}`,
			wantNotContains: []string{"bogus", "nope"},
		},
		"class with text style methods builds combined style": {
			input: `package x
templ App() {
	<div class="font-bold text-red p-1"></div>
}`,
			wantContains: []string{
				"tui.WithPadding(1)",
				"Bold()",
				"Foreground(tui.Red)",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := parseAndGenerateSkipImports("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}
			code := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, code)
				}
			}
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(code, notWant) {
					t.Errorf("output contains unexpected %q\nGot:\n%s", notWant, code)
				}
			}
		})
	}
}

func TestGenerator_RefBindingForms(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"single ref uses Set": {
			input: `package x
templ App() {
	header := tui.NewRef()
	<div ref={header}></div>
}`,
			wantContains: []string{"header.Set(__tui_0)"},
		},
		"ref in loop uses Append": {
			input: `package x
templ App(items []string) {
	rows := tui.NewRefList()
	<div>
		for _, item := range items {
			<span ref={rows}>{item}</span>
		}
	</div>
}`,
			wantContains: []string{"rows.Append(__tui_"},
		},
		"keyed ref in loop uses Put": {
			input: `package x
templ App(items []string) {
	rows := tui.NewRefMap[string]()
	<div>
		for _, item := range items {
			<span ref={rows} key={item}>{item}</span>
		}
	</div>
}`,
			wantContains: []string{"rows.Put(item, __tui_"},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := parseAndGenerateSkipImports("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}
			code := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, code)
				}
			}
		})
	}
}

func TestGenerator_ComponentElements(t *testing.T) {
	type tc struct {
		input        string
		wantContains []string
	}

	tests := map[string]tc{
		"input element with attributes and handlers": {
			input: `package x

type form struct{}

templ (c *form) Render() {
	<input placeholder="name" width={20} onSubmit={c.submit} onChange={c.change} autoFocus={true} />
}`,
			wantContains: []string{
				"app.MountPersistent(c, 0, func() tui.Component {",
				"tui.NewInput(",
				`tui.WithInputPlaceholder("name")`,
				"tui.WithInputWidth(20)",
				"tui.WithInputOnSubmit(c.submit)",
				"tui.WithInputOnChange(c.change)",
				"tui.WithInputAutoFocus(true)",
			},
		},
		"input value string literal wrapped in NewState": {
			input: `package x

type form struct{}

templ (c *form) Render() {
	<input value="initial" />
}`,
			wantContains: []string{
				`tui.WithInputValue(tui.NewState("initial"))`,
			},
		},
		"textarea with submit handler and cursor": {
			input: `package x

type form struct{}

templ (c *form) Render() {
	<textarea onSubmit={c.send} cursor={'|'} maxHeight={5} />
}`,
			wantContains: []string{
				"tui.NewTextArea(",
				"tui.WithTextAreaOnSubmit(c.send)",
				"tui.WithTextAreaCursor('|')",
				"tui.WithTextAreaMaxHeight(5)",
			},
		},
		"textarea with ref uses Set": {
			input: `package x

type form struct{}

templ (c *form) Render() {
	<textarea ref={c.ta} />
}`,
			wantContains: []string{
				"c.ta.Set(__tui_0)",
			},
		},
		"textarea inside loop uses runtime mount index and ref append": {
			input: `package x

type form struct{}

templ (c *form) Render() {
	<div>
		for i, f := range c.fields {
			<textarea ref={c.areas} placeholder={f} />
		}
	</div>
}`,
			wantContains: []string{
				"app.MountPersistent(c, (1)*1000000+i, func() tui.Component {",
				"c.areas.Append(__tui_",
				"tui.WithTextAreaPlaceholder(f)",
			},
		},
		"modal with state and options": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	<modal open={c.showModal} backdrop="blank" closeOnEscape={false} trapFocus={true}>
		<span>content</span>
	</modal>
}`,
			wantContains: []string{
				"tui.NewModal(",
				"tui.WithModalOpen(c.showModal)",
				`tui.WithModalBackdrop("blank")`,
				"tui.WithModalCloseOnEscape(false)",
				"tui.WithModalTrapFocus(true)",
				`tui.WithText("content")`,
				".AddChild(",
			},
		},
		"modal open bool literal wrapped in NewState": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	<modal open=true></modal>
}`,
			wantContains: []string{
				"tui.WithModalOpen(tui.NewState(true))",
			},
		},
		"modal class becomes WithModalElementOptions": {
			input: `package x

type shell struct{}

templ (c *shell) Render() {
	<modal open={c.show} class="border-rounded p-2 font-bold"></modal>
}`,
			wantContains: []string{
				"tui.WithModalElementOptions(",
				"tui.WithBorder(tui.BorderRounded)",
				"tui.WithPadding(2)",
				"Bold()",
			},
		},
		"markdown with source width and theme": {
			input: `package x

type docs struct{}

templ (c *docs) Render() {
	<markdown source={c.body} width={80} theme={c.theme} />
}`,
			wantContains: []string{
				"tui.NewMarkdown(",
				"tui.WithMarkdownSource(c.body)",
				"tui.WithMarkdownWidth(80)",
				"tui.WithMarkdownTheme(c.theme)",
			},
		},
		"markdown with reactive state": {
			input: `package x

type docs struct{}

templ (c *docs) Render() {
	<markdown state={c.content} />
}`,
			wantContains: []string{
				"tui.WithMarkdownState(c.content)",
			},
		},
		"component element with no options calls bare constructor": {
			input: `package x

type form struct{}

templ (c *form) Render() {
	<textarea />
}`,
			wantContains: []string{
				"return tui.NewTextArea()",
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			output, err := parseAndGenerateSkipImports("test.gsx", tt.input)
			if err != nil {
				t.Fatalf("generation failed: %v", err)
			}
			code := string(output)
			for _, want := range tt.wantContains {
				if !strings.Contains(code, want) {
					t.Errorf("output missing %q\nGot:\n%s", want, code)
				}
			}
		})
	}
}
