// Package repolint holds test-only guards that lint the aikata repository
// itself rather than any user project. It ships no runtime code and is not
// imported by the binary; its sole purpose is to fail CI when the tracked
// tree picks up content it should never contain.
//
// The secret / privacy scan lives here (see secretscan_test.go). It is the
// v0.8.0 "secret-scan CI gate" from ADR 0022, realised as a Go test so it
// runs inside the existing `go test ./...` matrix on every supported OS
// without a separate workflow or shell-escaping fragility.
package repolint
