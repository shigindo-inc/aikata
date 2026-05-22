---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# ARCHITECTURE — どう作るか

> 本書は samplekata の**作り方**を説明する。
> **何を／なぜ**は [SPEC.md](./SPEC.md) を参照。
> 個別の判断は [`docs/adr/`](./docs/adr/) に。Flutter 固有規約
> （lint、状態管理、build_runner、null safety スタンス）は
> [`docs/stacks/flutter.md`](./docs/stacks/flutter.md) に。

---

## 1. 実装言語とランタイム

- **Dart** 3.x + **Flutter** 3.x（チャンネル: _TODO — stable | beta_）。
- 最低 SDK 制約は `pubspec.yaml` に固定。

## 2. リポジトリ構成

```
samplekata/
├── README.md
├── AGENTS.md
├── SPEC.md
├── ARCHITECTURE.md
├── GLOSSARY.md
├── .env.example
├── .gitignore
├── .ai/
│   └── aikata.yaml
├── docs/
│   ├── adr/
│   │   └── 0001-record-architecture-decisions.md
│   ├── stacks/
│   │   └── flutter.md
│   ├── tasks/
│   │   └── current.md
│   ├── troubleshooting.md
│   └── prompts.md
├── lib/
│   └── main.dart            # アプリのエントリ、widget tree のルート
├── test/
│   └── …                    # lib/ と 1:1
├── pubspec.yaml
├── analysis_options.yaml    # lint（docs/stacks/flutter.md 参照）
└── (プラットフォームディレクトリ: ios/, android/, web/, macos/, linux/, windows/)
```

_TODO: ソースコードのツリーが定まったら拡張する。_

## 3. 主要コンポーネント

_TODO: `lib/` 配下の主要機能フォルダ（画面・ドメインごと）を列挙し、
それぞれ 1 段落で説明する。_

## 4. データフロー

_TODO: データがアプリ内をどう流れるかを記述。状態管理（Provider /
Riverpod / Bloc / 他）の選択は ADR で確定、根拠は
`docs/stacks/flutter.md` に。_

## 5. 依存

_TODO: pub.dev のパッケージをそれぞれ 1 行の採用理由付きで列挙。
flutter.dev 公式や長期メンテのパッケージを優先する。
標準ライブラリの薄いラッパーは避ける。_

## 6. エラー処理とログ

_TODO: エラーラップの規約（Result / Either / Exception）、
クラッシュポリシー、採用するロガーを記述。_

## 7. テスト戦略

- **単体テスト** を `test/` 配下に `lib/` と 1:1 で配置。
- **Widget tests** を `lib/` で公開する全カスタム widget に対して用意。
- **Golden tests** をピクセルクリティカルな UI に対して用意（`flutter_test` の `matchesGoldenFile`）。
- **Integration tests** を `integration_test/` 配下で複数画面フローに対して用意。
- CI は `flutter analyze` → `flutter test`（必要に応じて
  `flutter test integration_test/`）を実行する。

## 8. 配布とリリース

_TODO: 成果物のパッケージ／配布手段を記述（TestFlight / Google Play
内部 / Web ホスティング 等）。_
