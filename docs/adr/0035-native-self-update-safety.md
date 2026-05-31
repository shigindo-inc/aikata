---
project: aikata
status: draft
version: 0.0.1
updated: 2026-06-01
audience: [human, agent]
---

# ADR 0035 - Native self-update (`aikata update --apply`) safety model

- **Status**: Accepted
- **Date**: 2026-06-01
- **Deciders**: aikata maintainers
- **Related**: ADR 0009 (`update` vs `sync` split), ADR 0023 (release
  signing & supply-chain hardening), ADR 0032 (channel-publication
  split — D2 schedules native self-update at v0.9.4), `SECURITY.md`
  (Agent Safety: no new remote-code-execution behaviour without an
  ADR + review)

## Context

[ADR 0032 D2](./0032-split-channel-publication-by-distribution-value.md)
scheduled `aikata update --apply` for v0.9.4, covering the channels that
exist today (`install-script`, `go-install`, `github-release`) and
stubbing the not-yet-real `homebrew` / `npm` branches. The foundation
shipped in v0.6.0: `internal/install.Detect()` identifies the install
channel, and `scripts/install.sh` writes the `aikata.install-source`
marker.

`update --apply` downloads a binary from the network and replaces the
running executable. That is exactly the kind of remote-code-execution
surface `SECURITY.md` requires to pass an ADR + review before shipping.
This ADR records the safety model so the review is on the record.

## Decision

### D1 - Per-channel action (no one-size-fits-all swap)

`update --apply` picks the *safe path for how aikata was installed*,
read from `internal/install.Detect("")`:

- **`install-script`, `github-release`** — prebuilt-binary channels with
  the binary at a user-writable path. `--apply` performs an in-place
  **binary swap** (download → verify → extract → atomic replace). This is
  the headline feature; the `curl … | sh` audience is aikata's main
  no-Go install path.
- **`go-install`** — the user owns the Go toolchain, so the safe,
  channel-native upgrade is `go install
  github.com/shigindo-inc/aikata/cmd/aikata@latest`. `--apply` prints that
  command rather than swapping a GoReleaser binary over a toolchain-built
  one (which would be surprising and version-stamp-inconsistent). This is
  still "covered" per ADR 0032 D2 — covered means the right action, not
  necessarily a swap.
- **`homebrew`, `npm`** — not real yet (deferred to v0.9.9). `--apply`
  prints an actionable "use your package manager" stub.
- **`unknown`** — `--apply` prints manual-download guidance (the release
  URL) rather than guessing.

### D2 - Verify before swap; SHA-256 over an HTTPS transport anchor

The swap path mirrors `scripts/install.sh` exactly:

1. Resolve the host asset name `aikata_<version>_<GOOS>_<GOARCH>.tar.gz`
   (`runtime.GOOS` / `runtime.GOARCH` already equal GoReleaser's
   `{{.Os}}`/`{{.Arch}}`; no uname-style mapping is needed).
2. Download the **archive** and `checksums.txt` from the release.
3. **Verify the archive's SHA-256 against its `checksums.txt` entry, and
   only then extract.** `checksums.txt` lists archive hashes, not the
   inner binary, so verification happens on the archive (identical to
   install.sh's `verify_sha256` running before `tar -xzf`).
4. Extract the tar entry named exactly `aikata` (the archive also bundles
   LICENSE / README / CHANGELOG).
5. Atomically replace the running binary (D3).

**Trust model**: `checksums.txt` is fetched over the same HTTPS channel
as the archive, so the SHA-256 check guards against corruption / partial
downloads, with **HTTPS-to-github.com as the transport trust anchor**.
Cosign verification of `checksums.txt.sigstore.json` (ADR 0023) is a
**documented non-goal for v0.9.4** — install.sh does not do it either,
and adding a cosign dependency to the binary is out of scope here.
Revisit if a threat model beyond transport integrity is required.

### D3 - POSIX atomic replace; Windows and permission errors are honest

- **Replace** writes the verified new binary to a temp file in the same
  directory as the target, `chmod 0755`, then `os.Rename` over the
  target. On POSIX, rename is atomic and replacing a running binary is
  allowed (the live process keeps the old inode), so no backup copy and
  no window where the binary is missing.
- **Windows** cannot overwrite a running `.exe`. `--apply` detects
  `runtime.GOOS == "windows"` **before any download** and returns an
  actionable manual-download / `go install` message, consistent with
  install.sh's manual-Windows stance.
- **Permission denied** (e.g. a root-owned `/usr/local/bin`) is caught
  specifically and returned as an actionable message (re-run with
  sufficient permissions, or reinstall via the install script). No temp
  file is left behind.

### D4 - Apply-time gates before any download

`--apply` reuses `release.Client.CheckLatest` and only proceeds to
download on `StatusUpdateAvailable`:

- **dev-build** → refuse ("self-update applies to released binaries").
- **up-to-date / ahead-of-latest** → no-op with a clear message.
- **update-available** → run the swap (or the channel-native action).

The `aikata.install-source` sibling marker is left untouched after a
swap — it still accurately records the channel that placed the binary.

## Consequences

- `curl … | sh` and direct GitHub-Release users get one-command
  self-update; Go users get the correct toolchain command; package-manager
  and Windows users get honest, actionable guidance.
- The download/verify boundary lives in `internal/release` (httptest
  seam: `Endpoint` for the check, `DownloadBase` for the artifacts); the
  extract/replace mechanism and orchestration live in
  `internal/selfupdate`, with `exePath` always injected so tests never
  risk the test runner.
- **Verification**: a paired test proves the security claim — a *tampered*
  archive (wrong `checksums.txt` hash) makes `--apply` error with the
  target file **unchanged**, while a *correct* hash replaces the target
  with the new bytes. Plus per-channel routing tests and a Windows-gate
  test.
- No new third-party dependency; the binary stays dependency-free.

## Alternatives Considered

- **Binary swap for `go-install` too.** Rejected: replacing a
  toolchain-managed binary with a GoReleaser artifact is surprising and
  breaks the `go install` mental model; the native command is safer.
- **Require cosign verification before swap.** Rejected for v0.9.4: it
  adds a dependency the dependency-free installer deliberately avoids, and
  install.sh sets the transport-integrity bar this matches. Documented as
  a future option, not a gap.
- **Keep a `.bak` copy and restore on failure.** Rejected: POSIX atomic
  rename already guarantees no partial state; a backup adds a cleanup
  path and a second failure mode for no safety gain.
- **Defer all of `--apply` to v0.9.9 with brew/npm.** Rejected by
  ADR 0032 D2: the install-script branch is high-value and depends on
  nothing new.
