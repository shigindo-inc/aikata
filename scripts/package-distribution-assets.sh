#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_root"

tar -czf dist/aikata-universal-skill.tar.gz -C dist universal-skill
tar -czf dist/aikata-codex-plugin.tar.gz -C dist/codex plugin
