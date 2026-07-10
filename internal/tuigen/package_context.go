package tuigen

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PackageContext holds declarations from sibling files of the same package,
// so generating or analyzing one .gsx file can detect collisions with code
// declared elsewhere in the package. Callers decide which files to feed in;
// generated files (*_gsx.go) and test files must be excluded, or the
// generator would detect its own previous output as user code.
type PackageContext struct {
	methods         map[string]map[string]bool // type name -> method name
	funcs           map[string]bool            // top-level function names
	types           map[string]bool            // top-level type names
	templs          map[string]bool            // function templ names
	renderReceivers map[string]bool            // receiver types with a method templ
}

// NewPackageContext creates an empty package context.
func NewPackageContext() *PackageContext {
	return &PackageContext{
		methods:         make(map[string]map[string]bool),
		funcs:           make(map[string]bool),
		types:           make(map[string]bool),
		templs:          make(map[string]bool),
		renderReceivers: make(map[string]bool),
	}
}

// AddGoSource parses a sibling .go file and records its top-level methods,
// functions, and types.
func (c *PackageContext) AddGoSource(filename, src string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, src, parser.SkipObjectResolution)
	if err != nil {
		return err
	}
	c.addParsedFile(f)
	return nil
}

// AddGSXFile records the declarations a sibling .gsx file contributes to the
// package: its top-level Go functions and types (which pass through to the
// generated file), its function templ names, and the receiver types of its
// method templs.
func (c *PackageContext) AddGSXFile(file *File) {
	for _, fn := range file.Funcs {
		// GoFunc.Code is a complete declaration; parse it as a tiny file so
		// method receivers are extracted accurately (not by regex). Snippets
		// that fail to parse are skipped; the sibling reports its own errors.
		c.addSnippet(fn.Code)
	}
	for _, decl := range file.Decls {
		c.addSnippet(decl.Code)
	}
	for _, comp := range file.Components {
		if comp.Receiver == "" {
			c.templs[comp.Name] = true
		} else {
			c.renderReceivers[strings.TrimPrefix(comp.ReceiverType, "*")] = true
		}
	}
}

// HasMethod reports whether a sibling file declares the method on the type.
func (c *PackageContext) HasMethod(typeName, methodName string) bool {
	return c.methods[typeName][methodName]
}

// HasFunc reports whether a sibling file declares a top-level function.
func (c *PackageContext) HasFunc(name string) bool {
	return c.funcs[name]
}

// HasType reports whether a sibling file declares a type.
func (c *PackageContext) HasType(name string) bool {
	return c.types[name]
}

// HasTempl reports whether a sibling .gsx file declares a function templ.
func (c *PackageContext) HasTempl(name string) bool {
	return c.templs[name]
}

// HasRenderTempl reports whether a sibling .gsx file declares a method templ
// for the receiver type.
func (c *PackageContext) HasRenderTempl(typeName string) bool {
	return c.renderReceivers[typeName]
}

// generatedHeader matches the standard Go generated-file marker
// (https://go.dev/s/generatedcode) before the package clause.
var generatedHeader = regexp.MustCompile(`(?m)^// Code generated .* DO NOT EDIT\.$`)

// AddDirectory scans dir and adds every sibling source the context should
// know about: plain .go files and .gsx files. Always excluded: generated
// files (by _gsx.go suffix or generated-code header, so the generator never
// treats its own previous output as user code) and _test.go files (a
// test-only method must not suppress a method production builds need).
// skip filters further by filename; callers use it to exclude the file being
// processed and files they source elsewhere (e.g. open editor buffers).
// Siblings that fail to parse are skipped; they report their own errors when
// processed. Errors reading the directory degrade to no additions.
func (c *PackageContext) AddDirectory(dir string, skip func(filename string) bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || (skip != nil && skip(name)) {
			continue
		}

		switch {
		case strings.HasSuffix(name, ".gsx"):
			source, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				continue
			}
			lexer := NewLexer(name, string(source))
			parser := NewParser(lexer)
			file, err := parser.ParseFile()
			if err != nil || file == nil {
				continue
			}
			c.AddGSXFile(file)

		case strings.HasSuffix(name, ".go"):
			if strings.HasSuffix(name, "_gsx.go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			source, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil || isGeneratedFile(source) {
				continue
			}
			_ = c.AddGoSource(name, string(source))
		}
	}
}

// isGeneratedFile reports whether the source carries a generated-code header
// before its package clause.
func isGeneratedFile(source []byte) bool {
	src := source
	if idx := strings.Index(string(source), "\npackage "); idx >= 0 {
		src = source[:idx]
	}
	return generatedHeader.Match(src)
}

func (c *PackageContext) addSnippet(code string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "snippet.go", "package _\n"+code, parser.SkipObjectResolution)
	if err != nil {
		return
	}
	c.addParsedFile(f)
}

func (c *PackageContext) addParsedFile(f *ast.File) {
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Recv == nil || len(d.Recv.List) == 0 {
				c.funcs[d.Name.Name] = true
				continue
			}
			if typeName := receiverTypeName(d.Recv.List[0].Type); typeName != "" {
				if c.methods[typeName] == nil {
					c.methods[typeName] = make(map[string]bool)
				}
				c.methods[typeName][d.Name.Name] = true
			}
		case *ast.GenDecl:
			if d.Tok != token.TYPE {
				continue
			}
			for _, spec := range d.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					c.types[ts.Name.Name] = true
				}
			}
		}
	}
}

// receiverTypeName extracts the base type name from a receiver expression,
// unwrapping pointers and generic type parameters.
func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	case *ast.IndexExpr:
		return receiverTypeName(t.X)
	case *ast.IndexListExpr:
		return receiverTypeName(t.X)
	}
	return ""
}
