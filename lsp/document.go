// SPDX-License-Identifier: GPL-3.0-only

package lsp

import (
	"sync"

	"go.lsp.dev/protocol"
)

// Document represents an open file tracked by the LSP server.
type Document struct {
	URI     protocol.DocumentURI
	Content string
	Version int32
}

// Documents is a thread-safe store of open text documents. The server
// keeps one of these for the lifetime of the connection and updates it
// in response to didOpen / didChange / didClose notifications.
type Documents struct {
	mu   sync.RWMutex
	docs map[protocol.DocumentURI]*Document
}

func NewDocuments() *Documents {
	return &Documents{docs: make(map[protocol.DocumentURI]*Document)}
}

func (d *Documents) Open(uri protocol.DocumentURI, content string, version int32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.docs[uri] = &Document{URI: uri, Content: content, Version: version}
}

func (d *Documents) Change(uri protocol.DocumentURI, content string, version int32) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if doc, ok := d.docs[uri]; ok {
		doc.Content = content
		doc.Version = version
	}
}

// ApplyIncremental applies incremental content changes to a document.
func (d *Documents) ApplyIncremental(
	uri protocol.DocumentURI,
	changes []protocol.TextDocumentContentChangeEvent,
	version int32,
) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	doc, ok := d.docs[uri]
	if !ok {
		return ""
	}
	for _, ch := range changes {
		if ch.Range == (protocol.Range{}) {
			// Full replacement.
			doc.Content = ch.Text
		} else {
			doc.Content = applyEdit(doc.Content, ch.Range, ch.Text)
		}
	}
	doc.Version = version
	return doc.Content
}

func applyEdit(content string, r protocol.Range, newText string) string {
	startOff := posToOffset(content, r.Start)
	endOff := posToOffset(content, r.End)
	if startOff < 0 || endOff < 0 || startOff > endOff {
		return content
	}
	return content[:startOff] + newText + content[endOff:]
}

func posToOffset(content string, pos protocol.Position) int {
	line := uint32(0)
	for i, ch := range content {
		if line == pos.Line {
			col := uint32(0)
			for j := i; j < len(content); j++ {
				if col == pos.Character {
					return j
				}
				if content[j] == '\n' {
					break
				}
				col++
			}
			return i + int(pos.Character)
		}
		if ch == '\n' {
			line++
		}
	}
	// Past end of content.
	if line == pos.Line {
		return len(content)
	}
	return -1
}

// ContentForURI returns the current content string for the given URI.
// Used by DidChange after applying incremental changes.
func (d *Documents) ContentForURI(uri protocol.DocumentURI) string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if doc, ok := d.docs[uri]; ok {
		return doc.Content
	}
	return ""
}

func (d *Documents) Close(uri protocol.DocumentURI) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.docs, uri)
}

func (d *Documents) Get(uri protocol.DocumentURI) (*Document, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	doc, ok := d.docs[uri]
	return doc, ok
}

// All returns a snapshot of every open document.
func (d *Documents) All() []Document {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]Document, 0, len(d.docs))
	for _, doc := range d.docs {
		out = append(out, *doc)
	}
	return out
}
