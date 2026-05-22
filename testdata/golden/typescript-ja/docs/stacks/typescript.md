---
project: samplekata
status: draft
version: 0.0.1
updated: 2026-05-21
audience: [human, agent]
---

# TypeScript — スタック固有ルール

> samplekata が TypeScript プロジェクトであることに由来する規約。
> [AGENTS.md](../../AGENTS.md) の不変則に *追加* されるルール。
> 衝突した場合は AGENTS.md が勝つ。

---

## 1. tsconfig

- `strict: true`（および strict ファミリのフラグはすべて ON のまま）。
  strict ファミリのいずれかを無効化するには ADR が必要。
- `noUncheckedIndexedAccess: true` — 配列／レコードのインデックス
  アクセスを `T | undefined` にする。
- `exactOptionalPropertyTypes: true` — プロパティ未指定と `undefined` を
  区別する。
- `target` と `module` はランタイムに合わせる（§2 / §3 参照）。
- `incremental: true` で再ビルドを速く。`.tsbuildinfo` は gitignore 済み。

## 2. モジュール形式（ESM vs CJS）

- _TODO: 採用方式をここに固定し、それを決めた ADR にリンクする。_
- ファイル拡張子:
  - ESM: `moduleResolution: nodenext` の場合、import 指定子に明示的に
    `.js` / `.ts` を付ける。
  - CJS: 拡張子無しで OK。混在させない。
- 1 つのパッケージから両形式を同時出荷するには ADR を要する。

## 3. ランタイム

- _TODO: Node.js LTS / Bun / Deno / ブラウザのいずれかを固定。_
- `package.json` の `engines` が最低バージョンを強制する。

## 4. パッケージマネージャ

- _TODO: npm / pnpm / yarn / bun を固定。lockfile をコミット。CI は
  `<pm> ci`（または同等の再現性ある install）を実行。_
- 1 プロジェクトで複数のパッケージマネージャを混用しない。

## 5. Lint

- ESLint は最低でも `@typescript-eslint` の recommended-type-checked を使う。
  プロジェクト固有の追加は `eslint.config.*` に 1 行コメントで根拠を添えて
  書く。
- `eslint .` は**警告 0** でなければコミットしない。CI が強制する。
- フォーマットは `prettier`（または `dprint`）を使い、空白を手調整しない。

## 6. 型規律

- **`any` を使わない**。やむを得ない場合はインラインコメントで理由を述べる。
  `unknown` を優先し、narrowing で絞る。
- **`as` キャストを避ける**。ユーザー定義型ガードまたは `satisfies` を優先。
- **`!` non-null assertion は避ける**。呼び出し側で narrowing できない場合に
  限って使用し、各使用箇所にコメントを添える。
- 引数が変更されない場合は `readonly` を優先する。関数・クラス境界で
  immutability を明示する。
- 型のみの import には `import type` を使い、emit を tree-shake-friendly に保つ。

## 7. テストランナー

- _TODO: **vitest** または **jest** を固定し、ADR にリンクする。
  ADR 無しでランナーを混在させない。_
- テストは `test/` 配下に `src/` と 1:1 で配置。
- 作業完了の宣言前にテストコマンドを実行する。

## 8. エラー

- 新しいエラー分類のために `Error` をサブクラス化する。生の文字列や
  オブジェクトリテラルを throw しない。
- 公開サーフェスは、呼び出し側が期待しうるエラーを文書化する。
- 元のスタックを失わないよう、エラーチェインには `cause`（ES2022）を使う。

## 9. 非同期

- `Promise` を非同期の単位とする。生の `.then` チェーンより
  `async`/`await` を優先する。
- 非テストコードで promise を浮かせない — `await` するか、明示的に
  「コメントを添えて fire-and-forget」のいずれかを選ぶ。
- 早期に reject し、空の `catch` でエラーを握りつぶさない。

## 10. 本書を改訂するタイミング

- 状態・IO・ビルドの挙動を大きく変える依存を追加 → セクションを追加。
- 単一の ADR に収まらないチーム横断的な好み → ここに追加。
- 規定したルールが誤りと判明 → 削除し、コミットメッセージで理由を述べる。
