// SPDX-License-Identifier: GPL-3.0-only

package target

import (
	"context"
	"testing"
)

func TestMemREST_ReplaysCannedResponse(t *testing.T) {
	mock := NewMemREST(map[string]MemRESTResponse{
		"GET /v1/sites": {
			StatusCode: 200,
			Body:       []byte(`{"sites":[]}`),
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
		},
	})

	resp, err := mock.Do(context.Background(), HTTPRequest{
		Method: "GET",
		Path:   "/v1/sites",
	})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if string(resp.Body) != `{"sites":[]}` {
		t.Errorf("body = %q", resp.Body)
	}
	if got := resp.Headers["Content-Type"]; len(got) != 1 || got[0] != "application/json" {
		t.Errorf("Content-Type header = %v", got)
	}
}

func TestMemREST_UnmatchedRouteReturns404(t *testing.T) {
	mock := NewMemREST(nil)

	resp, err := mock.Do(context.Background(), HTTPRequest{
		Method: "POST",
		Path:   "/missing",
	})
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("status = %d, want 404", resp.StatusCode)
	}
	if want := `"no route for POST /missing"`; !contains(string(resp.Body), want) {
		t.Errorf("body %q missing %q", resp.Body, want)
	}
}

func TestMemREST_RecordsCallsInOrder(t *testing.T) {
	mock := NewMemREST(nil)
	ctx := context.Background()

	_, _ = mock.Do(ctx, HTTPRequest{Method: "POST", Path: "/a", Body: []byte("first")})
	_, _ = mock.Do(ctx, HTTPRequest{Method: "GET", Path: "/b"})
	_, _ = mock.Do(ctx, HTTPRequest{Method: "POST", Path: "/a", Body: []byte("second")})

	calls := mock.Calls()
	if len(calls) != 3 {
		t.Fatalf("got %d calls, want 3", len(calls))
	}
	if calls[0].Path != "/a" || string(calls[0].Body) != "first" {
		t.Errorf("call[0] = %+v", calls[0])
	}
	if calls[1].Method != "GET" {
		t.Errorf("call[1] = %+v", calls[1])
	}
	if calls[2].Path != "/a" || string(calls[2].Body) != "second" {
		t.Errorf("call[2] = %+v", calls[2])
	}

	matching := mock.CallsMatching("POST", "/a")
	if len(matching) != 2 {
		t.Errorf("CallsMatching = %d, want 2", len(matching))
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
