---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-24
audience: [human, agent]
---

# Japanese Users

このページは、日本語で aikata を使いたい人向けの入口です。
詳細な仕様・設計・貢献ルールの正本は英語ドキュメントです。

## まず知ること

aikata は、AI コーディング時代のプロジェクトに必要な Markdown
ドキュメントと AI ツール別設定ファイルを生成する軽量 CLI です。

リポジトリ自体の正本ドキュメントは英語です。一方、aikata が生成する
利用者プロジェクトのドキュメントは `--lang ja` で日本語にできます。

```bash
aikata init my-app --preset standard --lang ja --no-interactive
```

Flutter / TypeScript プリセットでも同じ指定が使えます。

```bash
aikata init my-flutter-app --preset flutter --lang ja --no-interactive
aikata init my-ts-app --preset typescript --lang ja --no-interactive
```

## 読む順番

| 知りたいこと | ドキュメント |
|---|---|
| 概要と最短手順 | [README.md](../README.md) |
| 何を解決するか | [SPEC.md](../SPEC.md) |
| どう実装しているか | [ARCHITECTURE.md](../ARCHITECTURE.md) |
| 今後の予定 | [ROADMAP.md](../ROADMAP.md) |
| 用語 | [GLOSSARY.md](../GLOSSARY.md) |
| エージェント / contributor ルール | [AGENTS.md](../AGENTS.md) |

## 言語方針

- aikata リポジトリの正本ドキュメントは英語です。
- 日本語利用者向けドキュメントは、正本の全文翻訳ではなく入口・補助です。
- `--lang ja` は、生成されるプロジェクトドキュメントの言語を指定します。
- CLI の表示言語、issue / discussion で使う言語、将来の翻訳補助機能は
  `--lang` とは別の設計対象です。

この方針は
[ADR 0006](./adr/0006-locale-and-japanese-documentation-policy.md)
で記録しています。

## 日本語での問い合わせ

日本語での issue / discussion は歓迎します。ただし、設計判断や
長期的に参照されるドキュメントは、必要に応じて英語の正本へ反映します。

英語の文面作成が負担な場合は、日本語で背景・期待する動作・困っている点を
具体的に書いてください。maintainer 側で英語の issue / PR 文面に整理できます。
