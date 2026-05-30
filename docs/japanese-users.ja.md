---
project: aikata
status: draft
version: 0.0.1
updated: 2026-05-28
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
aikata init my-app --scope standard --lang ja --no-interactive
```

## インストール

**aikata の利用に Go のセットアップは不要です。** Go が必要なのは
ソースから入れる場合だけです。

- **推奨（Go 不要・手動）**: [Releases ページ](https://github.com/shigindo-inc/aikata/releases/latest)
  から OS / アーキテクチャに合った tar / zip をダウンロード →
  `checksums.txt` で SHA-256 を検証 → `aikata` バイナリを PATH の通った
  ディレクトリ（例: `$HOME/.local/bin/`）に配置。詳細手順は英語 README
  の "Install" セクションに表でまとめてあります。
- **convenience（Linux / macOS, v0.2.1 以降）**: 1 行 install スクリプト
  で OS / arch 検出・ダウンロード・SHA-256 検証・配置までまとめて実行：

  ```bash
  curl -fsSL https://raw.githubusercontent.com/shigindo-inc/aikata/main/scripts/install.sh | sh
  ```

  バージョンを固定したい場合は `AIKATA_VERSION=v0.2.1 sh` を pipe 先に
  指定してください。配置先は `AIKATA_INSTALL_DIR` で変更できます。
  Windows は手動ダウンロード経路をご利用ください。
- **Go ユーザー向け**: `go install github.com/shigindo-inc/aikata/cmd/aikata@latest`
  （Go 1.21 以上）。`$(go env GOPATH)/bin` が PATH に入っている必要が
  あります。
- **今後の予定**: Homebrew tap と `npx aikata` は v0.9.x に延期しました。
  [ROADMAP.md](../ROADMAP.md) を参照してください。

Flutter / TypeScript プリセットでも同じ指定が使えます。

```bash
aikata init my-flutter-app --scope standard --stack flutter --lang ja --no-interactive
aikata init my-ts-app --scope standard --stack typescript --lang ja --no-interactive
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
