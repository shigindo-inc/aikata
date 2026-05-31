package cli

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shigindo-inc/aikata/internal/install"
)

// runUpdate is a test harness that builds the update subcommand,
// captures both stdout and stderr, and returns them along with the
// terminal error (if any). The version flows through the production
// path so tests exercise the same wiring main.go uses.
func runUpdate(t *testing.T, version string, args []string) (stdout, stderr string, err error) {
	t.Helper()
	cmd := newUpdateCmd(version)
	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return outBuf.String(), errBuf.String(), err
}

func TestUpdate_BareCommandPrintsGuidance(t *testing.T) {
	out, _, err := runUpdate(t, "v0.4.1", nil)
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	for _, needle := range []string{"--check", "--apply"} {
		if !strings.Contains(out, needle) {
			t.Errorf("guidance missing %q:\n%s", needle, out)
		}
	}
}

// fakeLatestServer returns an httptest.Server that emits a
// well-formed GitHub Releases API response. The body is a literal
// string, served back to the aikata HTTP client (not a browser), so
// the no-direct-write-to-responsewriter rule does not apply here.
//
//nolint:gosec // test fixture; body is a literal, not user input
func fakeLatestServer(t *testing.T, tag string) *httptest.Server {
	t.Helper()
	body := `{"tag_name":"` + tag + `","html_url":"https://github.com/shigindo-inc/aikata/releases/tag/` + tag + `"}`
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		// nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(ts.Close)
	return ts
}

// withReleaseEndpoint is a test seam that swaps the runtime endpoint
// for an httptest.Server. The hook lives in update.go via a package
// variable so production code stays untouched.
func TestUpdate_CheckUpToDateText(t *testing.T) {
	ts := fakeLatestServer(t, "v0.4.1")
	t.Setenv("AIKATA_RELEASE_ENDPOINT_FOR_TEST", ts.URL)
	out, _, err := runUpdate(t, "v0.4.1", []string{"--check"})
	if err != nil {
		t.Fatalf("update --check: %v", err)
	}
	if !strings.Contains(out, "v0.4.1 is up to date") {
		t.Errorf("expected up-to-date message; got:\n%s", out)
	}
}

func TestUpdate_CheckUpdateAvailableText(t *testing.T) {
	ts := fakeLatestServer(t, "v0.4.2")
	t.Setenv("AIKATA_RELEASE_ENDPOINT_FOR_TEST", ts.URL)
	out, _, err := runUpdate(t, "v0.4.1", []string{"--check"})
	if err != nil {
		t.Fatalf("update --check: %v", err)
	}
	for _, needle := range []string{"v0.4.1 (current)", "v0.4.2 (latest available)", "go install", "curl -fsSL", "manual download"} {
		if !strings.Contains(out, needle) {
			t.Errorf("text output missing %q:\n%s", needle, out)
		}
	}
}

func TestUpdate_CheckDevBuildText(t *testing.T) {
	ts := fakeLatestServer(t, "v0.4.2")
	t.Setenv("AIKATA_RELEASE_ENDPOINT_FOR_TEST", ts.URL)
	out, _, err := runUpdate(t, "0.0.1-dev", []string{"--check"})
	if err != nil {
		t.Fatalf("update --check: %v", err)
	}
	if !strings.Contains(out, "development build") || !strings.Contains(out, "v0.4.2") {
		t.Errorf("dev-build text missing expected fragments:\n%s", out)
	}
}

func TestUpdate_CheckJSONEnvelope(t *testing.T) {
	ts := fakeLatestServer(t, "v0.4.2")
	t.Setenv("AIKATA_RELEASE_ENDPOINT_FOR_TEST", ts.URL)
	out, _, err := runUpdate(t, "v0.4.1", []string{"--check", "--json"})
	if err != nil {
		t.Fatalf("update --check --json: %v", err)
	}
	var got updateJSONReport
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode JSON: %v\n%s", err, out)
	}
	want := updateJSONReport{
		Version: 1, Kind: "update-check",
		Current: "v0.4.1", Latest: "v0.4.2",
		Status:     "update-available",
		ReleaseURL: "https://github.com/shigindo-inc/aikata/releases/tag/v0.4.2",
	}
	if got != want {
		t.Errorf("JSON envelope:\n got %+v\nwant %+v", got, want)
	}
}

func TestUpdate_CheckNetworkFailureIsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts.Close()
	t.Setenv("AIKATA_RELEASE_ENDPOINT_FOR_TEST", ts.URL)
	_, _, err := runUpdate(t, "v0.4.1", []string{"--check"})
	if err == nil {
		t.Fatal("expected non-nil error against a closed server")
	}
	if !strings.Contains(err.Error(), "update:") {
		t.Errorf("error should be wrapped with the subcommand name; got %v", err)
	}
}

func TestUpdate_Check5xxIsError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(ts.Close)
	t.Setenv("AIKATA_RELEASE_ENDPOINT_FOR_TEST", ts.URL)
	_, _, err := runUpdate(t, "v0.4.1", []string{"--check"})
	if err == nil {
		t.Fatal("expected error on 500 response")
	}
}

// TestApplyUpdate_ChannelRouting asserts `update --apply` returns an
// actionable, swap-free message for the channels that do not perform an
// in-place binary swap (ADR 0035 D1). The detected source is injected so
// the test does not depend on how the test binary itself was installed.
func TestApplyUpdate_ChannelRouting(t *testing.T) {
	cases := []struct {
		src     install.Source
		needles []string
	}{
		{install.SourceHomebrew, []string{"Homebrew", "brew upgrade"}},
		{install.SourceNpm, []string{"npm", "npx"}},
		{install.SourceGoInstall, []string{"go install", "@latest"}},
		{install.SourceUnknown, []string{"could not determine", "releases/latest"}},
	}
	for _, tc := range cases {
		var out bytes.Buffer
		if err := applyUpdate(&out, "v0.9.4", tc.src); err != nil {
			t.Fatalf("applyUpdate(%s): %v", tc.src, err)
		}
		for _, needle := range tc.needles {
			if !strings.Contains(out.String(), needle) {
				t.Errorf("applyUpdate(%s) output missing %q:\n%s", tc.src, needle, out.String())
			}
		}
	}
}
