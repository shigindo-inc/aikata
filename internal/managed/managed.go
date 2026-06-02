// Package managed provides the idempotent block-marker writer for
// project-owned generic files that aikata only contributes a small
// section of (today `.gitignore`; future targets tracked in ADR 0018).
//
// The writer never rewrites the whole file. Instead it appends or
// refreshes a clearly-marked aikata-owned block while preserving
// every user-owned byte outside the block. See
// docs/adr/0018-managed-append-for-project-owned-files.md for the
// contract and Q-INTEROP-04 for the historical leading position.
package managed

import (
	"bytes"
	"fmt"
)

// Block marker text. Once shipped these strings are effectively
// permanent — users will end up with the markers in their files. Any
// future change has to migrate every existing aikata project's
// `.gitignore`. The `>>> ... <<<` shape mirrors the well-known
// conda-init / shell-init convention so the visual signal travels.
const (
	BlockStart = "# >>> aikata managed >>>"
	BlockEnd   = "# <<< aikata managed <<<"
)

// ApplyBlock returns the new contents of a file after merging the
// aikata-managed block. Behaviour:
//
//   - If existing has a complete BlockStart…BlockEnd pair, the body
//     between the markers is replaced with newBlock (the markers
//     themselves are re-emitted from the constants above so a stale
//     marker line gets normalised).
//   - If existing has no marker pair, a fresh block (markers +
//     newBlock) is appended at EOF, separated by a single blank line
//     from any preceding non-empty content.
//   - If existing is empty, the returned bytes are just the block.
//
// The newBlock argument must NOT include the markers — callers pass
// the body they want managed, and ApplyBlock owns the framing.
//
// Errors are reserved for malformed input (e.g. a start marker
// without a matching end). Idempotency: running ApplyBlock twice
// with the same newBlock produces byte-identical output the second
// time.
func ApplyBlock(existing, newBlock []byte) ([]byte, error) {
	startIdx, endIdx, err := findMarkers(existing)
	if err != nil {
		return nil, err
	}

	framed := frameBlock(newBlock)

	if startIdx < 0 {
		// No existing block: append at EOF.
		if len(existing) == 0 {
			return framed, nil
		}
		var buf bytes.Buffer
		buf.Write(existing)
		// Ensure exactly one blank line between user content and our
		// block. Existing content might end without a newline, with
		// one newline, or with multiple — normalise to one trailing
		// newline before adding the blank-line separator.
		if !bytes.HasSuffix(existing, []byte("\n")) {
			buf.WriteByte('\n')
		}
		buf.WriteByte('\n')
		buf.Write(framed)
		return buf.Bytes(), nil
	}

	// Replace the block in place.
	var buf bytes.Buffer
	buf.Write(existing[:startIdx])
	buf.Write(framed)
	// Skip past the closing marker line (including its trailing
	// newline if present) in the source.
	tailStart := endIdx
	if tailStart < len(existing) && existing[tailStart] == '\n' {
		tailStart++
	}
	if tailStart < len(existing) {
		// Drop any blank line that immediately followed the old block
		// (so successive ApplyBlock calls don't accumulate blanks),
		// then re-emit a single newline separator if the tail has
		// non-newline content.
		for tailStart < len(existing) && existing[tailStart] == '\n' {
			tailStart++
		}
		if tailStart < len(existing) {
			buf.WriteByte('\n')
			buf.Write(existing[tailStart:])
		}
	}
	return buf.Bytes(), nil
}

// HasBlock reports whether existing contains a complete aikata
// managed block. Useful for one-shot detection without copying.
func HasBlock(existing []byte) bool {
	startIdx, endIdx, err := findMarkers(existing)
	return err == nil && startIdx >= 0 && endIdx > startIdx
}

// Frame wraps body in the aikata managed-block markers and returns the
// standalone framed block — the canonical on-disk representation a
// managed-append file always carries (the form a fresh `aikata init`
// writes, ADR 0038). Equivalent to ApplyBlock against empty existing
// content, exposed as a name so callers need not pass a nil slice to
// express "no user content yet".
func Frame(body []byte) []byte {
	return frameBlock(body)
}

// appendPaths is the set of project-owned generic files aikata manages
// through a marker block (ADR 0018), at BOTH init and sync time
// (ADR 0038). It is intentionally tiny and line-oriented. Prose files
// (CONTRIBUTING.md, SECURITY.md, …) are deliberately NOT here:
// splicing a managed block into prose was rejected (Q-INTEROP-04 (b))
// because prose merge is context-sensitive and it reintroduces the
// ownership drift ADR 0037 pushed back against. Those stay
// create-if-missing one-shot scaffolds (write discipline #4,
// ARCHITECTURE §3.4).
var appendPaths = map[string]struct{}{
	".gitignore": {},
}

// IsAppendPath reports whether rel — a slash-form path relative to the
// project root — is managed via the marker block. Shared by
// internal/scaffold (init) and internal/sync so the two stay in
// lockstep (ADR 0038).
func IsAppendPath(rel string) bool {
	_, ok := appendPaths[rel]
	return ok
}

// frameBlock wraps newBlock with the marker lines. The block is
// always terminated by a newline so consecutive ApplyBlock calls
// produce stable framing.
func frameBlock(newBlock []byte) []byte {
	var buf bytes.Buffer
	buf.WriteString(BlockStart)
	buf.WriteByte('\n')
	buf.Write(newBlock)
	if !bytes.HasSuffix(newBlock, []byte("\n")) {
		buf.WriteByte('\n')
	}
	buf.WriteString(BlockEnd)
	buf.WriteByte('\n')
	return buf.Bytes()
}

// findMarkers locates the start and end marker line positions in
// existing. Returns (startIdx, endIdx, error):
//   - startIdx points at the byte immediately before the start
//     marker line (i.e. the byte after the preceding newline, or 0
//     when the marker is at the file start).
//   - endIdx points at the byte immediately AFTER the end marker
//     line's text (before the optional trailing newline).
//   - When no markers are present, both indices are -1 and err is nil.
//   - A start without an end (or two starts) returns an error so
//     callers don't silently corrupt the file.
func findMarkers(existing []byte) (int, int, error) {
	startMarker := []byte(BlockStart)
	endMarker := []byte(BlockEnd)

	// Walk line by line so an embedded marker in user content (e.g.
	// a comment inside another tool's section) does not match. A
	// marker counts only when it is the entire trimmed line.
	startLineStart := -1
	endLineStart := -1
	i := 0
	for i < len(existing) {
		// Find the end of this line.
		j := bytes.IndexByte(existing[i:], '\n')
		var lineEnd int
		if j < 0 {
			lineEnd = len(existing)
		} else {
			lineEnd = i + j
		}
		line := bytes.TrimSpace(existing[i:lineEnd])
		switch {
		case bytes.Equal(line, startMarker):
			if startLineStart >= 0 {
				return -1, -1, fmt.Errorf("managed: duplicate start marker at byte %d (first at %d)", i, startLineStart)
			}
			startLineStart = i
		case bytes.Equal(line, endMarker):
			if startLineStart < 0 {
				return -1, -1, fmt.Errorf("managed: end marker at byte %d without a matching start", i)
			}
			endLineStart = lineEnd
		}
		if endLineStart >= 0 {
			return startLineStart, endLineStart, nil
		}
		if j < 0 {
			break
		}
		i = lineEnd + 1
	}
	if startLineStart >= 0 && endLineStart < 0 {
		return -1, -1, fmt.Errorf("managed: start marker at byte %d without a matching end", startLineStart)
	}
	return -1, -1, nil
}
