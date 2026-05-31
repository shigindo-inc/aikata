package release

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// defaultEndpoint is the GitHub Releases API URL for the aikata
// repository. The unauthenticated rate limit (60 req/h per IP) is more
// than enough for a human-invoked `aikata update --check` and avoids
// the operational overhead of issuing tokens.
const defaultEndpoint = "https://api.github.com/repos/shigindo-inc/aikata/releases/latest"

// defaultTimeout caps the HTTP request so `update --check` cannot hang
// a user's terminal indefinitely on a flaky network.
const defaultTimeout = 10 * time.Second

// Status enumerates the outcomes of a release check. The string values
// double as the JSON envelope's `status` field; treat the constants as
// the source of truth for that schema.
type Status string

const (
	// StatusUpToDate means the running binary's version equals the
	// latest published release.
	StatusUpToDate Status = "up-to-date"
	// StatusUpdateAvailable means the latest release is strictly
	// newer than the running binary.
	StatusUpdateAvailable Status = "update-available"
	// StatusDevBuild means the running binary is a development build
	// (no comparable semver). Latest is still reported informationally.
	StatusDevBuild Status = "dev-build"
	// StatusAhead means the running binary is strictly newer than the
	// latest published release (e.g. a freshly tagged build that
	// hasn't reached the API cache yet). Treated as "no action".
	StatusAhead Status = "ahead-of-latest"
)

// Result is the structured outcome of a release check. It is the DTO
// shared between the cli text formatter and the --json encoder so the
// two output modes cannot drift apart.
type Result struct {
	Current    string `json:"current"`
	Latest     string `json:"latest"`
	Status     Status `json:"status"`
	ReleaseURL string `json:"release_url"`
}

// Client fetches the latest aikata release from GitHub. Both fields
// are optional: zero values fall back to a default 10s-timeout HTTP
// client and the canonical GitHub Releases endpoint. Tests inject a
// custom Endpoint pointed at httptest.NewServer.
type Client struct {
	HTTPClient *http.Client
	Endpoint   string
	// DownloadBase overrides the release-download root used by
	// DownloadReleaseAsset (self-update). Empty falls back to the
	// canonical GitHub Releases download URL. Tests point both Endpoint
	// and DownloadBase at one httptest.Server.
	DownloadBase string
	UserAgent    string
}

// CheckLatest fetches the latest release and compares it against
// current. Returns a Result with Status set; the error channel is
// reserved for network / parse failures that the caller should
// surface (no silent degradation — `update --check` is opt-in, so a
// connectivity issue should be visible).
//
// current = "0.0.1-dev" yields StatusDevBuild without error.
func (c *Client) CheckLatest(ctx context.Context, current string) (Result, error) {
	endpoint := c.Endpoint
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultTimeout}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Result{}, fmt.Errorf("release: build request: %w", err)
	}
	// GitHub's recommended Accept header — version-pin the API surface.
	req.Header.Set("Accept", "application/vnd.github+json")
	userAgent := c.UserAgent
	if userAgent == "" {
		userAgent = "aikata-cli"
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("release: GET %s: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Read a bounded prefix so a runaway 5xx body cannot blow up
		// memory or hide the underlying status.
		preview, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return Result{}, fmt.Errorf("release: %s returned %d: %s",
			endpoint, resp.StatusCode, string(preview))
	}

	var payload struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&payload); err != nil {
		return Result{}, fmt.Errorf("release: decode response: %w", err)
	}
	if payload.TagName == "" {
		return Result{}, fmt.Errorf("release: response missing tag_name")
	}

	result := Result{
		Current:    current,
		Latest:     payload.TagName,
		ReleaseURL: payload.HTMLURL,
	}

	if IsDevBuild(current) {
		result.Status = StatusDevBuild
		return result, nil
	}

	cur, err := ParseSemVer(current)
	if err != nil {
		return result, fmt.Errorf("release: parse current %q: %w", current, err)
	}
	latest, err := ParseSemVer(payload.TagName)
	if err != nil {
		return result, fmt.Errorf("release: parse latest %q: %w", payload.TagName, err)
	}
	switch cur.Compare(latest) {
	case 0:
		result.Status = StatusUpToDate
	case -1:
		result.Status = StatusUpdateAvailable
	case 1:
		result.Status = StatusAhead
	}
	return result, nil
}
