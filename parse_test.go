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
		expected KeyEvent
	}

	tests := map[string]tc{
		"ctrl+space":   {input: []byte{0x00}, expected: KeyEvent{Key: KeyRune, Rune: ' ', Mod: ModCtrl}},
		"ctrl+a":       {input: []byte{0x01}, expected: KeyEvent{Key: KeyRune, Rune: 'a', Mod: ModCtrl}},
		"ctrl+b":       {input: []byte{0x02}, expected: KeyEvent{Key: KeyRune, Rune: 'b', Mod: ModCtrl}},
		"ctrl+c":       {input: []byte{0x03}, expected: KeyEvent{Key: KeyRune, Rune: 'c', Mod: ModCtrl}},
		"ctrl+d":       {input: []byte{0x04}, expected: KeyEvent{Key: KeyRune, Rune: 'd', Mod: ModCtrl}},
		"ctrl+e":       {input: []byte{0x05}, expected: KeyEvent{Key: KeyRune, Rune: 'e', Mod: ModCtrl}},
		"ctrl+f":       {input: []byte{0x06}, expected: KeyEvent{Key: KeyRune, Rune: 'f', Mod: ModCtrl}},
		"ctrl+g":       {input: []byte{0x07}, expected: KeyEvent{Key: KeyRune, Rune: 'g', Mod: ModCtrl}},
		"backspace":    {input: []byte{0x08}, expected: KeyEvent{Key: KeyBackspace}},
		"tab":          {input: []byte{0x09}, expected: KeyEvent{Key: KeyTab}},
		"ctrl+j":       {input: []byte{0x0a}, expected: KeyEvent{Key: KeyRune, Rune: 'j', Mod: ModCtrl}},
		"ctrl+k":       {input: []byte{0x0b}, expected: KeyEvent{Key: KeyRune, Rune: 'k', Mod: ModCtrl}},
		"ctrl+l":       {input: []byte{0x0c}, expected: KeyEvent{Key: KeyRune, Rune: 'l', Mod: ModCtrl}},
		"enter":        {input: []byte{0x0d}, expected: KeyEvent{Key: KeyEnter}},
		"ctrl+n":       {input: []byte{0x0e}, expected: KeyEvent{Key: KeyRune, Rune: 'n', Mod: ModCtrl}},
		"ctrl+o":       {input: []byte{0x0f}, expected: KeyEvent{Key: KeyRune, Rune: 'o', Mod: ModCtrl}},
		"ctrl+p":       {input: []byte{0x10}, expected: KeyEvent{Key: KeyRune, Rune: 'p', Mod: ModCtrl}},
		"ctrl+q":       {input: []byte{0x11}, expected: KeyEvent{Key: KeyRune, Rune: 'q', Mod: ModCtrl}},
		"ctrl+r":       {input: []byte{0x12}, expected: KeyEvent{Key: KeyRune, Rune: 'r', Mod: ModCtrl}},
		"ctrl+s":       {input: []byte{0x13}, expected: KeyEvent{Key: KeyRune, Rune: 's', Mod: ModCtrl}},
		"ctrl+t":       {input: []byte{0x14}, expected: KeyEvent{Key: KeyRune, Rune: 't', Mod: ModCtrl}},
		"ctrl+u":       {input: []byte{0x15}, expected: KeyEvent{Key: KeyRune, Rune: 'u', Mod: ModCtrl}},
		"ctrl+v":       {input: []byte{0x16}, expected: KeyEvent{Key: KeyRune, Rune: 'v', Mod: ModCtrl}},
		"ctrl+w":       {input: []byte{0x17}, expected: KeyEvent{Key: KeyRune, Rune: 'w', Mod: ModCtrl}},
		"ctrl+x":       {input: []byte{0x18}, expected: KeyEvent{Key: KeyRune, Rune: 'x', Mod: ModCtrl}},
		"ctrl+y":       {input: []byte{0x19}, expected: KeyEvent{Key: KeyRune, Rune: 'y', Mod: ModCtrl}},
		"ctrl+z":       {input: []byte{0x1a}, expected: KeyEvent{Key: KeyRune, Rune: 'z', Mod: ModCtrl}},
		"escape":       {input: []byte{0x1b}, expected: KeyEvent{Key: KeyEscape}},
		"del":          {input: []byte{0x7f}, expected: KeyEvent{Key: KeyBackspace}},
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
			if ke.Key != tt.expected.Key || ke.Rune != tt.expected.Rune || ke.Mod != tt.expected.Mod {
				t.Errorf("parseInput(%x): got {Key: %v, Rune: %q, Mod: %v}, want {Key: %v, Rune: %q, Mod: %v}",
					tt.input, ke.Key, ke.Rune, ke.Mod, tt.expected.Key, tt.expected.Rune, tt.expected.Mod)
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
				{Key: KeyRune, Rune: 'c', Mod: ModCtrl},
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
		wantKey  Key
		wantRune rune
		wantMod  Modifier
	}

	tests := map[string]tc{
		"0x00 ctrl+space": {input: 0x00, wantKey: KeyRune, wantRune: ' ', wantMod: ModCtrl},
		"0x01 ctrl+a":     {input: 0x01, wantKey: KeyRune, wantRune: 'a', wantMod: ModCtrl},
		"0x1a ctrl+z":     {input: 0x1a, wantKey: KeyRune, wantRune: 'z', wantMod: ModCtrl},
		"0x08 backspace":  {input: 0x08, wantKey: KeyBackspace, wantRune: 0, wantMod: ModNone},
		"0x09 tab":        {input: 0x09, wantKey: KeyTab, wantRune: 0, wantMod: ModNone},
		"0x0d enter":      {input: 0x0d, wantKey: KeyEnter, wantRune: 0, wantMod: ModNone},
		"0x1b escape":     {input: 0x1b, wantKey: KeyEscape, wantRune: 0, wantMod: ModNone},
		"0x1c none":       {input: 0x1c, wantKey: KeyNone, wantRune: 0, wantMod: ModNone},
		"0x1f none":       {input: 0x1f, wantKey: KeyNone, wantRune: 0, wantMod: ModNone},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotKey, gotRune, gotMod := controlToKey(tt.input)
			if gotKey != tt.wantKey || gotRune != tt.wantRune || gotMod != tt.wantMod {
				t.Errorf("controlToKey(0x%02x) = (%v, %q, %v), want (%v, %q, %v)",
					tt.input, gotKey, gotRune, gotMod, tt.wantKey, tt.wantRune, tt.wantMod)
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

func TestParseInput_KittyCSIu(t *testing.T) {
	type tc struct {
		input    []byte
		expected KeyEvent
	}

	tests := map[string]tc{
		"backspace": {
			input:    []byte("\x1b[127;1u"),
			expected: KeyEvent{Key: KeyBackspace},
		},
		"enter": {
			input:    []byte("\x1b[13;1u"),
			expected: KeyEvent{Key: KeyEnter},
		},
		"tab": {
			input:    []byte("\x1b[9;1u"),
			expected: KeyEvent{Key: KeyTab},
		},
		"escape": {
			input:    []byte("\x1b[27;1u"),
			expected: KeyEvent{Key: KeyEscape},
		},
		"ctrl+h (distinct from backspace)": {
			input:    []byte("\x1b[104;5u"),
			expected: KeyEvent{Key: KeyRune, Rune: 'h', Mod: ModCtrl},
		},
		"ctrl+m (distinct from enter)": {
			input:    []byte("\x1b[109;5u"),
			expected: KeyEvent{Key: KeyRune, Rune: 'm', Mod: ModCtrl},
		},
		"ctrl+i (distinct from tab)": {
			input:    []byte("\x1b[105;5u"),
			expected: KeyEvent{Key: KeyRune, Rune: 'i', Mod: ModCtrl},
		},
		"ctrl+a": {
			input:    []byte("\x1b[97;5u"),
			expected: KeyEvent{Key: KeyRune, Rune: 'a', Mod: ModCtrl},
		},
		"shift+enter": {
			input:    []byte("\x1b[13;2u"),
			expected: KeyEvent{Key: KeyEnter, Mod: ModShift},
		},
		"alt+backspace": {
			input:    []byte("\x1b[127;3u"),
			expected: KeyEvent{Key: KeyBackspace, Mod: ModAlt},
		},
		"ctrl+shift+a": {
			input:    []byte("\x1b[97;6u"),
			expected: KeyEvent{Key: KeyRune, Rune: 'a', Mod: ModCtrl | ModShift},
		},
		"f1": {
			input:    []byte("\x1b[57364;1u"),
			expected: KeyEvent{Key: KeyF1},
		},
		"f12": {
			input:    []byte("\x1b[57375;1u"),
			expected: KeyEvent{Key: KeyF12},
		},
		"up arrow": {
			input:    []byte("\x1b[57352;1u"),
			expected: KeyEvent{Key: KeyUp},
		},
		"down arrow": {
			input:    []byte("\x1b[57353;1u"),
			expected: KeyEvent{Key: KeyDown},
		},
		"right arrow": {
			input:    []byte("\x1b[57354;1u"),
			expected: KeyEvent{Key: KeyRight},
		},
		"left arrow": {
			input:    []byte("\x1b[57355;1u"),
			expected: KeyEvent{Key: KeyLeft},
		},
		"home": {
			input:    []byte("\x1b[57345;1u"),
			expected: KeyEvent{Key: KeyHome},
		},
		"end": {
			input:    []byte("\x1b[57346;1u"),
			expected: KeyEvent{Key: KeyEnd},
		},
		"insert": {
			input:    []byte("\x1b[57348;1u"),
			expected: KeyEvent{Key: KeyInsert},
		},
		"delete": {
			input:    []byte("\x1b[57349;1u"),
			expected: KeyEvent{Key: KeyDelete},
		},
		"page up": {
			input:    []byte("\x1b[57350;1u"),
			expected: KeyEvent{Key: KeyPageUp},
		},
		"page down": {
			input:    []byte("\x1b[57351;1u"),
			expected: KeyEvent{Key: KeyPageDown},
		},
		"no modifier param": {
			input:    []byte("\x1b[97u"),
			expected: KeyEvent{Key: KeyRune, Rune: 'a'},
		},
		"space": {
			input:    []byte("\x1b[32;1u"),
			expected: KeyEvent{Key: KeyRune, Rune: ' '},
		},
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

func TestParseKittyQueryResponse(t *testing.T) {
	type tc struct {
		input    []byte
		expected bool
	}

	tests := map[string]tc{
		"valid flag 1":        {input: []byte("\x1b[?1u"), expected: true},
		"valid flag 3":        {input: []byte("\x1b[?3u"), expected: true},
		"valid flag 5":        {input: []byte("\x1b[?5u"), expected: true},
		"flag 0 no disambig":  {input: []byte("\x1b[?0u"), expected: false},
		"flag 2 no disambig":  {input: []byte("\x1b[?2u"), expected: false},
		"empty":               {input: []byte{}, expected: false},
		"no question mark":    {input: []byte("\x1b[1u"), expected: false},
		"no digits":           {input: []byte("\x1b[?u"), expected: false},
		"wrong terminator":    {input: []byte("\x1b[?1x"), expected: false},
		"with prefix garbage": {input: []byte("garbage\x1b[?1u"), expected: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := parseKittyQueryResponse(tt.input)
			if got != tt.expected {
				t.Errorf("parseKittyQueryResponse(%q) = %v, want %v", tt.input, got, tt.expected)
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

