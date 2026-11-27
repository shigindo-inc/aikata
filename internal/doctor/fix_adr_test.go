package doctor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFix_ADRDecisionInFrontmatter(t *testing.T) {
	head := func(status string) string {
		return "---\nproject: sample\nstatus: " + status + "\nversion: 0.0.1\nupdated: 2026-05-20\naudience: [human, agent]\n---\n\n# ADR 0002 — Case\n"
	}
	cases := []struct {
		name, body, want string
	}{
		{
			name: "status line agrees: only frontmatter changes",
			body: head("accepted") + "\n> **状態**: active（2026-10-06）\n",
			want: head("stable") + "\n> **状態**: active（2026-10-06）\n",
		},
		{
			name: "no status line: the frontmatter word moves into the body",
			body: head("superseded") + "\nText.\n",
			want: head("stable") + "\n- **Status**: superseded\n\nText.\n",
		},
		{
			name: "status line disagrees: file is left for a person",
			body: head("superseded") + "\n> **状態**: active\n",
			want: head("superseded") + "\n> **状態**: active\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			path := filepath.Join(tmp, "docs", "adr", "0002-case.md")
			writeMD(t, path, tc.body)
			_, err := Fix(Options{TargetDir: tmp, Now: fixedDate()}, []Issue{
				{Level: LevelWarning, File: "docs/adr/0002-case.md", Code: "adr.decision-in-frontmatter"},
			})
			if err != nil {
				t.Fatalf("Fix: %v", err)
			}
			got, _ := os.ReadFile(path)
			if string(got) != tc.want {
				t.Errorf("file contents:\n--- got ---\n%s\n--- want ---\n%s", got, tc.want)
			}
			if string(got) != tc.body {
				if issues := adrStatusIssues("docs/adr/0002-case.md", got); len(issues) != 0 {
					t.Errorf("issues after fix = %+v, want none", issues)
				}
			}
		})
	}
}
