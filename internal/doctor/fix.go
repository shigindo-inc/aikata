package doctor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shigindo-inc/aikata/internal/adr"
	"github.com/shigindo-inc/aikata/internal/docmap"
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
	"docmap.missing",
	"docmap.stale",
	"adr.decision-in-frontmatter",
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
		case iss.Code == "docmap.missing" || iss.Code == "docmap.stale":
			changed, err = fixDocMap(opts.TargetDir)
		case iss.Code == "adr.decision-in-frontmatter":
			changed, err = fixADRDecisionInFrontmatter(full)
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

// fixDocMap regenerates `.aikata/docmap.{yaml,md}` for targetDir. Unlike
// the per-file frontmatter fixers, the map is a derived artifact rebuilt
// wholesale, so the fix is a full regeneration rather than an in-place
// edit. The managed/external flag reuses the same ManagedIncludeGlobs as
// the freshness check so the rebuilt map matches what `aikata map`
// writes.
func fixDocMap(targetDir string) (bool, error) {
	opts, err := docmap.OptionsFor(targetDir, ManagedIncludeGlobs(targetDir))
	if err != nil {
		return false, err
	}
	if err := docmap.Generate(opts); err != nil {
		return false, err
	}
	return true, nil
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

// lifecycleStable replaces a decision word in ADR frontmatter. The
// decision itself lives on the body status line (ADR 0052).
const lifecycleStable = "stable"

// fixADRDecisionInFrontmatter moves an ADR's decision out of frontmatter:
// the frontmatter `status` becomes the lifecycle value `stable`, and an
// ADR without a body status line gets one carrying the old value. A file
// whose frontmatter and status line disagree is left alone; doctor
// reports it as adr.status-conflict for a person to settle.
func fixADRDecisionInFrontmatter(path string) (bool, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	fmValue, fmLine := adrFrontmatterStatus(body)
	fmClass, decision := adr.DecisionWord(fmValue)
	if !decision {
		return false, nil
	}
	status, hasLine := adr.ParseStatus(body)
	if hasLine && status.Class != fmClass {
		return false, nil
	}
	lines := strings.Split(string(body), "\n")
	line := lines[fmLine-1]
	cr := ""
	if strings.HasSuffix(line, "\r") {
		cr = "\r"
	}
	indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
	lines[fmLine-1] = indent + "status: " + lifecycleStable + cr
	if !hasLine {
		lines = insertStatusLine(lines, fmValue)
	}
	return true, writeFileAtomic(path, []byte(strings.Join(lines, "\n")))
}

// insertStatusLine adds `- **Status**: value` as its own paragraph after
// the first `# ` heading, or right after the frontmatter when there is no
// heading.
func insertStatusLine(lines []string, value string) []string {
	at := 0
	if len(lines) > 0 && strings.TrimRight(lines[0], "\r") == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimRight(lines[i], "\r") == "---" {
				at = i + 1
				break
			}
		}
	}
	for i := at; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "# ") {
			at = i + 1
			break
		}
	}
	block := []string{"", "- **Status**: " + value}
	if at < len(lines) && strings.TrimSpace(lines[at]) != "" {
		block = append(block, "")
	}
	out := make([]string, 0, len(lines)+len(block))
	out = append(out, lines[:at]...)
	out = append(out, block...)
	return append(out, lines[at:]...)
}

// hasFrontmatter reports whether body begins with a YAML frontmatter
// fence. docmeta.ParseFrontmatter handles full parsing; this
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
		_ = tmp.Close()
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
