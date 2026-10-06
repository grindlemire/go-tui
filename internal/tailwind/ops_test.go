package tailwind

import "testing"

func TestBorderStyleString(t *testing.T) {
      type tc struct {
          in   BorderStyle
          want string
      }

      tests := map[string]tc{
          "none":    {in: BorderNone, want: "BorderNone"},
          "single":  {in: BorderSingle, want: "BorderSingle"},
          "double":  {in: BorderDouble, want: "BorderDouble"},
          "rounded": {in: BorderRounded, want: "BorderRounded"},
          "thick":   {in: BorderThick, want: "BorderThick"},
          "unknown": {in: BorderStyle(99), want: "BorderStyle(99)"},
      }

      for name, tt := range tests {
          t.Run(name, func(t *testing.T) {
              if got := tt.in.String(); got != tt.want {
                  t.Errorf("BorderStyle(%d).String() = %q, want %q", int(tt.in), got, tt.want)               }
          })
      }
  }