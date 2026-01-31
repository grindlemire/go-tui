package lsp

import (
	"regexp"
	"strings"
	"sync"

	"github.com/grindlemire/go-tui/internal/tuigen"
)

// ComponentInfo stores information about a component definition.
type ComponentInfo struct {
	Name     string
	Location Location
	Params   []*tuigen.Param
}

// FuncInfo stores information about a helper function definition.
type FuncInfo struct {
	Name      string
	Location  Location
	Signature string // full function signature
	Params    []FuncParam
	Returns   string
}

// FuncParam represents a function parameter.
type FuncParam struct {
	Name     string
	Type     string
	Position Position // position within the function signature
}

// ParamInfo stores information about a component parameter usage.
type ParamInfo struct {
	Name          string
	Type          string
	ComponentName string
	Location      Location
}

// ComponentIndex tracks all components, functions, and parameters across the workspace.
type ComponentIndex struct {
	mu sync.RWMutex
	// component name -> component info
	Components map[string]*ComponentInfo
	// file URI -> component names defined in that file
	FileComponents map[string][]string
	// function name -> function info
	Functions map[string]*FuncInfo
	// file URI -> function names defined in that file
	FileFunctions map[string][]string
	// "componentName.paramName" -> param info
	Params map[string]*ParamInfo
	// file URI -> param keys defined in that file
	FileParams map[string][]string
}

// NewComponentIndex creates a new component index.
func NewComponentIndex() *ComponentIndex {
	return &ComponentIndex{
		Components:     make(map[string]*ComponentInfo),
		FileComponents: make(map[string][]string),
		Functions:      make(map[string]*FuncInfo),
		FileFunctions:  make(map[string][]string),
		Params:         make(map[string]*ParamInfo),
		FileParams:     make(map[string][]string),
	}
}

// Add adds a component from a parsed file to the index.
func (idx *ComponentIndex) Add(uri string, comp *tuigen.Component) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	info := &ComponentInfo{
		Name:   comp.Name,
		Params: comp.Params,
		Location: Location{
			URI: uri,
			Range: Range{
				Start: Position{
					Line:      comp.Position.Line - 1, // tuigen is 1-indexed, LSP is 0-indexed
					Character: comp.Position.Column - 1,
				},
				End: Position{
					Line:      comp.Position.Line - 1,
					Character: comp.Position.Column - 1 + len("@component") + 1 + len(comp.Name),
				},
			},
		},
	}

	idx.Components[comp.Name] = info
	idx.FileComponents[uri] = append(idx.FileComponents[uri], comp.Name)
}

// Remove removes all components, functions, and params from a file.
func (idx *ComponentIndex) Remove(uri string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Remove components that were defined in this file
	for _, name := range idx.FileComponents[uri] {
		delete(idx.Components, name)
	}
	delete(idx.FileComponents, uri)

	// Remove functions that were defined in this file
	for _, name := range idx.FileFunctions[uri] {
		delete(idx.Functions, name)
	}
	delete(idx.FileFunctions, uri)

	// Remove params that were defined in this file
	for _, key := range idx.FileParams[uri] {
		delete(idx.Params, key)
	}
	delete(idx.FileParams, uri)
}

// Lookup finds a component by name.
func (idx *ComponentIndex) Lookup(name string) (*ComponentInfo, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	info, ok := idx.Components[name]
	return info, ok
}

// All returns all component names in the index.
func (idx *ComponentIndex) All() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	names := make([]string, 0, len(idx.Components))
	for name := range idx.Components {
		names = append(names, name)
	}
	return names
}

// ComponentsInFile returns all component names defined in a file.
func (idx *ComponentIndex) ComponentsInFile(uri string) []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.FileComponents[uri]
}

// IndexDocument indexes all components, functions, and params from a document.
func (idx *ComponentIndex) IndexDocument(uri string, ast *tuigen.File) {
	// First remove old entries for this file
	idx.Remove(uri)

	if ast == nil {
		return
	}

	// Add all components from the file
	for _, comp := range ast.Components {
		idx.Add(uri, comp)
		// Index component parameters
		idx.AddComponentParams(uri, comp)
	}

	// Add all functions from the file
	for _, fn := range ast.Funcs {
		idx.AddFunc(uri, fn)
	}
}

// GetInfo returns component info for a given name.
func (idx *ComponentIndex) GetInfo(name string) *ComponentInfo {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return idx.Components[name]
}

// AddFunc adds a function from a parsed file to the index.
func (idx *ComponentIndex) AddFunc(uri string, fn *tuigen.GoFunc) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Parse function name and signature from the code
	name, sig, params, returns := parseFuncSignature(fn.Code)
	if name == "" {
		return
	}

	// Adjust param positions from code-relative to document-absolute (0-indexed)
	for i := range params {
		params[i].Position = Position{
			Line:      fn.Position.Line - 1,
			Character: fn.Position.Column - 1 + params[i].Position.Character,
		}
	}

	info := &FuncInfo{
		Name:      name,
		Signature: sig,
		Params:    params,
		Returns:   returns,
		Location: Location{
			URI: uri,
			Range: Range{
				Start: Position{
					Line:      fn.Position.Line - 1,
					Character: fn.Position.Column - 1,
				},
				End: Position{
					Line:      fn.Position.Line - 1,
					Character: fn.Position.Column - 1 + len("func") + 1 + len(name),
				},
			},
		},
	}

	idx.Functions[name] = info
	idx.FileFunctions[uri] = append(idx.FileFunctions[uri], name)
}

// AddComponentParams indexes all parameters for a component.
func (idx *ComponentIndex) AddComponentParams(uri string, comp *tuigen.Component) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	for _, p := range comp.Params {
		key := comp.Name + "." + p.Name
		info := &ParamInfo{
			Name:          p.Name,
			Type:          p.Type,
			ComponentName: comp.Name,
			Location: Location{
				URI: uri,
				Range: Range{
					Start: Position{
						Line:      p.Position.Line - 1,
						Character: p.Position.Column - 1,
					},
					End: Position{
						Line:      p.Position.Line - 1,
						Character: p.Position.Column - 1 + len(p.Name),
					},
				},
			},
		}
		idx.Params[key] = info
		idx.FileParams[uri] = append(idx.FileParams[uri], key)
	}
}

// LookupFunc finds a function by name.
func (idx *ComponentIndex) LookupFunc(name string) (*FuncInfo, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	info, ok := idx.Functions[name]
	return info, ok
}

// LookupParam finds a parameter by component name and param name.
func (idx *ComponentIndex) LookupParam(componentName, paramName string) (*ParamInfo, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	key := componentName + "." + paramName
	info, ok := idx.Params[key]
	return info, ok
}

// LookupParamInAnyComponent finds a parameter by name in any component.
func (idx *ComponentIndex) LookupParamInAnyComponent(paramName string) (*ParamInfo, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	for _, info := range idx.Params {
		if info.Name == paramName {
			return info, true
		}
	}
	return nil, false
}

// LookupFuncParam finds a function parameter by function name and param name.
// Returns the param and the function's URI.
func (idx *ComponentIndex) LookupFuncParam(funcName, paramName string) (*FuncParam, string, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	funcInfo, ok := idx.Functions[funcName]
	if !ok {
		return nil, "", false
	}
	for i := range funcInfo.Params {
		if funcInfo.Params[i].Name == paramName {
			return &funcInfo.Params[i], funcInfo.Location.URI, true
		}
	}
	return nil, "", false
}

// AllFunctions returns all function names in the index.
func (idx *ComponentIndex) AllFunctions() []string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	names := make([]string, 0, len(idx.Functions))
	for name := range idx.Functions {
		names = append(names, name)
	}
	return names
}

// parseFuncSignature extracts function name, signature, params and return type from Go code.
func parseFuncSignature(code string) (name, signature string, params []FuncParam, returns string) {
	// Match: func name(params) returns
	re := regexp.MustCompile(`func\s+(\w+)\s*\(([^)]*)\)\s*([^{]*)`)
	matches := re.FindStringSubmatch(code)
	if len(matches) < 2 {
		return "", "", nil, ""
	}

	name = matches[1]
	paramStr := ""
	if len(matches) > 2 {
		paramStr = strings.TrimSpace(matches[2])
	}
	if len(matches) > 3 {
		returns = strings.TrimSpace(matches[3])
	}

	// Build signature
	signature = "func " + name + "(" + paramStr + ")"
	if returns != "" {
		signature += " " + returns
	}

	// Parse params with positions relative to code start
	if paramStr != "" {
		// Find the opening paren in the original code to calculate positions
		parenIdx := strings.Index(code, "(")
		paramContentStart := parenIdx + 1

		paramParts := strings.Split(paramStr, ",")
		offset := 0
		for _, rawPart := range paramParts {
			trimmed := strings.TrimSpace(rawPart)
			fields := strings.Fields(trimmed)
			if len(fields) >= 2 {
				// Find where the name starts within the raw part
				nameInPart := strings.Index(rawPart, fields[0])
				charPos := paramContentStart + offset + nameInPart
				params = append(params, FuncParam{
					Name: fields[0],
					Type: strings.Join(fields[1:], " "),
					Position: Position{
						Character: charPos, // relative to code start; adjusted to absolute in AddFunc
					},
				})
			} else if len(fields) == 1 {
				// Type only, no name (or name only)
				params = append(params, FuncParam{
					Name: fields[0],
				})
			}
			offset += len(rawPart) + 1 // +1 for comma
		}
	}

	return name, signature, params, returns
}
