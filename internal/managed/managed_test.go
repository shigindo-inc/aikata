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

// TestFrame_EqualsApplyBlockOnEmpty pins that Frame is the standalone
// framed representation — identical to ApplyBlock against no existing
// content (ADR 0038), and that the result re-frames idempotently.
func TestFrame_EqualsApplyBlockOnEmpty(t *testing.T) {
	body := []byte("alpha\nbeta\n")
	framed := Frame(body)

	viaApply, err := ApplyBlock(nil, body)
	if err != nil {
		t.Fatalf("ApplyBlock: %v", err)
	}
	if !bytes.Equal(framed, viaApply) {
		t.Errorf("Frame(body) = %q, want = ApplyBlock(nil, body) = %q", framed, viaApply)
	}
	if !HasBlock(framed) {
		t.Errorf("Frame output should carry a complete block:\n%s", framed)
	}
	// Re-applying the same body to the framed form is a no-op.
	again, err := ApplyBlock(framed, body)
	if err != nil {
		t.Fatalf("ApplyBlock(framed): %v", err)
	}
	if !bytes.Equal(again, framed) {
		t.Errorf("ApplyBlock on framed block should be idempotent; got:\n%s", again)
	}
}

// TestIsAppendPath_OnlyGitignore is the Q-INTEROP-04 (b) guard: the
// managed-append set must contain ONLY line-oriented project-owned
// files (today just `.gitignore`) and never prose targets. Splicing a
// managed block into prose (CONTRIBUTING.md / SECURITY.md / …) was
// rejected — if someone adds one here, this test must fail and force a
// decision review rather than silently shipping prose mutation.
func TestIsAppendPath_OnlyGitignore(t *testing.T) {
	if !IsAppendPath(".gitignore") {
		t.Errorf("IsAppendPath(.gitignore) = false, want true")
	}
	prose := []string{
		"CONTRIBUTING.md", "SECURITY.md", "README.md", "AGENTS.md",
		"CLAUDE.md", "docs/adr/0001-foo.md", "SPEC.md",
	}
	for _, p := range prose {
		if IsAppendPath(p) {
			t.Errorf("IsAppendPath(%q) = true; prose must never be managed-append (Q-INTEROP-04 (b))", p)
		}
	}
	// Exhaustive: the set is exactly {.gitignore}. If this count grows,
	// the new entry must be a line-oriented file and reviewed against
	// ADR 0037 ownership boundaries.
	if got := len(appendPaths); got != 1 {
		t.Errorf("len(appendPaths) = %d, want 1 (only .gitignore)", got)
	}
}
