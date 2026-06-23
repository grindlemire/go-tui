package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

var testBin string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "tui-integration-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmp)

	testBin = filepath.Join(tmp, "tui")
	if runtime.GOOS == "windows" {
		testBin += ".exe"
	}
	cmd := exec.Command("go", "build", "-o", testBin, ".")
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n%s\n", err, out)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func TestCLI_Check(t *testing.T) {
	gsxFiles, _ := filepath.Glob("testdata/*.gsx")
	if len(gsxFiles) == 0 {
		t.Skip("no testdata/*.gsx files found")
	}

	for _, gsxFile := range gsxFiles {
		t.Run(filepath.Base(gsxFile), func(t *testing.T) {
			cmd := exec.Command(testBin, "check", gsxFile)
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Errorf("check %s failed: %v\n%s", gsxFile, err, out)
			}
		})
	}
}

func TestCLI_Fmt_Stdout(t *testing.T) {
	gsxFiles, _ := filepath.Glob("testdata/*.gsx")
	if len(gsxFiles) == 0 {
		t.Skip("no testdata/*.gsx files found")
	}

	for _, gsxFile := range gsxFiles {
		t.Run(filepath.Base(gsxFile), func(t *testing.T) {
			cmd := exec.Command(testBin, "fmt", "--stdout", gsxFile)
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

func TestCLI_Generate_Output(t *testing.T) {
	src := filepath.Join("testdata", "simple.gsx")
	if _, err := os.Stat(src); err != nil {
		t.Skipf("missing fixture %s: %v", src, err)
	}

	// Use a nested, non-existent directory to also exercise dir creation.
	outDir := filepath.Join(t.TempDir(), "gen", "nested")

	cmd := exec.Command(testBin, "generate", "-o", outDir, src)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generate -o failed: %v\n%s", err, out)
	}

	want := filepath.Join(outDir, "simple_gsx.go")
	if _, err := os.Stat(want); err != nil {
		t.Errorf("expected generated file %s: %v", want, err)
	}
}

func TestCLI_Generate_Output_MissingValue(t *testing.T) {
	cmd := exec.Command(testBin, "generate", "-o")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected error when -o has no value, got success:\n%s", out)
	}
}

func TestCLI_Version(t *testing.T) {
	cmd := exec.Command(testBin, "version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("version failed: %v\n%s", err, out)
	}
}

func TestCLI_Help(t *testing.T) {
	cmd := exec.Command(testBin, "help")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("help failed: %v\n%s", err, out)
	}
	if len(out) == 0 {
		t.Error("help output should not be empty")
	}
}
