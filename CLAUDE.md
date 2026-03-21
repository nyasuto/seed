# CLAUDE.md — chaosseed-core コーディング規約

## プロジェクト概要

カオスシード（風水回廊記）インスパイアのダンジョン経営シミュレーション・コアエンジン。
純Go、ゲームエンジン非依存、決定論的設計。

## Go バージョン

- Go 1.25 以上

## コーディング規約

### パッケージ設計

- パッケージ間の依存方向: `types` ← `world` ← `fengshui` ← `senju` ← `invasion` ← `economy` ← `scenario` ← `simulation`
- `types` は他のどのパッケージにも依存しない
- 循環依存は絶対に作らない
- 各パッケージは自身のドメインに閉じたロジックのみ持つ

### 命名規則

- 構造体: PascalCase（`Cave`, `Room`, `DragonVein`）
- インターフェース: 動詞 + er / 形容詞（`RNG`, `Evaluator`, `Serializable`）
- ファイル名: snake_case（`room_type.go`, `corridor_builder.go`）
- テストファイル: `*_test.go`（同一パッケージ内に配置）
- 定数: PascalCase（`Wood`, `North`）、グループは iota

### エラーハンドリング

- 可能な限り error を返す。panic は使わない
- エラーメッセージは小文字始まり、ピリオドなし（Go標準）
- パッケージ固有のエラーは `var ErrXxx = errors.New(...)` で定義
- エラーのラップは `fmt.Errorf("context: %w", err)`

### 決定論的設計（重要）

- `math/rand` のグローバル関数を直接使わない
- 乱数が必要な関数は必ず `types.RNG` を引数に取る
- テストでは `testutil.FixedRNG` または `testutil.NewTestRNG(seed)` を使う
- 同じ RNG seed → 同じ結果を保証する

### テスト

- テストカバレッジ目標: 80% 以上
- テーブル駆動テストを推奨
- テスト名は `Test<関数名>_<条件>` 形式（例: `TestCanPlaceRoom_OverlapRejected`）
- 外部依存なし。ファイルI/Oテストは `t.TempDir()` を使う
- JSON定義ファイルのテストは `testdata/` ディレクトリまたは embed で埋め込み

### ドキュメント

- エクスポートされる型・関数には必ず GoDoc コメント
- パッケージコメントは各パッケージの `doc.go` に記述
- 複雑なアルゴリズムはコード内コメントで意図を説明

### データ定義

- ゲームデータ（部屋タイプ、仙獣種族等）は JSON ファイルで外出し
- Go の embed パッケージでバイナリに埋め込む（`//go:embed`）
- スキーマ変更時は旧データとの互換性を考慮（バージョンフィールド）

## ビルドとテスト

```bash
# テスト実行
go test ./...

# race detector 付き
go test -race ./...

# vet
go vet ./...

# カバレッジ
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## ファイル構成テンプレート

```
<package>/
├── doc.go              # パッケージドキュメント
├── <domain>.go         # 主要な型と基本操作
├── <domain>_ops.go     # 複合操作・ビジネスロジック
├── <domain>_test.go    # ユニットテスト
├── integration_test.go # 統合テスト（必要な場合）
└── testdata/           # テスト用データファイル
```

## コミット規約

- 1タスク = 1コミット
- メッセージ: `phase1-A: add Pos and Direction types`
- フェーズ完了時にタグ: `v0.1.0`（Phase 1）, `v0.2.0`（Phase 2）...

## やってはいけないこと

- 外部ライブラリへの依存追加（標準ライブラリのみ）
- グローバル変数の使用（テスト以外）
- init() 関数の使用
- interface{} / any の多用（型安全を保つ）
- 未使用のコードを残す