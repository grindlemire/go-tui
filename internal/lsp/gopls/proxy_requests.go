package gopls

import (
	"encoding/json"
	"fmt"
)

// OpenVirtualFile opens a virtual .go file in gopls.
func (p *GoplsProxy) OpenVirtualFile(uri, content string, version int) error {
	p.virtualFilesMu.Lock()
	p.virtualFiles[uri] = &VirtualFile{
		URI:     uri,
		Content: content,
		Version: version,
	}
	p.virtualFilesMu.Unlock()

	params := map[string]any{
		"textDocument": map[string]any{
			"uri":        uri,
			"languageId": "go",
			"version":    version,
			"text":       content,
		},
	}

	return p.notify("textDocument/didOpen", params)
}

// UpdateVirtualFile updates a virtual .go file in gopls.
func (p *GoplsProxy) UpdateVirtualFile(uri, content string, version int) error {
	p.virtualFilesMu.Lock()
	if vf, ok := p.virtualFiles[uri]; ok {
		vf.Content = content
		vf.Version = version
	}
	p.virtualFilesMu.Unlock()

	params := map[string]any{
		"textDocument": map[string]any{
			"uri":     uri,
			"version": version,
		},
		"contentChanges": []map[string]any{
			{"text": content},
		},
	}

	return p.notify("textDocument/didChange", params)
}

// CloseVirtualFile closes a virtual .go file in gopls.
func (p *GoplsProxy) CloseVirtualFile(uri string) error {
	p.virtualFilesMu.Lock()
	delete(p.virtualFiles, uri)
	p.virtualFilesMu.Unlock()

	params := map[string]any{
		"textDocument": map[string]any{
			"uri": uri,
		},
	}

	return p.notify("textDocument/didClose", params)
}

// Completion requests completion items at a position.
func (p *GoplsProxy) Completion(uri string, pos Position) ([]CompletionItem, error) {
	params := CompletionParams{
		TextDocumentPositionParams: TextDocumentPositionParams{
			TextDocument: TextDocumentIdentifier{URI: uri},
			Position:     pos,
		},
	}

	result, err := p.call("textDocument/completion", params)
	if err != nil {
		return nil, err
	}

	// gopls returns either CompletionList or []CompletionItem
	var list CompletionList
	if err := json.Unmarshal(result, &list); err != nil {
		// Try as array
		var items []CompletionItem
		if err := json.Unmarshal(result, &items); err != nil {
			return nil, fmt.Errorf("parsing completion result: %w", err)
		}
		return items, nil
	}

	return list.Items, nil
}

// Hover requests hover information at a position.
func (p *GoplsProxy) Hover(uri string, pos Position) (*Hover, error) {
	params := TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     pos,
	}

	result, err := p.call("textDocument/hover", params)
	if err != nil {
		return nil, err
	}

	if result == nil || string(result) == "null" {
		return nil, nil
	}

	var hover Hover
	if err := json.Unmarshal(result, &hover); err != nil {
		return nil, fmt.Errorf("parsing hover result: %w", err)
	}

	return &hover, nil
}

// Definition requests the definition location at a position.
func (p *GoplsProxy) Definition(uri string, pos Position) ([]Location, error) {
	params := TextDocumentPositionParams{
		TextDocument: TextDocumentIdentifier{URI: uri},
		Position:     pos,
	}

	result, err := p.call("textDocument/definition", params)
	if err != nil {
		return nil, err
	}

	if result == nil || string(result) == "null" {
		return nil, nil
	}

	// gopls returns either Location or []Location
	var locs []Location
	if err := json.Unmarshal(result, &locs); err != nil {
		var loc Location
		if err := json.Unmarshal(result, &loc); err != nil {
			return nil, fmt.Errorf("parsing definition result: %w", err)
		}
		return []Location{loc}, nil
	}

	return locs, nil
}

// call sends a request and waits for the response.
func (p *GoplsProxy) call(method string, params any) (json.RawMessage, error) {
	id := p.nextID.Add(1)

	// Create response channel
	respChan := make(chan *Response, 1)
	p.pendingMu.Lock()
	p.pending[id] = respChan
	p.pendingMu.Unlock()

	defer func() {
		p.pendingMu.Lock()
		delete(p.pending, id)
		p.pendingMu.Unlock()
	}()

	req := Request{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := p.send(req); err != nil {
		return nil, err
	}

	select {
	case <-p.ctx.Done():
		return nil, p.ctx.Err()
	case resp := <-respChan:
		if resp.Error != nil {
			return nil, fmt.Errorf("gopls error %d: %s", resp.Error.Code, resp.Error.Message)
		}
		return resp.Result, nil
	}
}

// notify sends a notification (no response expected).
func (p *GoplsProxy) notify(method string, params any) error {
	req := Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	return p.send(req)
}
