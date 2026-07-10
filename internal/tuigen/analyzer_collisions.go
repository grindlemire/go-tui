package tuigen

import (
	"regexp"
	"strings"
)

// validateNameCollisions rejects declarations that would collide with
// generated code. Without these checks the collisions still fail, but as raw
// Go "redeclared" errors inside the generated file; catching them here points
// the error at the .gsx source with a hint.
func (a *Analyzer) validateNameCollisions(file *File) {
	seenFuncTempl := make(map[string]Position)
	seenMethodTempl := make(map[string]Position)

	for _, comp := range file.Components {
		if comp.Receiver == "" {
			a.validateFunctionTemplCollisions(file, comp, seenFuncTempl)
		} else {
			a.validateMethodTemplCollisions(file, comp, seenMethodTempl)
		}
	}
}

func (a *Analyzer) validateFunctionTemplCollisions(file *File, comp *Component, seen map[string]Position) {
	if first, ok := seen[comp.Name]; ok {
		a.errors.AddErrorf(comp.Position,
			"duplicate templ %q (first defined at %s)", comp.Name, first)
		return
	}
	seen[comp.Name] = comp.Position

	if a.pkgCtx != nil && a.pkgCtx.HasTempl(comp.Name) {
		a.errors.AddErrorf(comp.Position,
			"duplicate templ %q (also defined in another file of this package)", comp.Name)
		return
	}

	if hasTopLevelFunc(file.Decls, file.Funcs, comp.Name) ||
		(a.pkgCtx != nil && a.pkgCtx.HasFunc(comp.Name)) {
		a.errors.Add(NewErrorWithHint(comp.Position,
			"templ \""+comp.Name+"\" conflicts with a Go function of the same name",
			"rename the templ or the function"))
	}

	viewName := comp.Name + "View"
	if hasTypeDecl(file.Decls, viewName) ||
		(a.pkgCtx != nil && a.pkgCtx.HasType(viewName)) {
		a.errors.Add(NewErrorWithHint(comp.Position,
			"type \""+viewName+"\" conflicts with the view struct generated for templ \""+comp.Name+"\"",
			"rename the type; the generator reserves the <Name>View suffix for function templs"))
	}
}

func (a *Analyzer) validateMethodTemplCollisions(file *File, comp *Component, seen map[string]Position) {
	typeName := strings.TrimPrefix(comp.ReceiverType, "*")
	if first, ok := seen[typeName]; ok {
		a.errors.AddErrorf(comp.Position,
			"duplicate Render templ for receiver type %s (first defined at %s)", typeName, first)
		return
	}
	seen[typeName] = comp.Position

	if a.pkgCtx != nil && a.pkgCtx.HasRenderTempl(typeName) {
		a.errors.AddErrorf(comp.Position,
			"duplicate Render templ for receiver type %s (also defined in another file of this package)", typeName)
		return
	}

	if hasUserMethod(file.Decls, file.Funcs, comp.ReceiverType, "Render") ||
		(a.pkgCtx != nil && a.pkgCtx.HasMethod(typeName, "Render")) {
		a.errors.Add(NewErrorWithHint(comp.Position,
			typeName+" already declares a Render method; the templ generates Render",
			"remove the handwritten Render method or the templ"))
	}
}

// hasTopLevelFunc reports whether the file declares a plain (receiver-less)
// function with the given name.
func hasTopLevelFunc(decls []*GoDecl, funcs []*GoFunc, name string) bool {
	pattern := regexp.MustCompile(`func\s+` + regexp.QuoteMeta(name) + `\s*\(`)
	for _, decl := range decls {
		if decl.Kind == "func" && pattern.MatchString(decl.Code) {
			return true
		}
	}
	for _, fn := range funcs {
		if pattern.MatchString(fn.Code) {
			return true
		}
	}
	return false
}

// hasTypeDecl reports whether the file declares a type with the given name.
func hasTypeDecl(decls []*GoDecl, name string) bool {
	pattern := regexp.MustCompile(`type\s+` + regexp.QuoteMeta(name) + `\b`)
	for _, decl := range decls {
		if decl.Kind == "type" && pattern.MatchString(decl.Code) {
			return true
		}
	}
	return false
}
