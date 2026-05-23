---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-23
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

**Updated**: 2026-05-23

---

## `aikata --version` shows `0.0.1-dev` after `go install`

**Symptom**

After installing a tagged release with Go:

```bash
go install github.com/shigindo-inc/aikata/cmd/aikata@latest
aikata --version
```

older aikata binaries may print:

```bash
aikata version 0.0.1-dev
```

**Cause**

Release archives built by GoReleaser embed the release version with
`-ldflags`. Plain `go install` does not use those release-time flags.

Starting after v0.1.0, aikata falls back to Go build info when the
linked version is still the development default. That means installs
such as `@latest` and `@v0.1.0` can report the resolved module version
when Go recorded one.

**Fix**

Check the module version embedded in the installed binary:

```bash
go version -m "$(command -v aikata)"
```

Look for the `mod` line:

```text
mod     github.com/shigindo-inc/aikata  v0.1.0
```

If the module version is current, the binary itself is current even if
an older `aikata --version` output says `0.0.1-dev`.

**Updated**: 2026-05-23
