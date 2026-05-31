---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# samplekata

> TypeScript プロジェクトである samplekata を一行で説明する文章を記載してください。

このプロジェクトは
[aikata](https://github.com/shigindo-inc/aikata) の `typescript` プリセットで
雛形を生成しました。

## 最初に読むもの

| 目的 | ドキュメント |
|---|---|
| 何を／なぜ | [SPEC.md](./SPEC.md) |
| どう作るか（技術） | [ARCHITECTURE.md](./ARCHITECTURE.md) |
| 用語 | [GLOSSARY.md](./GLOSSARY.md) |
| エージェント運用ルール | [AGENTS.md](./AGENTS.md) |
| スタック固有ルール | [docs/stacks/typescript.md](./docs/stacks/typescript.md) |

### 設計・決定

- [`docs/adr/`](./docs/adr/) — Architecture Decision Records。

### 運用ノート（頻繁に更新）

- [`docs/tasks/current.md`](./docs/tasks/current.md) — 現在進行中の作業。
- [`docs/troubleshooting.md`](./docs/troubleshooting.md) — 既知の落とし穴。

## クイックスタート

```bash
# 採用するパッケージマネージャに合わせて読み替えてください。
npm install          # または: pnpm install / yarn / bun install
npm run lint
npm test
npm run build
```

## 設定

aikata 自身の設定は [`.aikata/aikata.yaml`](./.aikata/aikata.yaml) に保存します。
samplekata が期待する環境変数は
[`.env.example`](./.env.example) に記述してください。

## ライセンス

(未指定 — 公開前に `LICENSE` ファイルを追加してください)
