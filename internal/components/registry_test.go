package components

import (
	"sort"
	"testing"
)

func TestAll_ReturnsSortedSlice(t *testing.T) {
	got := All()
	if len(got) == 0 {
		t.Fatal("expected at least one component registered")
	}
	names := make([]string, len(got))
	for i, c := range got {
		names[i] = c.Name()
	}
	if !sort.StringsAreSorted(names) {
		t.Errorf("expected names sorted ascending, got %v", names)
	}
}

func TestGet_KnownAndUnknown(t *testing.T) {
	c, ok := GetCapability("memory")
	if !ok {
		t.Fatal("expected memory to be registered")
	}
	if c.Status() != StatusActive {
		t.Errorf("memory Status = %q, want active", c.Status())
	}
	if _, ok := GetCapability("nope"); ok {
		t.Error("expected GetCapability(nope) to return false")
	}
}

func TestActiveCapabilityNames_OnlyActive(t *testing.T) {
	got := ActiveCapabilityNames()
	for _, n := range got {
		c, ok := GetCapability(n)
		if !ok {
			t.Errorf("ActiveCapabilityNames returned unknown %q", n)
			continue
		}
		if c.Status() != StatusActive {
			t.Errorf("ActiveCapabilityNames included non-active %q", n)
		}
	}
}

func TestArtifacts_KnownAndUnknown(t *testing.T) {
	c, ok := GetArtifact("adr")
	if !ok {
		t.Fatal("expected adr to be registered as an artifact")
	}
	if c.Status() != StatusActive {
		t.Errorf("adr Status = %q, want active", c.Status())
	}
	if _, ok := GetCapability("adr"); ok {
		t.Error("adr should not appear in Capabilities()")
	}
	if _, ok := GetArtifact("memory"); ok {
		t.Error("memory should not appear in Artifacts()")
	}
}
