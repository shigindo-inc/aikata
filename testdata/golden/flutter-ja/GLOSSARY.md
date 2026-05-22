---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# 用語集

samplekata のドキュメントとソースコード全体で使用する用語。
**(domain)** マークはプロジェクト固有の概念。それ以外は業界標準の
用語で、人間と LLM 双方の曖昧さを減らすため本書で解釈を固定する。

> **なぜこのファイルが重要か**: 語彙を一箇所に固定することで、
> LLM 出力の翻訳ドリフトを減らし、レビュアーが用語の不整合に
> 気付きやすくなる。

---

## A

### ADR — Architecture Decision Record

ひとつのアーキテクチャ判断と、その文脈・帰結を記録する短い markdown
文書。`docs/adr/NNNN-title.md` で保存。書式は
[`docs/adr/0001-record-architecture-decisions.md`](./docs/adr/0001-record-architecture-decisions.md)
に従う。

### agent

本プロジェクトのドキュメントを読み、コードを生成・編集する
LLM ベースのコーディングアシスタント。

---

## B

### BuildContext

widget tree 上の位置を表すハンドル。継承された widget（テーマ、
ナビゲーター、provider など）の検索に使う。`await` の境界をまたいで
`BuildContext` を保持するのは安全でない — `await` の後で `mounted`
を必ず確認すること。

---

## C

### canonical source（規範ソース）

ある情報の**唯一の真実の源**。生成物と乖離した場合は規範ソースが
勝つ。

### Conventional Commits

コミットメッセージの規約（`<type>(<scope>): <subject>`）。本プロジェクトでは必須。
[AGENTS.md](./AGENTS.md) を参照。

---

## F

### frontmatter

markdown ファイル冒頭の `---` で囲まれた YAML ブロック。本プロジェクト
では `project`, `status`, `version`, `updated`, `audience` をクロス
ドキュメントのメタデータとして使う。

---

## N

### null safety

Dart のコンパイル時保証。non-nullable 型は `null` を保持できない。
`?` で nullable を表現し、初期化後に non-null が確定する場合は `?` よりも
`late` を優先する。`!`（force non-null）は理由を述べたコメント無しでは
避ける。

---

## S

### state management

mutable な状態を widget tree 全体で共有する仕組み。samplekata
における選択は ADR で記録し、`docs/stacks/flutter.md` で解説する。

---

## W

### widget

Flutter UI を構成するブロック。widget はイミュータブルな記述であり、
フレームワークが mutable な `Element` / `RenderObject` インスタンスに
具体化する。widget 記述ルールは [`docs/stacks/flutter.md`](./docs/stacks/flutter.md)
を参照。

---

## (プロジェクト固有の用語)

_TODO: samplekata で重要なドメイン用語に置き換える。_
