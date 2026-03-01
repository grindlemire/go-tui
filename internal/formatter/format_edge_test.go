package formatter

import (
	"testing"
)

func TestFormat_EdgeCases(t *testing.T) {
	type tc struct {
		input   string
		wantErr bool
	}

	tests := map[string]tc{
		"empty file": {
			input:   "",
			wantErr: true,
		},
		"package only": {
			input: "package test\n",
		},
		"deeply nested elements": {
			input: `package test

templ Deep() {
	<div>
		<div>
			<div>
				<div>
					<span>Deep</span>
				</div>
			</div>
		</div>
	</div>
}
`,
		},
		"multiple components": {
			input: `package test

templ A() {
	<span>A</span>
}

templ B() {
	<span>B</span>
}
`,
		},
		"component with many attributes": {
			input: `package test

templ Styled() {
	<div class="flex-col gap-2 p-2 border-rounded text-cyan bg-black">
		<span>Styled</span>
	</div>
}
`,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			f := New()
			_, err := f.Format("test.gsx", tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Format() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
