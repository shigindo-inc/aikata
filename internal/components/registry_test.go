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
	c, ok := Get("memory")
	if !ok {
		t.Fatal("expected memory to be registered")
	}
	if c.Status() != StatusActive {
		t.Errorf("memory Status = %q, want active", c.Status())
	}
	if _, ok := Get("nope"); ok {
		t.Error("expected Get(nope) to return false")
	}
}

func TestActiveNames_OnlyActive(t *testing.T) {
	got := ActiveNames()
	for _, n := range got {
		c, ok := Get(n)
		if !ok {
			t.Errorf("ActiveNames returned unknown %q", n)
			continue
		}
		if c.Status() != StatusActive {
			t.Errorf("ActiveNames included non-active %q", n)
		}
	}
}
