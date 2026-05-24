package release

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// fakeServer returns an httptest.Server that responds with the given
// status code and body. The body is a static, test-controlled string
// returned to the aikata HTTP client (never a browser), so the
// no-direct-write-to-responsewriter audit rule does not apply.
//
//nolint:gosec // test fixture; body is a literal in this file, not user input
func fakeServer(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
			t.Errorf("missing Accept header: got %q", got)
		}
		if got := r.Header.Get("User-Agent"); got == "" {
			t.Errorf("missing User-Agent header")
		}
		w.WriteHeader(status)
		// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(ts.Close)
	return ts
}

func TestCheckLatest_UpToDate(t *testing.T) {
	ts := fakeServer(t, http.StatusOK, `{"tag_name":"v0.4.1","html_url":"https://x/v0.4.1"}`)
	c := &Client{Endpoint: ts.URL}
	r, err := c.CheckLatest(context.Background(), "v0.4.1")
	if err != nil {
		t.Fatalf("CheckLatest: %v", err)
	}
	if r.Status != StatusUpToDate {
		t.Errorf("Status = %q, want %q", r.Status, StatusUpToDate)
	}
	if r.Current != "v0.4.1" || r.Latest != "v0.4.1" {
		t.Errorf("Current/Latest = %q/%q", r.Current, r.Latest)
	}
}

func TestCheckLatest_UpdateAvailable(t *testing.T) {
	ts := fakeServer(t, http.StatusOK, `{"tag_name":"v0.4.2","html_url":"https://x/v0.4.2"}`)
	c := &Client{Endpoint: ts.URL}
	r, err := c.CheckLatest(context.Background(), "v0.4.1")
	if err != nil {
		t.Fatalf("CheckLatest: %v", err)
	}
	if r.Status != StatusUpdateAvailable {
		t.Errorf("Status = %q, want %q", r.Status, StatusUpdateAvailable)
	}
	if r.ReleaseURL != "https://x/v0.4.2" {
		t.Errorf("ReleaseURL = %q", r.ReleaseURL)
	}
}

func TestCheckLatest_Ahead(t *testing.T) {
	ts := fakeServer(t, http.StatusOK, `{"tag_name":"v0.4.1","html_url":"https://x/v0.4.1"}`)
	c := &Client{Endpoint: ts.URL}
	r, err := c.CheckLatest(context.Background(), "v0.5.0")
	if err != nil {
		t.Fatalf("CheckLatest: %v", err)
	}
	if r.Status != StatusAhead {
		t.Errorf("Status = %q, want %q", r.Status, StatusAhead)
	}
}

func TestCheckLatest_DevBuild(t *testing.T) {
	ts := fakeServer(t, http.StatusOK, `{"tag_name":"v0.4.1","html_url":"https://x/v0.4.1"}`)
	c := &Client{Endpoint: ts.URL}
	r, err := c.CheckLatest(context.Background(), "0.0.1-dev")
	if err != nil {
		t.Fatalf("CheckLatest: %v", err)
	}
	if r.Status != StatusDevBuild {
		t.Errorf("Status = %q, want %q", r.Status, StatusDevBuild)
	}
	if r.Current != "0.0.1-dev" || r.Latest != "v0.4.1" {
		t.Errorf("Current/Latest = %q/%q", r.Current, r.Latest)
	}
}

func TestCheckLatest_4xxIsError(t *testing.T) {
	ts := fakeServer(t, http.StatusNotFound, `{"message":"Not Found"}`)
	c := &Client{Endpoint: ts.URL}
	_, err := c.CheckLatest(context.Background(), "v0.4.1")
	if err == nil {
		t.Fatal("expected error on 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should name the status code: %v", err)
	}
}

func TestCheckLatest_5xxIsError(t *testing.T) {
	ts := fakeServer(t, http.StatusInternalServerError, `unavailable`)
	c := &Client{Endpoint: ts.URL}
	_, err := c.CheckLatest(context.Background(), "v0.4.1")
	if err == nil {
		t.Fatal("expected error on 500")
	}
}

func TestCheckLatest_MissingTagName(t *testing.T) {
	ts := fakeServer(t, http.StatusOK, `{"html_url":"https://x"}`)
	c := &Client{Endpoint: ts.URL}
	_, err := c.CheckLatest(context.Background(), "v0.4.1")
	if err == nil {
		t.Fatal("expected error when tag_name is missing")
	}
	if !strings.Contains(err.Error(), "tag_name") {
		t.Errorf("error should name the missing field: %v", err)
	}
}

func TestCheckLatest_MalformedJSON(t *testing.T) {
	ts := fakeServer(t, http.StatusOK, `not-json`)
	c := &Client{Endpoint: ts.URL}
	_, err := c.CheckLatest(context.Background(), "v0.4.1")
	if err == nil {
		t.Fatal("expected error on malformed JSON")
	}
}

func TestCheckLatest_NetworkFailure(t *testing.T) {
	// Point at a closed endpoint to force a Do() error.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts.Close() // close immediately so Do() errors with connection refused.
	c := &Client{Endpoint: ts.URL}
	_, err := c.CheckLatest(context.Background(), "v0.4.1")
	if err == nil {
		t.Fatal("expected network error against closed server")
	}
}

func TestCheckLatest_UnparseableCurrent(t *testing.T) {
	ts := fakeServer(t, http.StatusOK, `{"tag_name":"v0.4.1","html_url":"x"}`)
	c := &Client{Endpoint: ts.URL}
	// "garbage" has no '-' or '+' so IsDevBuild is false; ParseSemVer
	// then rejects it. A pseudo-version like "v0.4.2-0.foo" would be
	// classified as a dev build instead of a parse failure.
	_, err := c.CheckLatest(context.Background(), "garbage")
	if err == nil {
		t.Fatal("expected parse error for unparseable current")
	}
}
