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
> 個別の判断は [`docs/adr/`](./docs/adr/) に。TypeScript 固有規約
> （tsconfig の strict 度、lint、テストランナー選択、ESM/CJS）は
> [`docs/stacks/typescript.md`](./docs/stacks/typescript.md) に。

---

## 1. 実装言語とランタイム

- **TypeScript** 5.x、`tsc`（および／またはバンドラ — 非自明なら ADR で
  記録）でコンパイル。
- ターゲットランタイム: _TODO — Node.js LTS | Bun | Deno | ブラウザ_。
  最低バージョンは `package.json` の `engines` で固定。
- モジュール形式: _TODO — ESM | CJS_。`docs/stacks/typescript.md` に固定。

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
├── .aikata/
│   └── aikata.yaml
├── docs/
│   ├── adr/
│   │   └── 0001-record-architecture-decisions.md
│   ├── stacks/
│   │   └── typescript.md
│   ├── tasks/
│   │   └── current.md
│   ├── troubleshooting.md
│   └── prompts.md
├── src/
│   └── index.ts             # エントリポイント
├── test/                    # src/ と 1:1
├── package.json
├── tsconfig.json
└── (任意) eslint.config.* | .eslintrc.*
```

_TODO: ソースコードのツリーが定まったら拡張する。_

## 3. 主要コンポーネント

_TODO: `src/` 配下の主要フォルダ（ドメイン or レイヤごと）を列挙し、
それぞれ 1 段落で説明する。_

## 4. データフロー

_TODO: データがシステム内でどう動くかを記述。状態 / キャッシュ /
イベントバスの選択は ADR で記録。_

## 5. 依存

_TODO: npm 依存を 1 行の採用理由付きで列挙。ランタイム依存と
devDependencies を区別する。stdlib の薄いラッパーは避ける。_

## 6. エラー処理とログ

_TODO: エラーラップ規約（カスタム Error サブクラス / Result 型 など）、
CLI サーフェスの終了コード、ログレベル、採用ロガーを記述。_

## 7. テスト戦略

- **単体テスト** を `test/` 配下に `src/` と 1:1 で配置。
- **結合テスト** を I/O を触るコードに対して用意。
- CI は `tsc --noEmit` → `eslint .` → テストランナー
  （`docs/stacks/typescript.md` で固定）の順に実行する。

## 8. 配布とリリース

_TODO: 成果物のパッケージ／配布手段を記述（npm publish / Docker /
serverless 等）。_
