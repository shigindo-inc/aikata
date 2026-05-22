---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: agent
---

# samplekata のエージェント指示書

## 1. プロジェクト概要

「何を／なぜ」については [SPEC.md](./SPEC.md) を参照してください。

## 2. 作業を始める前に

次の順に読みます。

1. [README.md](./README.md) — 概要
1. **本ファイル (AGENTS.md)** — 運用ルール
1. [SPEC.md](./SPEC.md) — 要件

## 3. 厳守ルール

- シークレットをコミットしない。代わりに `.env.example` のパターンを参照する。
- 要件が変わったら [SPEC.md](./SPEC.md) を更新する。
- [Conventional Commits](https://www.conventionalcommits.org/) を使う。
- コミットに AI 署名を入れない。

## 4. 詰まったとき

PR の説明欄またはコミットメッセージ本文に質問を書き、
アーキテクチャ上重要な点を黙って推測せず、必ずメンテナーに確認すること。
