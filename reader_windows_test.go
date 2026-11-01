//go:build windows

package tui

import (
	"testing"

	"golang.org/x/sys/windows"
)

func TestTranslateMouseEvent_Motion(t *testing.T) {
	type tc struct {
		motion  bool
		buttons uint32 // latched from a previous press
		wantOK  bool
		want    MouseEvent
	}

	tests := map[string]tc{
		"move without motion reporting is dropped": {motion: false, wantOK: false},
		"move with motion reporting": {
			motion: true,
			wantOK: true,
			want:   MouseEvent{Button: MouseNone, Action: MouseMove, X: 4, Y: 2},
		},
		"move with a button held is a drag either way": {
			motion:  false,
			buttons: windows.FROM_LEFT_1ST_BUTTON_PRESSED,
			wantOK:  true,
			want:    MouseEvent{Button: MouseLeft, Action: MouseDrag, X: 4, Y: 2},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := &stdinReader{lastMouseButtons: tt.buttons}
			r.setMouseMotion(tt.motion)
			got, ok := r.translateMouseEvent(&mouseEventRecord{
				mousePosition: coord{X: 4, Y: 2},
				buttonState:   tt.buttons,
				eventFlags:    windows.MOUSE_MOVED,
			})
			if ok != tt.wantOK || got != tt.want {
				t.Errorf("translateMouseEvent = (%+v, %v), want (%+v, %v)", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
