---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: agent
---

# samplekata のエージェント指示書

## 1. プロジェクト概要

何を／なぜは [SPEC.md](./SPEC.md) を参照。技術構造は
[ARCHITECTURE.md](./ARCHITECTURE.md)、用語は
[GLOSSARY.md](./GLOSSARY.md) に固定されています。

## 2. 作業を始める前に

次の順に読みます。

1. [README.md](./README.md) — 概要
2. **本ファイル (AGENTS.md)** — 運用ルール
3. [SPEC.md](./SPEC.md) — 要件
4. [ARCHITECTURE.md](./ARCHITECTURE.md) — 技術構造
5. [GLOSSARY.md](./GLOSSARY.md) — 用語
6. [`docs/tasks/current.md`](./docs/tasks/current.md) — 短期の作業状態

## 3. ナビゲーション

| 作業の種類 | まず読むファイル |
|---|---|
| 新機能の実装 | `SPEC.md`, `ARCHITECTURE.md` |
| 公開 API の変更 | `SPEC.md`, `ARCHITECTURE.md`, 関連テスト |
| バグ修正 | `docs/tasks/current.md`, 関連テスト |
| 設計判断の記録 | `docs/adr/` 配下に新規ファイル。書式は [ADR 0001](./docs/adr/0001-record-architecture-decisions.md) |
| 用語の更新 | `GLOSSARY.md` を更新後、旧表記を grep |
| 厄介な問題の調査 | まず `docs/troubleshooting.md` |

## 4. 厳守ルール

- **シークレットをコミットしない。** 代わりに `.env.example` のパターンを参照する。
- **作業の開始時と終了時に [`docs/tasks/current.md`](./docs/tasks/current.md) を更新する。**
- ドキュメント変更のみでない限り、**新規コードにはテストを追加する。**
- **完了を宣言する前に検証する。** プロジェクトにテストやビルドがあれば、
  実行して出力を確認してから変更の完了を宣言する。
- **[Conventional Commits](https://www.conventionalcommits.org/) を使う。**
  種別: feat, fix, docs, style, refactor, test, chore, perf, ci, build。
- **コミットに AI 署名を入れない** — ツールではなく変更内容を記述する。
- **設計判断は [`docs/adr/`](./docs/adr/) に ADR として記録する。**
  書式は [ADR 0001](./docs/adr/0001-record-architecture-decisions.md) に従う。
- **新しいドメイン用語が出てきたら [GLOSSARY.md](./GLOSSARY.md) を更新する。**

## 5. コンテキストの置き場所

次の置き場所を混同しないでください。

| 置き場所 | 寿命 | 用途 |
|---|---|---|
| `AGENTS.md` | 長期・不変 | 常に守るルール |
| `docs/tasks/current.md` | 短期・書き換え可 | 現在進行中の作業状態 |
| `docs/adr/` | 永続 | 設計判断とその理由 |

`docs/tasks/current.md` をバックログやタスク履歴として使わないでください。
引き継ぎ前に、永続化すべき情報は適切な置き場所へ移します。

## 6. 詰まったとき

優先順位は次のとおり。

1. [`docs/troubleshooting.md`](./docs/troubleshooting.md) を確認。
   問題が既知である可能性が高い。
2. [`docs/adr/`](./docs/adr/) で過去の設計判断を確認。
3. 質問を文書化し、`docs/tasks/current.md` に追記してメンテナーに
   surface する。アーキテクチャ上の重要点を黙って推測しない。
