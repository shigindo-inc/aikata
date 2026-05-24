// Package release queries GitHub Releases for the aikata CLI's latest
// version and computes whether the running binary is up-to-date. The
// package owns the HTTP boundary and the minimal semver comparison
// aikata's release stream needs; `aikata update --check` is the first
// caller. Future self-update work (v0.6) will reuse the fetcher and
// the comparator.
package release

import (
	"fmt"
	"strconv"
	"strings"
)

// devSentinel matches the value of cmd/aikata/main.devVersion. It is
// duplicated here (instead of imported) because internal/release is
// reused by other tools (tests, future self-update) that should not
// pull in the main package.
const devSentinel = "0.0.1-dev"

// SemVer is the parsed form of a `vX.Y.Z` (or `X.Y.Z`) version string.
// aikata does not ship pre-release or build-metadata identifiers yet,
// so the parser intentionally rejects anything that does not match
// the three-segment shape.
type SemVer struct {
	Major, Minor, Patch int
}

// ParseSemVer parses `vX.Y.Z` or `X.Y.Z`. Returns an error for the dev
// sentinel and for any input that does not contain exactly three
// non-negative integer segments. Callers that want to distinguish the
// dev build from a parse failure should check IsDevBuild first.
func ParseSemVer(s string) (SemVer, error) {
	if s == "" {
		return SemVer{}, fmt.Errorf("release: empty version string")
	}
	if IsDevBuild(s) {
		return SemVer{}, fmt.Errorf("release: %q is the dev sentinel, not a release version", s)
	}
	trimmed := strings.TrimPrefix(s, "v")
	parts := strings.Split(trimmed, ".")
	if len(parts) != 3 {
		return SemVer{}, fmt.Errorf("release: %q is not a three-segment semver", s)
	}
	out := SemVer{}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return SemVer{}, fmt.Errorf("release: %q segment %d (%q) is not a non-negative integer", s, i, p)
		}
		if n < 0 {
			return SemVer{}, fmt.Errorf("release: %q segment %d (%d) is negative", s, i, n)
		}
		switch i {
		case 0:
			out.Major = n
		case 1:
			out.Minor = n
		case 2:
			out.Patch = n
		}
	}
	return out, nil
}

// Compare returns -1 / 0 / +1 comparing v against other. Segment
// ordering is major > minor > patch, all integer-compared.
func (v SemVer) Compare(other SemVer) int {
	for _, pair := range [3][2]int{
		{v.Major, other.Major},
		{v.Minor, other.Minor},
		{v.Patch, other.Patch},
	} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	return 0
}

// IsDevBuild reports whether s should be treated as a development
// build for release-check purposes. The cases covered:
//
//   - The dev sentinel ("0.0.1-dev") embedded when no ldflags / module
//     version are available.
//   - A Go module pseudo-version
//     ("vX.Y.Z-0.YYYYMMDDHHMMSS-COMMIT[+dirty]") generated when the
//     binary is built from a local checkout instead of a tagged
//     release. The leading hyphen or trailing "+dirty" gives them
//     away.
//
// aikata does not ship semver pre-releases (v0.5.0-beta.1) today, so
// any "-" or "+" after the version core is interpreted as a non-tag
// build. When pre-releases become real, this check should learn the
// distinction; for now we trade fidelity for a clean dev-time UX.
func IsDevBuild(s string) bool {
	if s == devSentinel {
		return true
	}
	return strings.ContainsAny(s, "-+")
}
