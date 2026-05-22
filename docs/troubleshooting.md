---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-22
audience: [human, agent]
---

# Troubleshooting

Known setup problems for aikata itself. Keep entries short, concrete, and
based on real failures.

---

## `aikata: command not found` after `go install`

**Symptom**

After installing aikata with Go:

```bash
go install github.com/shigindo-inc/aikata/cmd/aikata@latest
```

the shell cannot find the command:

```bash
aikata --help
# aikata: command not found
```

**Cause**

`go install` succeeded, but the install directory is not in `PATH`.

Go writes installed binaries to `$GOBIN` when it is set. Otherwise, it
writes them to `$(go env GOPATH)/bin`, which is usually `$HOME/go/bin`.
Your shell must have that directory in `PATH` before `aikata` can be run
by name.

**Fix**

Check where Go installed the binary:

```bash
go env GOPATH GOBIN
ls "$(go env GOPATH)/bin/aikata"
```

If `GOBIN` prints a value, check that directory instead of
`$(go env GOPATH)/bin`.

Add the Go binary directory to your current shell:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
aikata --help
```

On macOS with zsh, make the change persistent:

```bash
echo 'export PATH="$(go env GOPATH)/bin:$PATH"' >> ~/.zshrc
exec zsh
```

Alternatively, install directly into a directory that is often already in
`PATH`:

```bash
GOBIN="$HOME/.local/bin" go install github.com/shigindo-inc/aikata/cmd/aikata@latest
```

Then verify:

```bash
aikata --version
# or
aikata --help
```

**Updated**: 2026-05-22
