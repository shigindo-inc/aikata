package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FixableCodes lists the Issue.Code values that Fix knows how to
// repair. Other codes (and issues with no code at all) are skipped.
// Keep this list in lockstep with the dispatch switch in Fix.
var FixableCodes = []string{
	"frontmatter.missing",
	"frontmatter.missing-key.project",
	"frontmatter.missing-key.status",
	"frontmatter.missing-key.version",
	"frontmatter.missing-key.updated",
	"frontmatter.missing-key.audience",
	"updated.stale",
}

// FixResult summarizes the outcome of a Fix pass.
type FixResult struct {
	// Fixed counts issues that were repaired.
	Fixed int
	// Skipped counts issues whose Code has no registered fixer.
	Skipped int
	// Files lists the distinct files whose contents changed, slash-
	// separated paths relative to Options.TargetDir.
	Files []string
}

// Fix repairs the trivially-fixable subset of issues in place. The
// caller is responsible for invoking Run beforehand to obtain the
// issue slice. Fix preserves issue order and stops at the first I/O
// error — earlier fixes remain on disk so the user can rerun with the
// underlying problem resolved.
//
// The fixable subset is intentionally narrow:
//   - "frontmatter.missing"          → write a default 5-key block.
//   - "frontmatter.missing-key.<k>"  → append the missing key with
//     a placeholder value.
//   - "updated.stale"                → bump `updated:` to today.
//
// Today is taken from opts.Now to keep tests deterministic; the CLI
// layer passes a zero Time so doctor's Run substitutes time.Now.
func Fix(opts Options, issues []Issue) (FixResult, error) {
	if opts.TargetDir == "" {
		return FixResult{}, fmt.Errorf("doctor: fix: target directory is required")
	}
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	today := now.Format("2006-01-02")

	var res FixResult
	touched := map[string]struct{}{}
	for _, iss := range issues {
		full := filepath.Join(opts.TargetDir, filepath.FromSlash(iss.File))
		var changed bool
		var err error
		switch {
		case iss.Code == "frontmatter.missing":
			changed, err = fixMissingFrontmatter(full, today)
		case strings.HasPrefix(iss.Code, "frontmatter.missing-key."):
			key := strings.TrimPrefix(iss.Code, "frontmatter.missing-key.")
			changed, err = fixMissingKey(full, key, today)
		case iss.Code == "updated.stale":
			changed, err = fixStaleUpdated(full, today)
		default:
			res.Skipped++
			continue
		}
		if err != nil {
			return res, fmt.Errorf("doctor: fix %s: %w", iss.File, err)
		}
		if changed {
			res.Fixed++
			touched[iss.File] = struct{}{}
		}
	}
	for f := range touched {
		res.Files = append(res.Files, f)
	}
	return res, nil
}

// defaultFrontmatter renders the minimal frontmatter block written
// when an entire block is missing. Values are deliberately
// placeholders rather than guesses — doctor surfaces them later so a
// human can replace them.
func defaultFrontmatter(today string) string {
	return strings.Join([]string{
		"---",
		"project: TODO",
		"status: draft",
		"version: 0.0.0",
		"updated: " + today,
		"audience: [human, agent]",
		"---",
		"",
	}, "\n") + "\n"
}

// placeholderForKey returns the default value Fix injects for a
// missing-but-required frontmatter key. The placeholders match
// defaultFrontmatter above so a file produced by either path is
// indistinguishable.
func placeholderForKey(key, today string) string {
	switch key {
	case "project":
		return "TODO"
	case "status":
		return "draft"
	case "version":
		return "0.0.0"
	case "updated":
		return today
	case "audience":
		return "[human, agent]"
	}
	return "TODO"
}

func fixMissingFrontmatter(path, today string) (bool, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	// Idempotent: if a frontmatter block already exists (the issue
	// may have been resolved between Run and Fix), leave it alone.
	if hasFrontmatter(body) {
		return false, nil
	}
	out := defaultFrontmatter(today) + string(body)
	return true, writeFileAtomic(path, []byte(out))
}

func fixMissingKey(path, key, today string) (bool, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	if !hasFrontmatter(body) {
		// Without a block to insert into, the missing-block fixer
		// handles this case; refuse to invent one here.
		return false, nil
	}
	text := string(body)
	// Locate the closing `---\n` of the frontmatter block.
	openLen := frontmatterOpenLen(text)
	if openLen == 0 {
		return false, nil
	}
	rest := text[openLen:]
	closeIdx := strings.Index(rest, "\n---\n")
	if closeIdx < 0 {
		closeIdx = strings.Index(rest, "\n---\r\n")
		if closeIdx < 0 {
			return false, nil
		}
	}
	absClose := openLen + closeIdx + 1 // points at `---` of the closer
	// Idempotent: re-running Fix on a now-clean file is a no-op.
	block := text[:absClose]
	if blockHasKey(block, key) {
		return false, nil
	}
	insert := key + ": " + placeholderForKey(key, today) + "\n"
	out := text[:absClose] + insert + text[absClose:]
	return true, writeFileAtomic(path, []byte(out))
}

func fixStaleUpdated(path, today string) (bool, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	if !hasFrontmatter(body) {
		return false, nil
	}
	text := string(body)
	openLen := frontmatterOpenLen(text)
	rest := text[openLen:]
	closeIdx := strings.Index(rest, "\n---\n")
	if closeIdx < 0 {
		closeIdx = strings.Index(rest, "\n---\r\n")
		if closeIdx < 0 {
			return false, nil
		}
	}
	absClose := openLen + closeIdx + 1
	block := text[:absClose]
	newBlock, replaced := replaceFrontmatterValue(block, "updated", today)
	if !replaced {
		return false, nil
	}
	out := newBlock + text[absClose:]
	return out != text, writeFileAtomic(path, []byte(out))
}

// hasFrontmatter reports whether body begins with a YAML frontmatter
// fence. parseFrontmatter (in checks.go) handles parsing; this
// lightweight check exists so the fixers don't depend on its return
// type.
func hasFrontmatter(body []byte) bool {
	text := string(body)
	return strings.HasPrefix(text, "---\n") || strings.HasPrefix(text, "---\r\n")
}

// frontmatterOpenLen returns the byte length of the opening fence
// (including its trailing newline), or 0 if no fence is present.
func frontmatterOpenLen(text string) int {
	if strings.HasPrefix(text, "---\n") {
		return len("---\n")
	}
	if strings.HasPrefix(text, "---\r\n") {
		return len("---\r\n")
	}
	return 0
}

// blockHasKey scans the frontmatter block for a `key:` declaration at
// the start of any line. The check ignores list-style nested entries.
func blockHasKey(block, key string) bool {
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, key+":") {
			return true
		}
	}
	return false
}

// replaceFrontmatterValue substitutes the value of `key:` inside the
// supplied frontmatter block. Returns the rewritten block and a flag
// indicating whether a substitution actually happened.
func replaceFrontmatterValue(block, key, value string) (string, bool) {
	lines := strings.Split(block, "\n")
	prefix := key + ":"
	for i, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if strings.HasPrefix(trimmed, prefix) {
			indentLen := len(line) - len(trimmed)
			lines[i] = line[:indentLen] + key + ": " + value
			return strings.Join(lines, "\n"), true
		}
	}
	return block, false
}

// writeFileAtomic writes b to path via a temporary file + rename so
// the destination never sees a partial write. Permission bits mirror
// the existing file or default to 0644 when the target does not yet
// exist.
func writeFileAtomic(path string, b []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".aikata-doctor-fix-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() {
		// If rename succeeded the file no longer exists; ignore that.
		_ = os.Remove(tmpPath)
	}()
	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if stat, err := os.Stat(path); err == nil {
		mode = stat.Mode().Perm()
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
