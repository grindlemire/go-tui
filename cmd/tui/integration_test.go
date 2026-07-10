package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
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

// TestCLI_Generate_SiblingFileDetection verifies that a lifecycle method
// declared in a sibling .go file of the package suppresses the generated one,
// and that collisions with sibling declarations are reported at check time.
func TestCLI_Generate_SiblingFileDetection(t *testing.T) {
	writeFile := func(t *testing.T, dir, name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("writing %s: %v", name, err)
		}
	}

	t.Run("sibling UpdateProps suppresses generation", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "row.gsx", `package x

type row struct{ v string }

templ (r *row) Render() {
	<span>{r.v}</span>
}
`)
		writeFile(t, dir, "row_lifecycle.go", `package x

import tui "github.com/grindlemire/go-tui"

func (r *row) UpdateProps(fresh tui.Component) {
	r.updatePropsFields(fresh)
}
`)

		cmd := exec.Command(testBin, "generate", dir)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("generate failed: %v\n%s", err, out)
		}

		generated, err := os.ReadFile(filepath.Join(dir, "row_gsx.go"))
		if err != nil {
			t.Fatalf("reading generated file: %v", err)
		}
		code := string(generated)
		if strings.Contains(code, "func (r *row) UpdateProps(") {
			t.Errorf("generated UpdateProps despite sibling declaration:\n%s", code)
		}
		if !strings.Contains(code, "func (r *row) updatePropsFields(") {
			t.Errorf("missing updatePropsFields helper:\n%s", code)
		}
	})

	t.Run("generated sibling files are not treated as user code", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "row.gsx", `package x

type row struct{ v string }

templ (r *row) Render() {
	<span>{r.v}</span>
}
`)

		// Generate twice: the first run's output must not suppress the second.
		for i := range 2 {
			cmd := exec.Command(testBin, "generate", dir)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("generate run %d failed: %v\n%s", i+1, err, out)
			}
		}

		generated, err := os.ReadFile(filepath.Join(dir, "row_gsx.go"))
		if err != nil {
			t.Fatalf("reading generated file: %v", err)
		}
		if !strings.Contains(string(generated), "func (r *row) UpdateProps(") {
			t.Errorf("second generate run suppressed UpdateProps; generated output was treated as user code:\n%s", generated)
		}
	})

	t.Run("check reports collision with sibling function", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "foo.gsx", `package x

templ Foo() {
	<span>hi</span>
}
`)
		writeFile(t, dir, "helpers.go", `package x

func Foo() string { return "x" }
`)

		cmd := exec.Command(testBin, "check", filepath.Join(dir, "foo.gsx"))
		out, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("check succeeded, expected collision error\n%s", out)
		}
		if !strings.Contains(string(out), "conflicts with a Go function") {
			t.Errorf("output missing collision error:\n%s", out)
		}
	})

	t.Run("test file declarations are ignored", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, dir, "row.gsx", `package x

type row struct{ v string }

templ (r *row) Render() {
	<span>{r.v}</span>
}
`)
		writeFile(t, dir, "row_test.go", `package x

import tui "github.com/grindlemire/go-tui"

func (r *row) UpdateProps(fresh tui.Component) {}
`)

		cmd := exec.Command(testBin, "generate", dir)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("generate failed: %v\n%s", err, out)
		}

		generated, err := os.ReadFile(filepath.Join(dir, "row_gsx.go"))
		if err != nil {
			t.Fatalf("reading generated file: %v", err)
		}
		if !strings.Contains(string(generated), "func (r *row) UpdateProps(") {
			t.Errorf("test-file declaration suppressed UpdateProps; prod builds would lack the method:\n%s", generated)
		}
	})
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
