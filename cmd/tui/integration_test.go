package main

import (
	"os/exec"
	"path/filepath"
	"testing"
)

func buildCLI(t *testing.T) string {
	t.Helper()
	tmpBin := filepath.Join(t.TempDir(), "tui")
	cmd := exec.Command("go", "build", "-o", tmpBin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}
	return tmpBin
}

func TestCLI_Check(t *testing.T) {
	bin := buildCLI(t)

	gsxFiles, _ := filepath.Glob("testdata/*.gsx")
	if len(gsxFiles) == 0 {
		t.Skip("no testdata/*.gsx files found")
	}

	for _, gsxFile := range gsxFiles {
		t.Run(filepath.Base(gsxFile), func(t *testing.T) {
			cmd := exec.Command(bin, "check", gsxFile)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("check %s failed: %v\n%s", gsxFile, err, out)
			}
		})
	}
}

func TestCLI_Fmt_Stdout(t *testing.T) {
	bin := buildCLI(t)

	gsxFiles, _ := filepath.Glob("testdata/*.gsx")
	if len(gsxFiles) == 0 {
		t.Skip("no testdata/*.gsx files found")
	}

	for _, gsxFile := range gsxFiles {
		t.Run(filepath.Base(gsxFile), func(t *testing.T) {
			// --stdout formats without modifying and writes to stdout
			cmd := exec.Command(bin, "fmt", "--stdout", gsxFile)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("fmt --stdout %s failed: %v\n%s", gsxFile, err, out)
			}
			if len(out) == 0 {
				t.Errorf("fmt --stdout %s produced empty output", gsxFile)
			}
		})
	}
}

func TestCLI_Version(t *testing.T) {
	bin := buildCLI(t)

	cmd := exec.Command(bin, "version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("version failed: %v\n%s", err, out)
	}
}

func TestCLI_Help(t *testing.T) {
	bin := buildCLI(t)

	cmd := exec.Command(bin, "help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("help failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Error("help output should not be empty")
	}
}
