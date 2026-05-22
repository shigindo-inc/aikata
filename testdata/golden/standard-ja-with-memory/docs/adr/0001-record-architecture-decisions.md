---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# ADR 0001 — アーキテクチャ判断の記録

- **Status**: Accepted
- **Date**: 2026-05-21
- **Deciders**: samplekata maintainers

## 文脈

アーキテクチャ上重要な判断を低摩擦で記録し、

- 将来のコントリビューター（人・LLM 問わず）が**なぜ**その選択がされたかを理解できる
- 判断は廃止されてもその歴史を失わない

ようにしたい。

重い決定プロセスは導入を妨げる。`docs/` 配下の自由形式ノートでは
発見性が低い。

## 決定

簡易化した Nygard / MADR 風の **Architecture Decision Records (ADR)**
を採用する。

- 1 つの ADR につき 1 件の判断。
- 保存先: `docs/adr/NNNN-kebab-title.md`。`NNNN` は単調増加。
- 必須 frontmatter キー: `project`, `status`, `version`, `updated`,
  `audience`。
- 本文の必須セクション（順序固定）: **文脈**、**決定**、**帰結**。
  任意で **代替案**。
- `Status` の値: `Proposed`, `Accepted`, `Deprecated`,
  `Superseded by ADR-NNNN`。
- `Accepted` になった ADR は**本文を編集しない**。変更は新しい ADR で
  supersede する。

## 帰結

**ポジティブ**:

- 1 件の判断記録に約 10 分。
- grep フレンドリー（`Status: Accepted` で現役判断が一覧できる）。
- supersede された判断も追跡可能。

**ネガティブ**:

- 並行 PR で番号衝突が起きる。緩和策: PR 作者がリベース時に再採番する。
- フォルダが単調増加する。フル履歴のためのトレードオフとして受容。

## 代替案

- **単一 `decisions.md` ログ** — モノリシック化しやすく、判断ごとの
  メタデータが扱いづらい。
- **Wiki / 外部システム** — 「ドキュメントはコードと一緒に住む」
  原則に反する。
