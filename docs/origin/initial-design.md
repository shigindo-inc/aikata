---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-19
audience: [human, agent]
---

# aikata — Design Document

> AI開発時代のプロジェクト立ち上げを、軽量な markdown scaffold と多ツール対応で支援する OSS ツール。

このドキュメントは aikata の設計思想・スコープ・実装方針・未解決の検討事項を一箇所にまとめたものです。LLM が読んで実装に着手できる粒度を目指します。

---

## 1. コンセプト

### 1.1 一行説明

**`aikata` は、AI Coding 時代のプロジェクトに必要な markdown ドキュメント群と各 AI ツール向け設定ファイルを、1コマンドで scaffold する軽量 CLI ツール。**

### 1.2 名前の由来

`aikata`（相方）= AI と人間が対等に協働する開発における「相棒」。LLM が長い context の中で迷わず、プロジェクトの文脈を共有する伴走者、という比喩。

### 1.3 ポジショニング

- **競合**: ai-rulez, Ruler, block/ai-rules, agentsmesh
- **既存ツールが共有する前提**: rules 中心、英語前提、汎用、init 時に一回設定して終わり
- **aikata の差別化**:
  - **ドキュメント（md）中心** — rules ではなく「人間とLLMが共に読むプロジェクト文書」が起点
  - **軽量・opinionated だが小さい** — Vite/Astro 的ポジション、ai-rulez の Terraform 的重さの対極
  - **実害ゼロ設計** — オプション機能（Obsidian, TDD, Flutter）を使わないユーザーに不利益が出ない
  - **段階的拡張** — 最小構成から始めて必要に応じて拡張、機能てんこ盛りではない

### 1.4 設計原則（Design Principles）

1. **Human-LLM dual readable**: 全てのドキュメントは人間にもLLMにも読みやすく
2. **Convention over configuration**: 最小設定で動く、ただし全て上書き可能
3. **Do no harm**: ユーザーが採用しない機能で実害が出ない（後述「実害ゼロ設計」）
4. **Top-level minimalism**: トップディレクトリは8ファイル以内
5. **Composable, not monolithic**: 機能はファイル単位で足し引き可能
6. **Stack-agnostic core, opinionated presets**: コアは汎用、preset で stack 特化
7. **Lossy generation is OK**: 各 AI ツール向け生成物は使い捨て前提、canonical source が真実

---

## 2. スコープ

### 2.1 やること（In Scope）

- プロジェクト立ち上げ時の md ドキュメント群の scaffold
- 各 AI ツール向け設定ファイル（CLAUDE.md, .cursor/rules/, AGENTS.md など）の生成
- 整合性チェック（doctor コマンド）
- テンプレート更新の差分マージ（update コマンド）
- Stack 別 preset（初期は Flutter）
- モノレポ対応（nested 構造）

### 2.2 やらないこと（Out of Scope）

- IDE プラグイン提供（CLI のみ）
- rule の real-time enforcement（ai-rulez の領域）
- LLM API 呼び出しでのドキュメント自動生成（v1 では）
- AI ツール側の設定 GUI
- プロジェクト管理機能（タスク追跡、issue tracker 連携など）

### 2.3 想定ユーザー

- **Primary**: 個人開発者・小規模チームで複数の AI Coding ツール（Claude Code, Cursor, Codex, Gemini CLI 等）を併用する人
- **Secondary**: OSS プロジェクトメンテナで contributor 向けに AI-friendly な構造を提供したい人
- **Tertiary**: 日本語圏で社内ドキュメント規約を持つチーム

---

## 3. ファイル構成（Generated Structure）

### 3.1 デフォルト構造（`aikata init` 実行後）

```
/project
├─ README.md              # human + LLM 入口（最薄、ナビゲーションのみ）
├─ AGENTS.md              # LLM 行動制御 + 他ドキュメントへのナビゲーション
├─ SPEC.md                # What / Why（要件・目的）
├─ ARCHITECTURE.md        # How（技術選定・構造）
├─ GLOSSARY.md            # ドメイン用語・略語定義
├─ .env.example           # 環境変数テンプレート
├─ .gitignore             # 生成物・秘密情報の除外設定
├─ docs/
│   ├─ adr/               # Architecture Decision Records
│   │   └─ 0001-record-architecture-decisions.md
│   ├─ stacks/            # Stack 固有ガイドライン
│   │   └─ (preset依存)
│   ├─ tasks/             # Agent が書き換える作業状況
│   │   └─ current.md
│   ├─ troubleshooting.md # 既知の罠・ハマりどころ
│   └─ prompts.md         # 再利用可能なプロンプト集
└─ .ai/                   # AI ツール向け生成物の置き場（gitignore対象）
    └─ aikata.yaml        # aikata 自体の設定ファイル
```

### 3.2 オプショナルファイル（init 時のフラグや preset で追加）

| ファイル | 追加条件 | 用途 |
|---|---|---|
| `UI.md` | `--with-ui` または UI 系 preset | UI/UX ガイドライン |
| `API.md` | `--with-api` または API 系 preset | API インターフェース仕様 |
| `docs/testing.md` | `--with-tdd` | テスト戦略（TDD 採用時） |
| `CHANGELOG.md` | `--with-changelog` | 変更履歴 |
| `CONTRIBUTING.md` | `--oss` | OSS 化時の contributor 向け |
| `SECURITY.md` | `--oss` | セキュリティポリシー |
| `ROADMAP.md` | `--oss` | ロードマップ |

### 3.3 各ファイルの責務

#### `README.md`
- **読者**: human (primary), LLM (secondary)
- **責務**: プロジェクトの一行説明、quickstart、他ドキュメントへのリンク
- **NOT 責務**: 詳細な仕様、技術選定の理由（→ SPEC.md, ARCHITECTURE.md）
- **長さ目安**: 100 行以内

#### `AGENTS.md`
- **読者**: LLM (primary), human (secondary)
- **責務**: AI agent への行動指示、他 md へのナビゲーション、変更タイプ別の参照ファイルマトリクス
- **形式**: agents.md open spec に準拠
- **長さ目安**: 200 行以内（context 食い過ぎ防止）

#### `SPEC.md`
- **責務**: What（何を作るか）と Why（なぜ作るか）
- **境界**: How は ARCHITECTURE.md、UI は UI.md
- **構成**: 目的、ユーザー、機能要件、非機能要件、スコープ外

#### `ARCHITECTURE.md`
- **責務**: How（どう作るか）、主要技術選定、システム構造
- **境界**: 個別の判断記録は `docs/adr/` へ
- **構成**: 概要、技術スタック、コンポーネント図、データフロー、ディレクトリ構造

#### `GLOSSARY.md`
- **責務**: ドメイン用語、社内用語、略語の定義
- **重要性**: 日本語プロジェクトで特に効く（LLM の誤訳防止）

#### `docs/adr/`
- **責務**: 設計判断ログ（Architecture Decision Records）
- **形式**: 一案件一ファイル、`NNNN-title.md` 命名
- **テンプレ**: Status / Context / Decision / Consequences

#### `docs/tasks/current.md`
- **責務**: 現在進行中の作業状況
- **特徴**: Agent が頻繁に書き換える前提、トップレベルから隔離

---

## 4. CLI 仕様

### 4.1 コマンド一覧

```
aikata init [name]            # 新規 scaffold
aikata add <component>        # 個別ファイル/機能の追加
aikata doctor                 # 整合性チェック
aikata generate               # AI ツール向けファイル生成
aikata update                 # テンプレート更新の差分マージ
aikata list                   # 利用可能な preset / component 一覧
```

### 4.2 `aikata init` 詳細

#### Flags

```
--preset <name>          # minimal | standard | flutter | typescript | ... (default: standard)
--with-ui                # UI.md を含める
--with-api               # API.md を含める
--with-tdd               # testing.md を含める
--with-changelog         # CHANGELOG.md を含める
--oss                    # OSS 向けファイル一式を含める
--monorepo               # モノレポ構造で初期化
--lang <ja|en>           # ドキュメントの言語 (default: en)
--ai-tools <list>        # claude,cursor,codex,gemini,copilot,windsurf,... (default: claude)
--minimal                # 最小構成（README, AGENTS, SPEC のみ）
--no-interactive         # 対話モードを使わない
--dry-run                # ファイル生成せず、計画のみ表示
```

#### 対話モード

`--no-interactive` 以外では、以下を順に質問:
1. プロジェクト名
2. 言語（ja/en）
3. 使用する AI ツール（複数選択）
4. Stack preset（minimal / flutter / typescript / custom）
5. オプション機能（UI, API, TDD, monorepo）
6. OSS にする予定か

#### 既存ディレクトリでの実行

- 既存ファイルがある場合は **上書きしない**、`.aikata-proposed/` に提案を出す
- `--force` フラグで上書き可能

### 4.3 `aikata add <component>` 詳細

例:
```
aikata add ui              # UI.md を追加
aikata add adr <title>     # 新しい ADR を追加
aikata add stack flutter   # Flutter stack を追加
aikata add ai-tool cursor  # Cursor 向け生成を有効化
```

### 4.4 `aikata doctor` 詳細

チェック項目:
- AGENTS.md が参照しているファイルが存在するか
- GLOSSARY.md で定義された用語が他のドキュメントで使われているか（逆方向は警告のみ）
- ADR のステータスが Deprecated になっていないか
- `.env.example` の変数が AGENTS.md / ARCHITECTURE.md に記載されているか
- 各ドキュメントの最終更新日が極端に古くないか（warning）
- frontmatter の必須フィールドが揃っているか

出力レベル: `error` / `warning` / `info`、`--fix` で自動修正可能なものは修正

### 4.5 `aikata generate` 詳細

`.ai/aikata.yaml` を読んで、有効化されている AI ツール向けにファイル生成。

例:
- `claude` 有効 → `CLAUDE.md` をルートに生成（AGENTS.md + 関連ファイルの統合版）
- `cursor` 有効 → `.cursor/rules/*.mdc` を生成
- `copilot` 有効 → `.github/copilot-instructions.md` を生成

**ポイント**:
- 生成物は `.gitignore` に自動追加（オプトアウト可能）
- 各 AI ツールの仕様変更に追随するためのバージョニング機構

### 4.6 `aikata update` 詳細

aikata のテンプレートが更新されたとき、ユーザー編集を保ったまま差分マージ。

- diff を表示して accept/reject を選択
- Copier の同様機能を参考

---

## 5. 設定ファイル: `.ai/aikata.yaml`

### 5.1 構造例

```yaml
version: 1
project:
  name: my-app
  lang: ja
  description: "日本語の説明"

# 有効化する AI ツール
ai_tools:
  - claude
  - cursor
  - codex

# 採用している stack
stacks:
  - flutter

# 機能フラグ
features:
  tdd: true
  obsidian_hints: false  # Obsidian 向けの軽量ヒントを含めるか
  monorepo: false

# ドキュメント設定
docs:
  generate_gitignore: true
  task_file_location: docs/tasks/current.md

# AI ツール別オーバーライド
overrides:
  claude:
    output: CLAUDE.md
    include: [AGENTS.md, SPEC.md, ARCHITECTURE.md, GLOSSARY.md]
  cursor:
    output_dir: .cursor/rules/
    split_by: domain
```

---

## 6. 実害ゼロ設計（Do No Harm Policy）

aikata の核心ポリシー。**採用しない機能のユーザーに不利益を与えない**。

### 6.1 Obsidian について

- **wikilink を使わない**: 標準 markdown link `[Foo](./Foo.md)` のみ使用
  - Obsidian は標準 link も解釈するため、Obsidian ユーザーにも実害なし
- **`.obsidian/` ディレクトリは生成しない**: ユーザーが Obsidian で開いた時のみ生成される
- **Dataview/Tasks クエリは書かない**: 書く場合は `docs/obsidian-views/` に隔離
- **frontmatter は最小限**: YAML で `updated`, `audience` 程度。Obsidian Properties としても、ただのメタデータとしても機能

### 6.2 TDD について

- `--with-tdd` を付けない限り `docs/testing.md` は生成されない
- AGENTS.md は条件付き参照: `## Testing (if testing.md exists)` のような構造
- TDD 採用フラグが OFF の時、AGENTS.md に「テストを先に書け」rule は **含まれない**

### 6.3 Flutter について

- v0 の preset として提供するが、コアは stack-agnostic
- `aikata init --preset minimal` で Flutter 要素ゼロのスケルトンも生成可能
- Flutter 固有の rule は `docs/stacks/flutter.md` に隔離、AGENTS.md からは include 形式

### 6.4 モノレポについて

- `--monorepo` フラグ無しでは単一プロジェクト構造
- モノレポ対応時もルート構造を破壊せず、`apps/*/AGENTS.md` のような nested 構造で対応

### 6.5 AI ツールの選択について

- `--ai-tools` のデフォルトは `claude` のみ
- 不要な `.cursor/rules/` などのディレクトリを勝手に作らない

### 6.6 言語について

- `--lang en` がデフォルト（OSS としての到達範囲を優先）
- `--lang ja` で日本語テンプレートに切り替え
- 将来的に bilingual モード（人間向け日本語 + LLM 向け英語）も検討

---

## 7. 実装方針

### 7.1 言語選定

候補:
- **Go**: 単一バイナリ配布が容易、ai-rulez と同じ
- **Rust**: パフォーマンス・配布性は良いが学習コスト
- **TypeScript (Node)**: npm エコシステム親和性、`npx aikata init` の手軽さ
- **Bash + POSIX sh**: 究極の軽量、ただし複雑な機能は厳しい

**推奨**: **Go** を第一候補、TypeScript を第二候補。
- 理由: 単一バイナリで `curl | sh` インストール可能、aikata の「軽量」哲学に合致
- npm 配布も併用したい場合は TypeScript を選ぶ

### 7.2 テンプレートエンジン

- 各 md ファイルを Go template / Handlebars 形式で記述
- frontmatter で各テンプレートのメタデータ管理
- preset は templates ディレクトリ階層で表現

### 7.3 配布

- GitHub Releases でバイナリ配布
- Homebrew tap (`brew install aikata`)
- npm パッケージ（`npx aikata`）— Go 製でも npm wrapper で配布可能
- `curl -sSL https://aikata.dev/install.sh | sh`

### 7.4 リポジトリ構造（aikata 自体の）

```
aikata/
├─ cmd/aikata/              # CLI エントリーポイント
├─ internal/
│   ├─ scaffold/            # ファイル生成ロジック
│   ├─ doctor/              # 整合性チェック
│   ├─ generate/            # AI ツール向け生成
│   ├─ presets/             # preset 定義
│   └─ templates/           # md テンプレート
├─ templates/               # 埋め込みテンプレート
│   ├─ base/                # 共通
│   ├─ presets/
│   │   ├─ minimal/
│   │   ├─ standard/
│   │   ├─ flutter/
│   │   └─ ...
│   └─ ai_tools/
│       ├─ claude/
│       ├─ cursor/
│       └─ ...
├─ docs/                    # aikata 自体のドキュメント（dogfooding）
├─ examples/                # 利用例
├─ SPEC.md
├─ ARCHITECTURE.md
├─ AGENTS.md
└─ README.md
```

### 7.5 テスト戦略

- ゴールデンテスト（生成結果が期待ファイル群と一致するか）
- preset ごとの統合テスト
- CI で複数 OS（macOS, Linux, Windows）

---

## 8. ロードマップ（暫定）

### v0.1 (MVP)
- `aikata init` で `--preset minimal` `--preset standard` の 2 種類
- 対応 AI ツール: Claude Code のみ
- 言語: en のみ
- 基本ファイル: README, AGENTS, SPEC, ARCHITECTURE, GLOSSARY, .env.example
- `aikata generate` の最小実装

### v0.2
- `--preset flutter` 追加
- `--lang ja` 対応
- Cursor / Codex 向け generate 対応
- `aikata doctor` の基本実装

### v0.3
- `aikata add` コマンド
- `--with-ui`, `--with-api`, `--with-tdd` オプション
- ADR の自動採番・テンプレート

### v0.4
- モノレポ対応（`--monorepo`）
- `aikata update` 実装

### v1.0
- 主要 AI ツール（Claude, Cursor, Codex, Gemini, Copilot, Windsurf）対応
- 安定 API
- 公式ドキュメントサイト
- preset の external repository 対応

### v1.x 以降の検討
- LLM API 統合でドキュメント自動生成支援
- VS Code 拡張機能
- 既存プロジェクトからのリバース解析（agentsmesh 的機能）

---

## 9. 今後の検討事項（Open Questions）

### 9.1 設計上の未解決事項

1. **AGENTS.md と CLAUDE.md の関係**
   - `AGENTS.md` を canonical source として `CLAUDE.md` を生成するか
   - それとも `AGENTS.md` 自体を Claude Code に読ませる（生成不要）か
   - 現状: agents.md open spec に賭けるなら後者、Claude 専用機能を活かすなら前者

2. **生成物を git で管理するか**
   - デフォルトで `.gitignore`、ただし「チーム全員が aikata を入れない」ケースもある
   - 解決案: フラグで切り替え、デフォルトは ignore

3. **多言語ドキュメントの構造**
   - `SPEC.md` と `SPEC.ja.md` の二重管理を避ける方法
   - 解決案: 翻訳テーブル方式、または LLM による on-demand 翻訳

4. **preset の合成**
   - `--preset flutter --preset oss` のような複数 preset 指定をどう解決するか
   - 解決案: preset を「機能フラグの組み合わせ」として定義

5. **`docs/tasks/current.md` の更新責任**
   - LLM が頻繁に書き換える前提、人間も書き換える
   - コンフリクト時の挙動をどう設計するか

6. **stack-agnostic core の境界**
   - Flutter preset がコアにあまり依存しないようにするには
   - plugin 機構を v1 までに用意するか、v2 以降か

### 9.2 名前・ブランディング

- `aikata` という名前が npm / PyPI / GitHub / domain で空いているか確認必要
- 海外ユーザーの発音・綴りの明確化（「アイカタ」等）
- ロゴ・サイトデザイン

### 9.3 既存ツールとの相互運用

- ai-rulez の config を import できるべきか
- agents.md open spec への準拠レベル
- 既存の `CLAUDE.md` をユーザーが持っている場合の取り込み手段

### 9.4 エコシステム

- 外部 preset の配布方法（GitHub repo 参照？）
- コミュニティテンプレートのレビュー体制
- 商用利用を見据えたライセンス（MIT or Apache 2.0 推奨）

### 9.5 検証すべき仮説

- **H1**: 「複数 AI ツール併用ユーザーは、ai-rulez よりも軽量なツールを好む」
- **H2**: 「ドキュメント中心の scaffold は rules 中心より受け入れられやすい」
- **H3**: 「Flutter 開発者は AI Coding scaffold を欲しがっている」
- **H4**: 「日本語 OSS としてのアイデンティティはむしろ強みになる」

各仮説を検証するための初期ユーザー（Satoshi 自身 + 周辺コミュニティ）でのドッグフーディング計画が必要。

### 9.6 競合差別化の継続課題

- ai-rulez が「軽量モード」を出してきた場合の対抗策
- agents.md spec が普及した場合に aikata の存在意義をどう保つか
- 「ドキュメント中心」というコンセプトを陳腐化させない継続的な機能追加

---

## 10. 参考リンク・先行事例

- [Goldziher/ai-rulez](https://github.com/Goldziher/ai-rulez) — 19+ ツール対応、enterprise 寄り
- [block/ai-rules](https://github.com/block/ai-rules) — シンプル、monorepo対応
- [intellectronica/ruler](https://github.com/intellectronica/ruler) — rule 一元化
- agentsmesh — 全 config 面の一元化を狙う新興
- ScaffoldAI — Web フォーム型 scaffold
- [agents.md](https://agents.md/) — open spec

---

## Appendix A: 用語

- **scaffold**: プロジェクト雛形の生成
- **preset**: stack や用途別のテンプレート組み合わせ
- **canonical source**: 真実の単一情報源
- **frontmatter**: markdown ファイル冒頭の YAML メタデータ
- **ADR**: Architecture Decision Record
- **dogfooding**: 自分のプロダクトを自分で使うこと

## Appendix B: 想定 AGENTS.md テンプレート（v0.1 サンプル）

```markdown
---
audience: agent
updated: 2026-05-19
---

# Agent Instructions

## Project Overview
This is `{{project_name}}`. See [SPEC.md](./SPEC.md) for what and why.

## Before You Start
Read these files in order:
1. [SPEC.md](./SPEC.md) — requirements
2. [ARCHITECTURE.md](./ARCHITECTURE.md) — technical structure
3. [GLOSSARY.md](./GLOSSARY.md) — terminology

## Navigation Matrix
| Task type | Read these files |
|---|---|
| Implement a new feature | SPEC.md, ARCHITECTURE.md |
| Modify UI | UI.md, DESIGN.md (if exists) |
| Change API | API.md, ARCHITECTURE.md |
| Refactor | ARCHITECTURE.md, docs/adr/ |

## Hard Rules
- Update `docs/tasks/current.md` when you start/finish work
- Never commit secrets; reference `.env.example`
- New design decisions go to `docs/adr/` as new entries
- Update GLOSSARY.md when introducing new domain terms

## Stack Notes
{{include stacks/*.md}}

## Testing
{{if testing}}See [docs/testing.md](./docs/testing.md){{end}}
```

---

*このドキュメント自体が aikata が生成する `ARCHITECTURE.md` + `SPEC.md` の合成版のような構造になっており、ある意味 dogfooding の最初のステップです。実装着手時には責務に応じて分割すべきです。*
