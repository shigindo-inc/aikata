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
