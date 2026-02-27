package tui

import "unicode/utf8"

type styledTokenKind int

const (
	tokenRune    styledTokenKind = iota // printable rune (runeWidth set)
	tokenNewline                        // '\n'
	tokenANSI                           // complete ANSI escape sequence
)

// styledByteScanner iterates over styled byte data, yielding tokens that are
// either printable runes (with display width), newlines, or ANSI escape sequences.
// Control characters (except \n, \t) are silently dropped. Tabs become spaces.
type styledByteScanner struct {
	data []byte
	pos  int

	// Current token (valid after next() returns true).
	kind      styledTokenKind
	start     int // start offset in data
	end       int // end offset in data (exclusive)
	runeWidth int // display width for tokenRune
	runeVal   rune
}

func (s *styledByteScanner) reset(data []byte) {
	s.data = data
	s.pos = 0
}

func (s *styledByteScanner) bytes() []byte {
	if s.kind == tokenRune && s.runeVal == ' ' && s.start < len(s.data) && s.data[s.start] == '\t' {
		// Tab was replaced with space — return space byte.
		return []byte{' '}
	}
	return s.data[s.start:s.end]
}

func (s *styledByteScanner) next() bool {
	for s.pos < len(s.data) {
		b := s.data[s.pos]

		// ANSI escape sequence.
		if b == 0x1b {
			return s.scanANSI()
		}

		// Newline.
		if b == '\n' {
			s.kind = tokenNewline
			s.start = s.pos
			s.pos++
			s.end = s.pos
			return true
		}

		// Tab → space.
		if b == '\t' {
			s.kind = tokenRune
			s.start = s.pos
			s.pos++
			s.end = s.pos
			s.runeVal = ' '
			s.runeWidth = 1
			return true
		}

		// Decode UTF-8 rune.
		r, size := utf8.DecodeRune(s.data[s.pos:])
		if r == utf8.RuneError && size == 1 {
			s.pos++
			continue
		}

		// Drop C0 control chars and DEL.
		if r < 0x20 || r == 0x7f {
			s.pos += size
			continue
		}

		s.kind = tokenRune
		s.start = s.pos
		s.pos += size
		s.end = s.pos
		s.runeVal = r
		s.runeWidth = RuneWidth(r)
		return true
	}
	return false
}

func (s *styledByteScanner) scanANSI() bool {
	start := s.pos
	s.pos++ // skip ESC

	if s.pos >= len(s.data) {
		// Lone ESC at end — skip it.
		return s.next()
	}

	switch s.data[s.pos] {
	case '[':
		// CSI: ESC [ ... final-byte (0x40-0x7e)
		s.pos++
		for s.pos < len(s.data) {
			c := s.data[s.pos]
			s.pos++
			if c >= 0x40 && c <= 0x7e {
				break
			}
		}
	case ']':
		// OSC: ESC ] ... BEL or ST
		s.pos++
		for s.pos < len(s.data) {
			c := s.data[s.pos]
			s.pos++
			if c == 0x07 {
				break
			}
			if c == 0x1b && s.pos < len(s.data) && s.data[s.pos] == '\\' {
				s.pos++
				break
			}
		}
	default:
		// Generic 2-byte escape.
		s.pos++
	}

	s.kind = tokenANSI
	s.start = start
	s.end = s.pos
	return true
}
