---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-19
audience: [human, agent]
purpose: code agent への初期セットアップ指針と、立ち上げに必要な追加情報
---

# aikata — Setup Guide & Agent Instructions

このドキュメントは、`DESIGN.md` を補完する形で、**実際にリポジトリを立ち上げて code agent に作業を依頼する際の指針**をまとめたものです。

`DESIGN.md` が「何を作るか」を語るのに対し、この文書は「どう立ち上げるか」「agent にどう指示するか」を扱います。

---

## 1. DESIGN.md の扱い

### 1.1 DESIGN.md の位置づけ

`DESIGN.md` は **企画フェーズの統合文書** であり、運用フェーズに入ったら以下のように分解する必要があります。

```
DESIGN.md (現状)
 ├── 1. コンセプト         → SPEC.md (What/Why部分) + README.md (一行説明)
 ├── 2. スコープ           → SPEC.md
 ├── 3. ファイル構成       → ARCHITECTURE.md (生成物の構造として)
 ├── 4. CLI 仕様           → SPEC.md (機能要件として)
 ├── 5. 設定ファイル        → ARCHITECTURE.md
 ├── 6. 実害ゼロ設計        → docs/adr/0001-do-no-harm-policy.md
 ├── 7. 実装方針            → ARCHITECTURE.md
 ├── 8. ロードマップ        → ROADMAP.md
 ├── 9. 今後の検討事項      → docs/decisions/open-questions.md
 └── 10. 参考リンク         → 各ドキュメント末尾 or docs/references.md
```

### 1.2 推奨フロー

1. **リポジトリ初期化時**: `DESIGN.md` を `docs/origin/initial-design.md` に保存（歴史的記録として）
2. **agent に分解依頼**: 上記マッピングに従って各ファイルを生成
3. **dogfooding**: aikata 自身が aikata で生成された構造を持つようになる

---

## 2. 初期セットアップ手順

### 2.1 リポジトリ作成（人間が実行）

```bash
# 1. GitHub でリポジトリ作成（プライベートで開始推奨）
gh repo create aikata --private --description "AI-readable markdown scaffold for modern projects"

# 2. ローカルにクローン
git clone git@github.com:<user>/aikata.git
cd aikata

# 3. このドキュメントと DESIGN.md を配置
mkdir -p docs/origin
cp /path/to/DESIGN.md docs/origin/initial-design.md
cp /path/to/SETUP.md docs/origin/initial-setup.md

# 4. 初期 commit
git add docs/origin
git commit -m "docs: add initial design and setup documents"
git push -u origin main
```

### 2.2 ライセンス・基本ファイル選定（人間が決める）

事前に決めておくべき項目：

| 項目 | 推奨 | 備考 |
|---|---|---|
| ライセンス | MIT または Apache 2.0 | OSS として最も受け入れられやすい |
| 言語 | Go (第一候補) | DESIGN.md 7.1 参照 |
| 最小 Go バージョン | 1.21+ | 安定版採用 |
| パッケージマネージャ | Go modules | 標準 |
| CI | GitHub Actions | 無料枠で十分 |
| Lint | golangci-lint | de facto |
| フォーマッタ | gofmt + goimports | 標準 |
| テスト | Go 標準 testing + testify | |
| リリースツール | GoReleaser | バイナリ配布の自動化 |

**Go か TypeScript か決まっていない場合**: 先に Go で start し、不都合があれば TypeScript に switch する方が後戻りしやすい。

### 2.3 code agent への最初のタスク

最初に agent に依頼するタスクの順序：

1. **DESIGN.md の分解** → 上記マッピングに従ったファイル生成
2. **Go プロジェクトの初期化** → `go mod init` + 基本ディレクトリ構造
3. **CI セットアップ** → `.github/workflows/ci.yml`
4. **最小機能の実装** → `aikata --version`, `aikata --help` だけ動く状態
5. **`aikata init --preset minimal` の実装**
6. **テンプレート埋め込み機構の実装**

各タスクの詳細プロンプトテンプレートは Appendix A 参照。

---

## 3. 技術的な前提・制約

### 3.1 サポート対象

- **OS**: macOS (Intel/ARM), Linux (x86_64/ARM64), Windows (x86_64)
- **Go バージョン**: 1.21+
- **配布形式**:
  - GitHub Releases バイナリ
  - Homebrew tap (`aikata-dev/tap`)
  - `curl -sSL https://aikata.dev/install.sh | sh`
  - npm wrapper（後続フェーズ）

### 3.2 依存方針

- **外部依存は最小化**: Go 標準ライブラリで済むものは標準で
- **採用候補ライブラリ**:
  - CLI フレームワーク: `cobra` (de facto)
  - 対話プロンプト: `huh` (charmbracelet) または `survey`
  - テンプレート: Go 標準 `text/template`
  - YAML: `gopkg.in/yaml.v3`
  - フォーマット出力: `lipgloss` (charmbracelet)
- **避けるもの**: 重い ORM、GUI ライブラリ、外部 API クライアント

### 3.3 ディレクトリ規約

```
aikata/
├── cmd/
│   └── aikata/
│       └── main.go             # エントリポイントのみ、ロジックなし
├── internal/                   # 外部から import 不可
│   ├── cli/                    # CLI コマンド定義（cobra）
│   ├── scaffold/               # ファイル生成ロジック
│   ├── doctor/                 # 整合性チェック
│   ├── generate/               # AI ツール向け生成
│   ├── config/                 # .ai/aikata.yaml の読み書き
│   ├── presets/                # preset 定義
│   └── templates/              # テンプレート処理
├── templates/                  # 埋め込み markdown テンプレート
│   ├── base/
│   ├── presets/
│   └── ai_tools/
├── docs/                       # aikata 自身のドキュメント
│   ├── adr/
│   ├── origin/                 # 企画段階の歴史的文書
│   └── decisions/
├── examples/                   # 利用例
├── testdata/                   # ゴールデンテスト用
├── .github/
│   └── workflows/
├── SPEC.md
├── ARCHITECTURE.md
├── ROADMAP.md
├── AGENTS.md
├── GLOSSARY.md
├── README.md
├── CHANGELOG.md
├── LICENSE
├── go.mod
├── go.sum
├── Makefile
└── .gitignore
```

### 3.4 エラーハンドリング方針

- **panic は禁止**: ライブラリ的に使われる可能性も考慮し、error を返す
- **エラーラッピング**: `fmt.Errorf("doing X: %w", err)` 形式で context を保持
- **ユーザー向けメッセージ**: 技術的詳細は `--verbose` でのみ表示、デフォルトは平易に
- **exit code**:
  - `0`: 成功
  - `1`: 一般的なエラー
  - `2`: ユーザー入力エラー
  - `3`: 整合性チェック失敗（doctor）

### 3.5 ロギング方針

- 標準出力: ユーザー向け人間可読出力
- 標準エラー出力: warning, error, verbose log
- `--quiet` で標準出力を抑制
- `--verbose` で詳細ログを表示
- `--json` で機械可読出力（CI連携用、v0.3以降）

---

## 4. 品質基準

### 4.1 テスト

- **カバレッジ目標**: コアロジック（`internal/scaffold`, `internal/doctor` など）は 80% 以上
- **ゴールデンテスト**: 各 preset の生成結果を `testdata/golden/` と比較
- **統合テスト**: 実際にバイナリを叩いて end-to-end 確認
- **CI で必須**: PR マージ前に全テスト pass

### 4.2 コードレビュー観点

agent が PR を作成する際の self-review チェックリスト：

- [ ] 公開 API（exported functions）には godoc コメントがあるか
- [ ] エラーメッセージは action item を含むか（"X failed" ではなく "X failed: try Y"）
- [ ] ユーザー入力（ファイルパス・YAML 値）の検証があるか
- [ ] テストケースで edge case（空文字、nil、巨大ファイル）をカバーしているか
- [ ] OS 依存のパス区切り文字を使っていないか（`filepath.Join` 使用）
- [ ] 既存テストが全て pass するか
- [ ] `golangci-lint run` で warning がないか

### 4.3 Commit Message 規約

**Conventional Commits** を採用：

```
<type>(<scope>): <subject>

<body>

<footer>
```

type: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `ci`, `build`

例:
```
feat(scaffold): add minimal preset

Implements the simplest preset that generates only README, AGENTS, SPEC.
This serves as the foundation for other presets.

Refs: #1
```

**AI 署名禁止**: commit message に `Co-Authored-By: Claude` 等の AI 署名を入れない（DESIGN.md 6 節と整合）。

### 4.4 Branch 戦略

- `main`: 常に release 可能
- `feat/*`, `fix/*`, `docs/*`: 機能/修正ブランチ
- PR ベースで `main` にマージ、squash merge 推奨
- リリースは tag (`v0.1.0`) でトリガー

---

## 5. agent 利用のベストプラクティス

### 5.1 agent に依頼する単位

- **1 task = 1 PR**: 大きすぎる task は分割する
- **明確な完了条件**: 「テスト pass」「ドキュメント更新」「golangci-lint pass」を必ず含める
- **既存コード読込指示**: 関連ファイルを明示的に列挙

### 5.2 推奨プロンプトパターン

各 task のプロンプトは以下の構造に従う：

```
## 背景
[なぜこの task が必要か]

## 完了条件
- [ ] X が動くこと
- [ ] Y のテストが pass すること
- [ ] Z のドキュメントが更新されていること

## 参照すべきファイル
- DESIGN.md (該当節)
- SETUP.md (該当節)
- 既存の関連ファイル

## 制約
- [禁止事項、避けるべきパターン]

## 想定アウトプット
- ファイル一覧
- 動作確認コマンド
```

### 5.3 agent への共通指示（AGENTS.md に書くべき内容）

**重要**: aikata 自身の `AGENTS.md` には以下を含める：

- DESIGN.md / SETUP.md / SPEC.md / ARCHITECTURE.md の参照順序
- Commit message 規約
- AI 署名禁止
- テスト追加必須
- 既存テストを壊さない
- `docs/tasks/current.md` を更新する

具体的テンプレートは Appendix B 参照。

---

## 6. 立ち上げチェックリスト

### Phase 0: 準備（人間が実施）

- [ ] GitHub username/organization の決定
- [ ] npm / PyPI / domain (aikata.dev) の空き確認
- [ ] ライセンス確定（MIT 推奨）
- [ ] 言語確定（Go 推奨）
- [ ] リポジトリ作成（private で開始）
- [ ] DESIGN.md / SETUP.md を `docs/origin/` に配置
- [ ] 初期 commit

### Phase 1: 雛形（agent に依頼可能）

- [ ] DESIGN.md の分解（SPEC.md, ARCHITECTURE.md, ROADMAP.md, ADR 生成）
- [ ] README.md 初版
- [ ] LICENSE ファイル
- [ ] .gitignore (Go 標準 + aikata 固有)
- [ ] Makefile (build, test, lint, install)
- [ ] AGENTS.md（aikata 自身の agent 指示）
- [ ] GLOSSARY.md（aikata 自身の用語集）

### Phase 2: Go プロジェクト初期化（agent に依頼可能）

- [ ] `go mod init github.com/<user>/aikata`
- [ ] ディレクトリ構造作成
- [ ] `cmd/aikata/main.go` でバージョン表示のみ
- [ ] cobra 導入、`--version`, `--help` 動作
- [ ] golangci-lint 設定
- [ ] CI セットアップ（GitHub Actions）

### Phase 3: MVP 実装（複数 PR に分割、agent に依頼可能）

- [ ] テンプレート埋め込み機構（`embed.FS`）
- [ ] `aikata init --preset minimal` 実装
- [ ] ゴールデンテスト基盤
- [ ] `aikata init` の対話モード
- [ ] `.ai/aikata.yaml` の読み書き

### Phase 4: 拡張（v0.2 以降）

- [ ] Flutter preset
- [ ] `--lang ja` 対応
- [ ] `aikata generate` （Claude Code 向け）
- [ ] `aikata doctor`

各 Phase の完了基準は `ROADMAP.md` を参照。

---

## 7. agent 起動時のクイックスタート

新しい agent session で作業を始める時、以下の順で読ませる：

1. `README.md` (プロジェクト概要)
2. `AGENTS.md` (agent 行動指針)
3. `docs/origin/initial-design.md` (= DESIGN.md, 全体像把握)
4. `docs/origin/initial-setup.md` (= SETUP.md, 立ち上げ手順)
5. `SPEC.md` / `ARCHITECTURE.md` (該当 task に応じて)
6. `docs/tasks/current.md` (現在の作業状況)

prompt 例：
```
このリポジトリは aikata という OSS の CLI ツールです。
まず README.md, AGENTS.md, docs/origin/ を読んで全体像を把握してください。
その後、docs/tasks/current.md に書かれた次のタスクを進めてください。
```

---

## 8. リスクと対策

### 8.1 想定リスク

| リスク | 対策 |
|---|---|
| agent が DESIGN.md を読まずに実装を始める | AGENTS.md で必読指定、初手で読込確認 |
| npm / PyPI / ドメイン `aikata` の衝突 | Phase 0 で空き確認 |
| Go の学習コスト | 既存 OSS（cobra, charm ツール群）のコード参考 |
| AI ツール側の仕様変更 | 各 AI ツール向け generate を modular に設計 |
| Scope creep | ROADMAP.md と open-questions.md で常に整理 |

### 8.2 早めに decision したい事項

- ライセンス（MIT or Apache 2.0）
- Go か TypeScript か（後戻りコスト大）
- 言語デフォルト（en or ja）

---

## Appendix A: 初期タスク用プロンプトテンプレート

### A.1 DESIGN.md 分解タスク

```
## 背景
docs/origin/initial-design.md に aikata の統合設計書がある。
これを運用ドキュメントに分解する。

## 完了条件
- [ ] SPEC.md が生成され、What/Why が記載されている
- [ ] ARCHITECTURE.md が生成され、技術選定とディレクトリ構造が記載されている
- [ ] ROADMAP.md が生成され、v0.1〜v1.0 のマイルストーンが記載されている
- [ ] docs/adr/0001-do-no-harm-policy.md が生成されている
- [ ] docs/decisions/open-questions.md が生成されている
- [ ] README.md が最薄構成で生成されている
- [ ] 各ファイルに frontmatter (updated, audience) がある

## 参照すべきファイル
- docs/origin/initial-design.md
- docs/origin/initial-setup.md (このファイル) の 1.1 節のマッピング表

## 制約
- DESIGN.md の内容を失わない
- ただし各ファイルが「最初に分解した時点での状態」になるよう、不要な重複は削除

## 想定アウトプット
- 6 ファイルの新規作成
- commit: docs: split initial design into operational documents
```

### A.2 Go プロジェクト初期化タスク

```
## 背景
Go で CLI ツール aikata を実装する準備をする。

## 完了条件
- [ ] go.mod が初期化されている (module: github.com/<user>/aikata)
- [ ] cmd/aikata/main.go が存在し、aikata --version が動く
- [ ] cobra が導入されている
- [ ] internal/ 配下に空ディレクトリ構造ができている
- [ ] Makefile に build, test, lint, install ターゲットがある
- [ ] .golangci.yml が設定されている
- [ ] .github/workflows/ci.yml で test と lint が動く
- [ ] make build でバイナリができる
- [ ] make test で（テストがなくても）通る

## 参照すべきファイル
- ARCHITECTURE.md (ディレクトリ構造)
- docs/origin/initial-setup.md (技術的前提)

## 制約
- 外部依存は cobra のみ
- panic 禁止、エラーは return

## 想定アウトプット
- 上記ファイル群
- commit: chore: initialize Go project structure
```

### A.3 `aikata init --preset minimal` 実装タスク

```
## 背景
MVP の核機能。最小 preset で README, AGENTS, SPEC のみ生成する。

## 完了条件
- [ ] templates/presets/minimal/ に 3 ファイルのテンプレートがある
- [ ] embed.FS でテンプレートが埋め込まれている
- [ ] aikata init --preset minimal --no-interactive --name foo が動く
- [ ] 生成された README.md にプロジェクト名が反映されている
- [ ] testdata/golden/minimal/ とのゴールデンテストが pass する
- [ ] 既存ディレクトリでの実行は --force なしではエラー
- [ ] go test ./... が全て pass

## 参照すべきファイル
- SPEC.md (4.2 節 aikata init 詳細)
- ARCHITECTURE.md

## 制約
- ファイル生成は atomic に（途中で失敗したら全部 rollback）
- パスは filepath.Join で OS 中立に

## 想定アウトプット
- internal/scaffold/scaffold.go (主実装)
- internal/scaffold/scaffold_test.go
- templates/presets/minimal/ 配下のテンプレート
- testdata/golden/minimal/ 配下の期待出力
- commit: feat(scaffold): implement minimal preset
```

---

## Appendix B: aikata 自身の AGENTS.md テンプレート

aikata リポジトリ自身に置くべき AGENTS.md：

```markdown
---
audience: agent
updated: 2026-05-19
---

# Agent Instructions for aikata Development

## Project Overview
aikata is a CLI tool that scaffolds AI-readable markdown documents for projects.
See [SPEC.md](./SPEC.md) for what and why.

## Before You Start
Read in order:
1. README.md
2. This file (AGENTS.md)
3. SPEC.md
4. ARCHITECTURE.md
5. docs/tasks/current.md (current work)

For context, the original design document is at docs/origin/initial-design.md.

## Navigation Matrix
| Task type | Read these files |
|---|---|
| Add a new preset | ARCHITECTURE.md, internal/presets/, templates/presets/ |
| Modify CLI | SPEC.md (4節), internal/cli/ |
| Fix bug | docs/tasks/current.md, related test files |
| Add AI tool support | ARCHITECTURE.md, internal/generate/, templates/ai_tools/ |
| Update docs | docs/, relevant top-level .md |

## Hard Rules
- **Read DESIGN.md (docs/origin/initial-design.md) before non-trivial changes** — it has full context
- **Update docs/tasks/current.md when you start/finish work**
- **Never commit secrets** — reference .env.example
- **Add tests for new code** — no exceptions for core logic
- **Use Conventional Commits** — feat/fix/docs/refactor/test/chore
- **No AI signatures in commits** — no "Co-Authored-By: Claude" etc.
- **Run before commit**: `make test && make lint`
- **OS-neutral paths** — always use filepath.Join

## Code Style
- panic 禁止、error を返す
- exported functions には godoc コメント
- error wrap: `fmt.Errorf("doing X: %w", err)`
- 外部依存は最小化（理由なく増やさない）

## When Stuck
- Check docs/decisions/open-questions.md
- Check docs/adr/ for past decisions
- If still unclear, document the question and ask the human
```

---

## Appendix C: 確認しておくべき外部要素

立ち上げ前に人間が確認/決定すべき項目：

- [ ] `aikata` の npm 可用性: `npm view aikata`
- [ ] `aikata` の PyPI 可用性: `pip index versions aikata`
- [ ] `aikata` の GitHub organization 可用性
- [ ] `aikata.dev` ドメイン可用性
- [ ] `aikata.io` ドメイン可用性

> **注意**: 既存に同名 OSS がある可能性が高いため、リリース前に必ず GitHub / npm / PyPI で検索すること。

---

*このドキュメントは agent に依頼する前提のため、DESIGN.md より具体的・操作的に書かれています。実際の運用では `docs/origin/` に保管し、必要に応じて agent に読ませてください。*
