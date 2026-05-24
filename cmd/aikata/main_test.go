package main

import (
	"runtime/debug"
	"testing"
)

func TestResolveVersion_LinkedVersionWins(t *testing.T) {
	got := resolveVersion("v1.2.3", &debug.BuildInfo{
		Main: debug.Module{Version: "v9.9.9"},
	}, true)
	if got != "v1.2.3" {
		t.Fatalf("resolveVersion() = %q, want %q", got, "v1.2.3")
	}
}

func TestResolveVersion_UsesModuleVersionForGoInstall(t *testing.T) {
	got := resolveVersion(devVersion, &debug.BuildInfo{
		Main: debug.Module{Version: "v0.1.0"},
	}, true)
	if got != "v0.1.0" {
		t.Fatalf("resolveVersion() = %q, want %q", got, "v0.1.0")
	}
}

func TestResolveVersion_KeepsDevVersionForLocalBuild(t *testing.T) {
	got := resolveVersion(devVersion, &debug.BuildInfo{
		Main: debug.Module{Version: "(devel)"},
	}, true)
	if got != devVersion {
		t.Fatalf("resolveVersion() = %q, want %q", got, devVersion)
	}
}

func TestResolveVersion_NormalizesLdflagsWithoutVPrefix(t *testing.T) {
	// GoReleaser <= v0.2.1 injected the bare semver string. New
	// releases inject the v-prefixed tag, but the normalization
	// path must still cover bare strings so that custom builds
	// (e.g. `go build -ldflags "-X main.version=0.3.0"`) produce
	// the same `aikata --version` output as `go install`.
	got := resolveVersion("0.3.0", nil, false)
	if got != "v0.3.0" {
		t.Fatalf("resolveVersion(\"0.3.0\") = %q, want %q", got, "v0.3.0")
	}
}

func TestResolveVersion_NormalizesBuildInfoWithoutVPrefix(t *testing.T) {
	got := resolveVersion(devVersion, &debug.BuildInfo{
		Main: debug.Module{Version: "0.2.1"},
	}, true)
	if got != "v0.2.1" {
		t.Fatalf("resolveVersion() = %q, want %q", got, "v0.2.1")
	}
}

func TestResolveVersion_PreservesPrereleaseSuffix(t *testing.T) {
	got := resolveVersion("0.3.0-rc1", nil, false)
	if got != "v0.3.0-rc1" {
		t.Fatalf("resolveVersion(\"0.3.0-rc1\") = %q, want %q", got, "v0.3.0-rc1")
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{devVersion, devVersion},
		{"v0.3.0", "v0.3.0"},
		{"0.3.0", "v0.3.0"},
		{"0.3.0-rc1", "v0.3.0-rc1"},
	}
	for _, tt := range tests {
		if got := normalizeVersion(tt.in); got != tt.want {
			t.Errorf("normalizeVersion(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}
