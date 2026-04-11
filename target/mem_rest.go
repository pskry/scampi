// SPDX-License-Identifier: GPL-3.0-only

package target

import (
	"context"
	"maps"
	"sync"

	"scampi.dev/scampi/capability"
)

// MemREST is an in-memory HTTPClient that replays canned responses
// keyed by "METHOD /path" and records every incoming request for
// later inspection. It is the REST counterpart to MemTarget — both
// are real `target.Target` implementations usable from Go tests and
// from the test framework runner.
//
// Routes that have no entry return a 404 with a small JSON body so
// the engine doesn't crash on unexpected calls.
type MemREST struct {
	mu     sync.Mutex
	routes map[string]MemRESTResponse
	calls  []MemRESTCall
}

// MemRESTResponse is a canned HTTP response stored in MemREST's
// route table.
type MemRESTResponse struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}

// MemRESTCall captures a single request made through MemREST so
// tests / verifiers can assert on what the engine actually sent.
// Field layout matches HTTPRequest exactly so the recording path
// can convert directly via MemRESTCall(req).
type MemRESTCall HTTPRequest

// NewMemREST builds a MemREST with the given route table. Routes is
// keyed by "METHOD /path" (e.g. "POST /v1/sites").
func NewMemREST(routes map[string]MemRESTResponse) *MemREST {
	return &MemREST{routes: routes}
}

// Capabilities reports the REST capability so the engine can
// downcast to HTTPClient.
func (m *MemREST) Capabilities() capability.Capability {
	return capability.REST
}

// Do records the request and replays the matching route, or returns
// a 404 if no route matches.
func (m *MemREST) Do(_ context.Context, req HTTPRequest) (*HTTPResponse, error) {
	m.mu.Lock()
	m.calls = append(m.calls, MemRESTCall(req))
	m.mu.Unlock()

	key := req.Method + " " + req.Path
	if resp, ok := m.routes[key]; ok {
		headers := make(map[string][]string, len(resp.Headers))
		maps.Copy(headers, resp.Headers)
		return &HTTPResponse{
			StatusCode: resp.StatusCode,
			Headers:    headers,
			Body:       resp.Body,
		}, nil
	}

	return &HTTPResponse{
		StatusCode: 404,
		Body:       []byte(`{"error":"no route for ` + key + `"}`),
	}, nil
}

// Calls returns a snapshot of every recorded request, in order.
func (m *MemREST) Calls() []MemRESTCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]MemRESTCall, len(m.calls))
	copy(out, m.calls)
	return out
}

// CallsMatching returns the recorded requests whose method and path
// match the given pair. Useful for unit tests; the test framework
// verifier uses Calls() and does its own matching.
func (m *MemREST) CallsMatching(method, path string) []MemRESTCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []MemRESTCall
	for _, c := range m.calls {
		if c.Method == method && c.Path == path {
			out = append(out, c)
		}
	}
	return out
}
