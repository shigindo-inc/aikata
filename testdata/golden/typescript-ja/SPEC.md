---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# SPEC — samplekata

> 本書では samplekata が**何を**するか、**なぜ**そうするかを述べる。
> **どう**作るかは [ARCHITECTURE.md](./ARCHITECTURE.md)、用語は
> [GLOSSARY.md](./GLOSSARY.md)、TypeScript 固有の規約は
> [docs/stacks/typescript.md](./docs/stacks/typescript.md) を参照。

---

## 1. 目的

### 1.1 一行サマリ

> samplekata は _TODO: 一行のプロダクト説明に置き換える_

### 1.2 解決したい問題

_TODO: 解決したい問題を 2〜3 段落で述べる。_

### 1.3 解決策

_TODO: 高レベルで解決方法を述べる。実装詳細は `ARCHITECTURE.md` に。_

---

## 2. ゴール／非ゴール

### 2.1 ゴール（スコープ内）

- _TODO: 達成すべき主要な成果を列挙_

### 2.2 非ゴール（スコープ外）

- _TODO: スコープクリープ防止のため明示的な非ゴールを列挙_

### 2.3 ユーザー

| 階層 | ペルソナ | ニーズ |
|---|---|---|
| 主要 | _TODO_ | _TODO_ |
| 副次 | _TODO_ | _TODO_ |

### 2.4 ランタイム

- _TODO: 対象ランタイムと最低バージョン（Node.js LTS / Bun / Deno /
  ブラウザ）を記載。モジュール形式（ESM / CJS）は
  `docs/stacks/typescript.md` に固定する。_

---

## 3. 機能要件

_TODO: externally に観測可能なふるまいを列挙。主要サーフェスごとに
サブセクションを置く（CLI コマンド、HTTP ルート、エクスポート関数、
UI 画面など）。_

---

## 4. 非機能要件

- **信頼性**: _TODO_
- **性能**: _TODO_
- **可搬性**: _TODO（対象ランタイム — §2.4 参照）_
- **互換性**: _TODO（最低 TypeScript / Node のバージョン、公開 API の semver 保証）_

---

## 5. 未決定事項

まだ決まっていない決定事項。解決したものは [`docs/adr/`](./docs/adr/)
に移動する。

- _TODO_
