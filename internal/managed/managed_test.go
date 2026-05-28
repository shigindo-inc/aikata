package managed

import (
	"bytes"
	"strings"
	"testing"
)

// TestApplyBlock_EmptyFile asserts the simplest case: writing into
// an empty (or non-existent) file produces just the framed block.
func TestApplyBlock_EmptyFile(t *testing.T) {
	got, err := ApplyBlock(nil, []byte("alpha\nbeta\n"))
	if err != nil {
		t.Fatalf("ApplyBlock: %v", err)
	}
	want := BlockStart + "\nalpha\nbeta\n" + BlockEnd + "\n"
	if string(got) != want {
		t.Errorf("ApplyBlock(empty) = %q, want %q", got, want)
	}
}

// TestApplyBlock_AppendsToExistingFile verifies that a file without
// any markers grows by exactly one framed block at EOF, separated
// by a blank line, and the original content is untouched.
func TestApplyBlock_AppendsToExistingFile(t *testing.T) {
	existing := []byte("user.local\n.cache/\n")
	got, err := ApplyBlock(existing, []byte("CLAUDE.md\n"))
	if err != nil {
		t.Fatalf("ApplyBlock: %v", err)
	}
	want := "user.local\n.cache/\n\n" + BlockStart + "\nCLAUDE.md\n" + BlockEnd + "\n"
	if string(got) != want {
		t.Errorf("ApplyBlock append:\n got: %q\nwant: %q", got, want)
	}
}

// TestApplyBlock_ReplacesExistingBlock verifies the refresh path:
// when the existing file already has an aikata block, only the
// content between markers is replaced; bytes outside survive.
func TestApplyBlock_ReplacesExistingBlock(t *testing.T) {
	existing := []byte("# user\nuser.local\n\n" + BlockStart + "\nold.md\n" + BlockEnd + "\n\n# footer\nfoot.local\n")
	got, err := ApplyBlock(existing, []byte("new.md\n"))
	if err != nil {
		t.Fatalf("ApplyBlock: %v", err)
	}
	if !strings.Contains(string(got), "new.md") {
		t.Errorf("ApplyBlock should have replaced block; got:\n%s", got)
	}
	if strings.Contains(string(got), "old.md") {
		t.Errorf("ApplyBlock should have removed old block content; got:\n%s", got)
	}
	if !strings.Contains(string(got), "user.local") {
		t.Errorf("user content above the block must be preserved; got:\n%s", got)
	}
	if !strings.Contains(string(got), "foot.local") {
		t.Errorf("user content below the block must be preserved; got:\n%s", got)
	}
}

// TestApplyBlock_IsIdempotent asserts the no-drift property: running
// ApplyBlock twice with the same new block produces byte-identical
// output the second time.
func TestApplyBlock_IsIdempotent(t *testing.T) {
	existing := []byte("user.local\n")
	once, err := ApplyBlock(existing, []byte("CLAUDE.md\n"))
	if err != nil {
		t.Fatalf("first ApplyBlock: %v", err)
	}
	twice, err := ApplyBlock(once, []byte("CLAUDE.md\n"))
	if err != nil {
		t.Fatalf("second ApplyBlock: %v", err)
	}
	if !bytes.Equal(once, twice) {
		t.Errorf("ApplyBlock not idempotent:\nonce:\n%s\ntwice:\n%s", once, twice)
	}
}

// TestApplyBlock_NoTrailingNewlineOnExisting handles the edge case
// where the user's file lacks a final newline. The writer must add
// one before the separator + block.
func TestApplyBlock_NoTrailingNewlineOnExisting(t *testing.T) {
	existing := []byte("user.local")
	got, err := ApplyBlock(existing, []byte("CLAUDE.md\n"))
	if err != nil {
		t.Fatalf("ApplyBlock: %v", err)
	}
	want := "user.local\n\n" + BlockStart + "\nCLAUDE.md\n" + BlockEnd + "\n"
	if string(got) != want {
		t.Errorf("ApplyBlock without trailing newline:\n got: %q\nwant: %q", got, want)
	}
}

// TestApplyBlock_DuplicateStartMarkerIsError protects against a
// malformed file where someone has hand-inserted a stray marker.
// Silent corruption would be worse than refusing the write.
func TestApplyBlock_DuplicateStartMarkerIsError(t *testing.T) {
	existing := []byte(BlockStart + "\nfoo\n" + BlockStart + "\nbar\n" + BlockEnd + "\n")
	_, err := ApplyBlock(existing, []byte("new\n"))
	if err == nil {
		t.Fatalf("expected error for duplicate start marker")
	}
}

// TestApplyBlock_StartWithoutEndIsError similarly refuses files
// where a block was opened but never closed.
func TestApplyBlock_StartWithoutEndIsError(t *testing.T) {
	existing := []byte(BlockStart + "\nfoo\n")
	_, err := ApplyBlock(existing, []byte("new\n"))
	if err == nil {
		t.Fatalf("expected error for start without end")
	}
}

// TestApplyBlock_EndWithoutStartIsError catches the inverse.
func TestApplyBlock_EndWithoutStartIsError(t *testing.T) {
	existing := []byte("user\n" + BlockEnd + "\n")
	_, err := ApplyBlock(existing, []byte("new\n"))
	if err == nil {
		t.Fatalf("expected error for end without start")
	}
}

// TestHasBlock probes the detection helper independently of
// ApplyBlock.
func TestHasBlock(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		want bool
	}{
		{"empty", nil, false},
		{"plain", []byte("foo\nbar\n"), false},
		{"start only", []byte(BlockStart + "\nfoo\n"), false},
		{"end only", []byte("foo\n" + BlockEnd + "\n"), false},
		{"complete", []byte(BlockStart + "\nfoo\n" + BlockEnd + "\n"), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := HasBlock(tc.in); got != tc.want {
				t.Errorf("HasBlock(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

// TestApplyBlock_MarkersInUserContentDoNotMatch makes sure the
// markers are matched against full lines only, so a stray substring
// inside a user comment does not trigger detection.
func TestApplyBlock_MarkersInUserContentDoNotMatch(t *testing.T) {
	// User has the start marker embedded inside another sentence.
	existing := []byte("# example: " + BlockStart + " inline\nuser.local\n")
	got, err := ApplyBlock(existing, []byte("new\n"))
	if err != nil {
		t.Fatalf("ApplyBlock: %v", err)
	}
	// Should have appended (since the inline mention isn't a real
	// marker line) — the result still contains the original
	// example: prefix.
	if !strings.Contains(string(got), "# example:") {
		t.Errorf("inline marker mention should be preserved; got:\n%s", got)
	}
}
