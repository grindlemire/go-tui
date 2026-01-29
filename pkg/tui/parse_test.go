package tui

import "testing"

func TestParseInput_PrintableCharacters(t *testing.T) {
	type tc struct {
		input    []byte
		expected []KeyEvent
	}

	tests := map[string]tc{
		"single letter a":       {input: []byte("a"), expected: []KeyEvent{{Key: KeyRune, Rune: 'a'}}},
		"single letter z":       {input: []byte("z"), expected: []KeyEvent{{Key: KeyRune, Rune: 'z'}}},
		"uppercase A":           {input: []byte("A"), expected: []KeyEvent{{Key: KeyRune, Rune: 'A'}}},
		"digit 0":               {input: []byte("0"), expected: []KeyEvent{{Key: KeyRune, Rune: '0'}}},
		"digit 9":               {input: []byte("9"), expected: []KeyEvent{{Key: KeyRune, Rune: '9'}}},
		"space":                 {input: []byte(" "), expected: []KeyEvent{{Key: KeyRune, Rune: ' '}}},
		"special char !":        {input: []byte("!"), expected: []KeyEvent{{Key: KeyRune, Rune: '!'}}},
		"special char @":        {input: []byte("@"), expected: []KeyEvent{{Key: KeyRune, Rune: '@'}}},
		"multiple chars":        {input: []byte("abc"), expected: []KeyEvent{{Key: KeyRune, Rune: 'a'}, {Key: KeyRune, Rune: 'b'}, {Key: KeyRune, Rune: 'c'}}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events := parseInput(tt.input)
			if len(events) != len(tt.expected) {
				t.Fatalf("parseInput(%q) returned %d events, want %d", tt.input, len(events), len(tt.expected))
			}
			for i, e := range events {
				ke, ok := e.(KeyEvent)
				if !ok {
					t.Fatalf("event %d is not KeyEvent", i)
				}
				if ke.Key != tt.expected[i].Key || ke.Rune != tt.expected[i].Rune {
					t.Errorf("event %d: got {Key: %v, Rune: %q}, want {Key: %v, Rune: %q}",
						i, ke.Key, ke.Rune, tt.expected[i].Key, tt.expected[i].Rune)
				}
			}
		})
	}
}

func TestParseInput_UTF8Characters(t *testing.T) {
	type tc struct {
		input    []byte
		expected []KeyEvent
	}

	tests := map[string]tc{
		"japanese char":     {input: []byte("日"), expected: []KeyEvent{{Key: KeyRune, Rune: '日'}}},
		"emoji":             {input: []byte("😀"), expected: []KeyEvent{{Key: KeyRune, Rune: '😀'}}},
		"german umlaut":     {input: []byte("ü"), expected: []KeyEvent{{Key: KeyRune, Rune: 'ü'}}},
		"mixed ascii utf8":  {input: []byte("a日b"), expected: []KeyEvent{{Key: KeyRune, Rune: 'a'}, {Key: KeyRune, Rune: '日'}, {Key: KeyRune, Rune: 'b'}}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events := parseInput(tt.input)
			if len(events) != len(tt.expected) {
				t.Fatalf("parseInput(%q) returned %d events, want %d", tt.input, len(events), len(tt.expected))
			}
			for i, e := range events {
				ke, ok := e.(KeyEvent)
				if !ok {
					t.Fatalf("event %d is not KeyEvent", i)
				}
				if ke.Key != tt.expected[i].Key || ke.Rune != tt.expected[i].Rune {
					t.Errorf("event %d: got {Key: %v, Rune: %q}, want {Key: %v, Rune: %q}",
						i, ke.Key, ke.Rune, tt.expected[i].Key, tt.expected[i].Rune)
				}
			}
		})
	}
}

func TestParseInput_ControlCharacters(t *testing.T) {
	type tc struct {
		input    []byte
		expected Key
	}

	tests := map[string]tc{
		"ctrl+space":   {input: []byte{0x00}, expected: KeyCtrlSpace},
		"ctrl+a":       {input: []byte{0x01}, expected: KeyCtrlA},
		"ctrl+b":       {input: []byte{0x02}, expected: KeyCtrlB},
		"ctrl+c":       {input: []byte{0x03}, expected: KeyCtrlC},
		"ctrl+d":       {input: []byte{0x04}, expected: KeyCtrlD},
		"ctrl+e":       {input: []byte{0x05}, expected: KeyCtrlE},
		"ctrl+f":       {input: []byte{0x06}, expected: KeyCtrlF},
		"ctrl+g":       {input: []byte{0x07}, expected: KeyCtrlG},
		"backspace":    {input: []byte{0x08}, expected: KeyBackspace},
		"tab":          {input: []byte{0x09}, expected: KeyTab},
		"ctrl+j":       {input: []byte{0x0a}, expected: KeyCtrlJ},
		"ctrl+k":       {input: []byte{0x0b}, expected: KeyCtrlK},
		"ctrl+l":       {input: []byte{0x0c}, expected: KeyCtrlL},
		"enter":        {input: []byte{0x0d}, expected: KeyEnter},
		"ctrl+n":       {input: []byte{0x0e}, expected: KeyCtrlN},
		"ctrl+o":       {input: []byte{0x0f}, expected: KeyCtrlO},
		"ctrl+p":       {input: []byte{0x10}, expected: KeyCtrlP},
		"ctrl+q":       {input: []byte{0x11}, expected: KeyCtrlQ},
		"ctrl+r":       {input: []byte{0x12}, expected: KeyCtrlR},
		"ctrl+s":       {input: []byte{0x13}, expected: KeyCtrlS},
		"ctrl+t":       {input: []byte{0x14}, expected: KeyCtrlT},
		"ctrl+u":       {input: []byte{0x15}, expected: KeyCtrlU},
		"ctrl+v":       {input: []byte{0x16}, expected: KeyCtrlV},
		"ctrl+w":       {input: []byte{0x17}, expected: KeyCtrlW},
		"ctrl+x":       {input: []byte{0x18}, expected: KeyCtrlX},
		"ctrl+y":       {input: []byte{0x19}, expected: KeyCtrlY},
		"ctrl+z":       {input: []byte{0x1a}, expected: KeyCtrlZ},
		"escape":       {input: []byte{0x1b}, expected: KeyEscape},
		"del":          {input: []byte{0x7f}, expected: KeyBackspace},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events := parseInput(tt.input)
			if len(events) != 1 {
				t.Fatalf("parseInput(%x) returned %d events, want 1", tt.input, len(events))
			}
			ke, ok := events[0].(KeyEvent)
			if !ok {
				t.Fatalf("event is not KeyEvent")
			}
			if ke.Key != tt.expected {
				t.Errorf("parseInput(%x): got Key = %v, want %v", tt.input, ke.Key, tt.expected)
			}
		})
	}
}

func TestParseInput_ArrowKeys(t *testing.T) {
	type tc struct {
		input    []byte
		expected KeyEvent
	}

	tests := map[string]tc{
		"up":    {input: []byte("\x1b[A"), expected: KeyEvent{Key: KeyUp}},
		"down":  {input: []byte("\x1b[B"), expected: KeyEvent{Key: KeyDown}},
		"right": {input: []byte("\x1b[C"), expected: KeyEvent{Key: KeyRight}},
		"left":  {input: []byte("\x1b[D"), expected: KeyEvent{Key: KeyLeft}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events := parseInput(tt.input)
			if len(events) != 1 {
				t.Fatalf("parseInput(%q) returned %d events, want 1", tt.input, len(events))
			}
			ke, ok := events[0].(KeyEvent)
			if !ok {
				t.Fatalf("event is not KeyEvent")
			}
			if ke.Key != tt.expected.Key || ke.Mod != tt.expected.Mod {
				t.Errorf("parseInput(%q): got {Key: %v, Mod: %v}, want {Key: %v, Mod: %v}",
					tt.input, ke.Key, ke.Mod, tt.expected.Key, tt.expected.Mod)
			}
		})
	}
}

func TestParseInput_ArrowKeysWithModifiers(t *testing.T) {
	type tc struct {
		input    []byte
		expected KeyEvent
	}

	tests := map[string]tc{
		"up+shift":        {input: []byte("\x1b[1;2A"), expected: KeyEvent{Key: KeyUp, Mod: ModShift}},
		"up+alt":          {input: []byte("\x1b[1;3A"), expected: KeyEvent{Key: KeyUp, Mod: ModAlt}},
		"up+shift+alt":    {input: []byte("\x1b[1;4A"), expected: KeyEvent{Key: KeyUp, Mod: ModShift | ModAlt}},
		"up+ctrl":         {input: []byte("\x1b[1;5A"), expected: KeyEvent{Key: KeyUp, Mod: ModCtrl}},
		"up+ctrl+shift":   {input: []byte("\x1b[1;6A"), expected: KeyEvent{Key: KeyUp, Mod: ModCtrl | ModShift}},
		"up+ctrl+alt":     {input: []byte("\x1b[1;7A"), expected: KeyEvent{Key: KeyUp, Mod: ModCtrl | ModAlt}},
		"up+all":          {input: []byte("\x1b[1;8A"), expected: KeyEvent{Key: KeyUp, Mod: ModCtrl | ModAlt | ModShift}},
		"down+shift":      {input: []byte("\x1b[1;2B"), expected: KeyEvent{Key: KeyDown, Mod: ModShift}},
		"right+ctrl":      {input: []byte("\x1b[1;5C"), expected: KeyEvent{Key: KeyRight, Mod: ModCtrl}},
		"left+alt":        {input: []byte("\x1b[1;3D"), expected: KeyEvent{Key: KeyLeft, Mod: ModAlt}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events := parseInput(tt.input)
			if len(events) != 1 {
				t.Fatalf("parseInput(%q) returned %d events, want 1", tt.input, len(events))
			}
			ke, ok := events[0].(KeyEvent)
			if !ok {
				t.Fatalf("event is not KeyEvent")
			}
			if ke.Key != tt.expected.Key || ke.Mod != tt.expected.Mod {
				t.Errorf("parseInput(%q): got {Key: %v, Mod: %v}, want {Key: %v, Mod: %v}",
					tt.input, ke.Key, ke.Mod, tt.expected.Key, tt.expected.Mod)
			}
		})
	}
}

func TestParseInput_FunctionKeys_SS3(t *testing.T) {
	type tc struct {
		input    []byte
		expected Key
	}

	tests := map[string]tc{
		"f1": {input: []byte("\x1bOP"), expected: KeyF1},
		"f2": {input: []byte("\x1bOQ"), expected: KeyF2},
		"f3": {input: []byte("\x1bOR"), expected: KeyF3},
		"f4": {input: []byte("\x1bOS"), expected: KeyF4},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events := parseInput(tt.input)
			if len(events) != 1 {
				t.Fatalf("parseInput(%q) returned %d events, want 1", tt.input, len(events))
			}
			ke, ok := events[0].(KeyEvent)
			if !ok {
				t.Fatalf("event is not KeyEvent")
			}
			if ke.Key != tt.expected {
				t.Errorf("parseInput(%q): got Key = %v, want %v", tt.input, ke.Key, tt.expected)
			}
		})
	}
}

func TestParseInput_FunctionKeys_CSI(t *testing.T) {
	type tc struct {
		input    []byte
		expected Key
	}

	tests := map[string]tc{
		"f1 csi":  {input: []byte("\x1b[11~"), expected: KeyF1},
		"f2 csi":  {input: []byte("\x1b[12~"), expected: KeyF2},
		"f3 csi":  {input: []byte("\x1b[13~"), expected: KeyF3},
		"f4 csi":  {input: []byte("\x1b[14~"), expected: KeyF4},
		"f5":      {input: []byte("\x1b[15~"), expected: KeyF5},
		"f6":      {input: []byte("\x1b[17~"), expected: KeyF6},
		"f7":      {input: []byte("\x1b[18~"), expected: KeyF7},
		"f8":      {input: []byte("\x1b[19~"), expected: KeyF8},
		"f9":      {input: []byte("\x1b[20~"), expected: KeyF9},
		"f10":     {input: []byte("\x1b[21~"), expected: KeyF10},
		"f11":     {input: []byte("\x1b[23~"), expected: KeyF11},
		"f12":     {input: []byte("\x1b[24~"), expected: KeyF12},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events := parseInput(tt.input)
			if len(events) != 1 {
				t.Fatalf("parseInput(%q) returned %d events, want 1", tt.input, len(events))
			}
			ke, ok := events[0].(KeyEvent)
			if !ok {
				t.Fatalf("event is not KeyEvent")
			}
			if ke.Key != tt.expected {
				t.Errorf("parseInput(%q): got Key = %v, want %v", tt.input, ke.Key, tt.expected)
			}
		})
	}
}

func TestParseInput_NavigationKeys(t *testing.T) {
	type tc struct {
		input    []byte
		expected KeyEvent
	}

	tests := map[string]tc{
		"home csi H":  {input: []byte("\x1b[H"), expected: KeyEvent{Key: KeyHome}},
		"end csi F":   {input: []byte("\x1b[F"), expected: KeyEvent{Key: KeyEnd}},
		"home csi 1~": {input: []byte("\x1b[1~"), expected: KeyEvent{Key: KeyHome}},
		"insert":      {input: []byte("\x1b[2~"), expected: KeyEvent{Key: KeyInsert}},
		"delete":      {input: []byte("\x1b[3~"), expected: KeyEvent{Key: KeyDelete}},
		"end csi 4~":  {input: []byte("\x1b[4~"), expected: KeyEvent{Key: KeyEnd}},
		"pageup":      {input: []byte("\x1b[5~"), expected: KeyEvent{Key: KeyPageUp}},
		"pagedown":    {input: []byte("\x1b[6~"), expected: KeyEvent{Key: KeyPageDown}},
		"backtab":     {input: []byte("\x1b[Z"), expected: KeyEvent{Key: KeyTab, Mod: ModShift}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events := parseInput(tt.input)
			if len(events) != 1 {
				t.Fatalf("parseInput(%q) returned %d events, want 1", tt.input, len(events))
			}
			ke, ok := events[0].(KeyEvent)
			if !ok {
				t.Fatalf("event is not KeyEvent")
			}
			if ke.Key != tt.expected.Key || ke.Mod != tt.expected.Mod {
				t.Errorf("parseInput(%q): got {Key: %v, Mod: %v}, want {Key: %v, Mod: %v}",
					tt.input, ke.Key, ke.Mod, tt.expected.Key, tt.expected.Mod)
			}
		})
	}
}

func TestParseInput_AltCombinations(t *testing.T) {
	type tc struct {
		input    []byte
		expected KeyEvent
	}

	tests := map[string]tc{
		"alt+a": {input: []byte("\x1ba"), expected: KeyEvent{Key: KeyRune, Rune: 'a', Mod: ModAlt}},
		"alt+z": {input: []byte("\x1bz"), expected: KeyEvent{Key: KeyRune, Rune: 'z', Mod: ModAlt}},
		"alt+A": {input: []byte("\x1bA"), expected: KeyEvent{Key: KeyRune, Rune: 'A', Mod: ModAlt}},
		"alt+1": {input: []byte("\x1b1"), expected: KeyEvent{Key: KeyRune, Rune: '1', Mod: ModAlt}},
		"alt+ ": {input: []byte("\x1b "), expected: KeyEvent{Key: KeyRune, Rune: ' ', Mod: ModAlt}},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events := parseInput(tt.input)
			if len(events) != 1 {
				t.Fatalf("parseInput(%q) returned %d events, want 1", tt.input, len(events))
			}
			ke, ok := events[0].(KeyEvent)
			if !ok {
				t.Fatalf("event is not KeyEvent")
			}
			if ke.Key != tt.expected.Key || ke.Rune != tt.expected.Rune || ke.Mod != tt.expected.Mod {
				t.Errorf("parseInput(%q): got {Key: %v, Rune: %q, Mod: %v}, want {Key: %v, Rune: %q, Mod: %v}",
					tt.input, ke.Key, ke.Rune, ke.Mod, tt.expected.Key, tt.expected.Rune, tt.expected.Mod)
			}
		})
	}
}

func TestParseInput_MultipleEvents(t *testing.T) {
	type tc struct {
		input    []byte
		expected []KeyEvent
	}

	tests := map[string]tc{
		"two letters": {
			input: []byte("ab"),
			expected: []KeyEvent{
				{Key: KeyRune, Rune: 'a'},
				{Key: KeyRune, Rune: 'b'},
			},
		},
		"letter then arrow": {
			input: []byte("a\x1b[A"),
			expected: []KeyEvent{
				{Key: KeyRune, Rune: 'a'},
				{Key: KeyUp},
			},
		},
		"arrow then letter": {
			input: []byte("\x1b[Aa"),
			expected: []KeyEvent{
				{Key: KeyUp},
				{Key: KeyRune, Rune: 'a'},
			},
		},
		"two arrows": {
			input: []byte("\x1b[A\x1b[B"),
			expected: []KeyEvent{
				{Key: KeyUp},
				{Key: KeyDown},
			},
		},
		"ctrl then letter": {
			input: []byte{0x03, 'x'},
			expected: []KeyEvent{
				{Key: KeyCtrlC},
				{Key: KeyRune, Rune: 'x'},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events := parseInput(tt.input)
			if len(events) != len(tt.expected) {
				t.Fatalf("parseInput(%q) returned %d events, want %d", tt.input, len(events), len(tt.expected))
			}
			for i, e := range events {
				ke, ok := e.(KeyEvent)
				if !ok {
					t.Fatalf("event %d is not KeyEvent", i)
				}
				if ke.Key != tt.expected[i].Key || ke.Rune != tt.expected[i].Rune || ke.Mod != tt.expected[i].Mod {
					t.Errorf("event %d: got {Key: %v, Rune: %q, Mod: %v}, want {Key: %v, Rune: %q, Mod: %v}",
						i, ke.Key, ke.Rune, ke.Mod, tt.expected[i].Key, tt.expected[i].Rune, tt.expected[i].Mod)
				}
			}
		})
	}
}

func TestParseInput_LoneEscape(t *testing.T) {
	events := parseInput([]byte{0x1b})
	if len(events) != 1 {
		t.Fatalf("parseInput lone escape returned %d events, want 1", len(events))
	}
	ke, ok := events[0].(KeyEvent)
	if !ok {
		t.Fatal("event is not KeyEvent")
	}
	if ke.Key != KeyEscape {
		t.Errorf("got Key = %v, want KeyEscape", ke.Key)
	}
}

func TestControlToKey(t *testing.T) {
	type tc struct {
		input    byte
		expected Key
	}

	tests := map[string]tc{
		"0x00": {input: 0x00, expected: KeyCtrlSpace},
		"0x01": {input: 0x01, expected: KeyCtrlA},
		"0x1a": {input: 0x1a, expected: KeyCtrlZ},
		"0x1b": {input: 0x1b, expected: KeyEscape},
		"0x1c": {input: 0x1c, expected: KeyNone},
		"0x1f": {input: 0x1f, expected: KeyNone},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := controlToKey(tt.input)
			if got != tt.expected {
				t.Errorf("controlToKey(0x%02x) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDecodeModifier(t *testing.T) {
	type tc struct {
		param    int
		expected Modifier
	}

	tests := map[string]tc{
		"0 none":        {param: 0, expected: ModNone},
		"1 none":        {param: 1, expected: ModNone},
		"2 shift":       {param: 2, expected: ModShift},
		"3 alt":         {param: 3, expected: ModAlt},
		"4 shift+alt":   {param: 4, expected: ModShift | ModAlt},
		"5 ctrl":        {param: 5, expected: ModCtrl},
		"6 ctrl+shift":  {param: 6, expected: ModCtrl | ModShift},
		"7 ctrl+alt":    {param: 7, expected: ModCtrl | ModAlt},
		"8 all":         {param: 8, expected: ModCtrl | ModAlt | ModShift},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := decodeModifier(tt.param)
			if got != tt.expected {
				t.Errorf("decodeModifier(%d) = %v, want %v", tt.param, got, tt.expected)
			}
		})
	}
}

func TestParseSS3(t *testing.T) {
	type tc struct {
		input    byte
		expected Key
	}

	tests := map[string]tc{
		"P f1":    {input: 'P', expected: KeyF1},
		"Q f2":    {input: 'Q', expected: KeyF2},
		"R f3":    {input: 'R', expected: KeyF3},
		"S f4":    {input: 'S', expected: KeyF4},
		"A up":    {input: 'A', expected: KeyUp},
		"B down":  {input: 'B', expected: KeyDown},
		"C right": {input: 'C', expected: KeyRight},
		"D left":  {input: 'D', expected: KeyLeft},
		"H home":  {input: 'H', expected: KeyHome},
		"F end":   {input: 'F', expected: KeyEnd},
		"unknown": {input: 'X', expected: KeyNone},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := parseSS3(tt.input)
			if got != tt.expected {
				t.Errorf("parseSS3('%c') = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseMouseSGR(t *testing.T) {
	type tc struct {
		input            []byte
		expectedEvent    MouseEvent
		expectedConsumed int
	}

	tests := map[string]tc{
		"left press at 1,1": {
			input:            []byte("\x1b[<0;1;1M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MousePress, X: 0, Y: 0},
			expectedConsumed: 9,
		},
		"left release at 1,1": {
			input:            []byte("\x1b[<0;1;1m"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MouseRelease, X: 0, Y: 0},
			expectedConsumed: 9,
		},
		"middle press at 10,20": {
			input:            []byte("\x1b[<1;10;20M"),
			expectedEvent:    MouseEvent{Button: MouseMiddle, Action: MousePress, X: 9, Y: 19},
			expectedConsumed: 11,
		},
		"right press at 5,5": {
			input:            []byte("\x1b[<2;5;5M"),
			expectedEvent:    MouseEvent{Button: MouseRight, Action: MousePress, X: 4, Y: 4},
			expectedConsumed: 9,
		},
		"wheel up": {
			input:            []byte("\x1b[<64;10;10M"),
			expectedEvent:    MouseEvent{Button: MouseWheelUp, Action: MousePress, X: 9, Y: 9},
			expectedConsumed: 12,
		},
		"wheel down": {
			input:            []byte("\x1b[<65;10;10M"),
			expectedEvent:    MouseEvent{Button: MouseWheelDown, Action: MousePress, X: 9, Y: 9},
			expectedConsumed: 12,
		},
		"left drag": {
			input:            []byte("\x1b[<32;15;25M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MouseDrag, X: 14, Y: 24},
			expectedConsumed: 12,
		},
		"shift+left click": {
			input:            []byte("\x1b[<4;5;5M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MousePress, X: 4, Y: 4, Mod: ModShift},
			expectedConsumed: 9,
		},
		"alt+left click": {
			input:            []byte("\x1b[<8;5;5M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MousePress, X: 4, Y: 4, Mod: ModAlt},
			expectedConsumed: 9,
		},
		"ctrl+left click": {
			input:            []byte("\x1b[<16;5;5M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MousePress, X: 4, Y: 4, Mod: ModCtrl},
			expectedConsumed: 10,
		},
		"ctrl+shift+alt+left click": {
			input:            []byte("\x1b[<28;5;5M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MousePress, X: 4, Y: 4, Mod: ModCtrl | ModShift | ModAlt},
			expectedConsumed: 10,
		},
		"large coordinates": {
			input:            []byte("\x1b[<0;200;100M"),
			expectedEvent:    MouseEvent{Button: MouseLeft, Action: MousePress, X: 199, Y: 99},
			expectedConsumed: 13,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			event, consumed := parseMouseSGR(tt.input)
			if consumed != tt.expectedConsumed {
				t.Errorf("parseMouseSGR(%q) consumed %d bytes, want %d", tt.input, consumed, tt.expectedConsumed)
			}
			if event.Button != tt.expectedEvent.Button {
				t.Errorf("parseMouseSGR(%q) button = %v, want %v", tt.input, event.Button, tt.expectedEvent.Button)
			}
			if event.Action != tt.expectedEvent.Action {
				t.Errorf("parseMouseSGR(%q) action = %v, want %v", tt.input, event.Action, tt.expectedEvent.Action)
			}
			if event.X != tt.expectedEvent.X {
				t.Errorf("parseMouseSGR(%q) X = %d, want %d", tt.input, event.X, tt.expectedEvent.X)
			}
			if event.Y != tt.expectedEvent.Y {
				t.Errorf("parseMouseSGR(%q) Y = %d, want %d", tt.input, event.Y, tt.expectedEvent.Y)
			}
			if event.Mod != tt.expectedEvent.Mod {
				t.Errorf("parseMouseSGR(%q) Mod = %v, want %v", tt.input, event.Mod, tt.expectedEvent.Mod)
			}
		})
	}
}

func TestParseMouseSGR_Invalid(t *testing.T) {
	type tc struct {
		input []byte
	}

	tests := map[string]tc{
		"empty":                     {input: []byte{}},
		"too short":                 {input: []byte("\x1b[<")},
		"missing M":                 {input: []byte("\x1b[<0;1;1")},
		"wrong prefix":              {input: []byte("\x1b[0;1;1M")},
		"missing x":                 {input: []byte("\x1b[<0;;1M")},
		"missing y":                 {input: []byte("\x1b[<0;1;M")},
		"non-numeric button":        {input: []byte("\x1b[<a;1;1M")},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, consumed := parseMouseSGR(tt.input)
			if consumed != 0 {
				t.Errorf("parseMouseSGR(%q) consumed %d bytes, want 0 for invalid input", tt.input, consumed)
			}
		})
	}
}

func TestParseInput_MouseEvents(t *testing.T) {
	type tc struct {
		input    []byte
		expected MouseEvent
	}

	tests := map[string]tc{
		"left click": {
			input:    []byte("\x1b[<0;10;20M"),
			expected: MouseEvent{Button: MouseLeft, Action: MousePress, X: 9, Y: 19},
		},
		"right click": {
			input:    []byte("\x1b[<2;5;5M"),
			expected: MouseEvent{Button: MouseRight, Action: MousePress, X: 4, Y: 4},
		},
		"wheel up": {
			input:    []byte("\x1b[<64;1;1M"),
			expected: MouseEvent{Button: MouseWheelUp, Action: MousePress, X: 0, Y: 0},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			events := parseInput(tt.input)
			if len(events) != 1 {
				t.Fatalf("parseInput(%q) returned %d events, want 1", tt.input, len(events))
			}
			me, ok := events[0].(MouseEvent)
			if !ok {
				t.Fatalf("event is not MouseEvent, got %T", events[0])
			}
			if me.Button != tt.expected.Button {
				t.Errorf("parseInput(%q) button = %v, want %v", tt.input, me.Button, tt.expected.Button)
			}
			if me.Action != tt.expected.Action {
				t.Errorf("parseInput(%q) action = %v, want %v", tt.input, me.Action, tt.expected.Action)
			}
			if me.X != tt.expected.X || me.Y != tt.expected.Y {
				t.Errorf("parseInput(%q) pos = (%d,%d), want (%d,%d)", tt.input, me.X, me.Y, tt.expected.X, tt.expected.Y)
			}
		})
	}
}
