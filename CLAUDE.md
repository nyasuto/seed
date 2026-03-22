# CLAUDE.md — chaosseed コーディング規約

## プロジェクト概要

カオスシード（風水回廊記）インスパイアのダンジョン経営シミュレーション。
Go モノレポ構成、決定論的設計。

### リポジトリ構成

```
seed/
├── go.work              # Go Workspace
├── core/                # コアメカニクスエンジン（ライブラリ）
├── sim/                 # CLIシミュレーター
├── game/                # Ebitengine GUIクライアント（将来）
├── DECISIONS.md         # プロジェクト横断の設計判断
├── HANDOFF.md           # プロジェクト横断の引き継ぎ文書
└── LESSONS.md           # プロジェクト横断の知見
```

### レイヤーの役割と依存方向

```
game → sim → core
       ↓
      core
```

- **core**: 純粋なゲームロジックライブラリ。sim と game から import される
- **sim**: CLI ツール。core を import し、ターミナル UI・バッチ実行・メトリクスを提供
- **game**: GUI クライアント。core を import し、Ebitengine で描画

## Go バージョン

- Go 1.26 以上

## 外部ライブラリポリシー

レイヤーごとに依存ポリシーが異なる。

| レイヤー | ポリシー | 理由 |
|----------|---------|------|
| **core** | **外部依存ゼロ（標準ライブラリのみ）** | 全レイヤーの基盤。依存を入れると全体に波及する |
| **sim** | **必要に応じて厳選** | ターミナルUI系のみ許可。それ以外は標準ライブラリ |
| **game** | **ゲームエンジン + 必要なライブラリ** | GUI は外部ライブラリなしでは現実的でない |

### sim で許可するライブラリ

| 用途 | 候補 | 備考 |
|------|------|------|
| ターミナルUI | bubbletea / lipgloss（Charm系）または tcell / tview | Human Mode のプレイ体験に直結 |

### sim で使わないもの（標準ライブラリで十分）

- テストフレームワーク（testify 等）→ 標準 testing で十分
- ログライブラリ → log/slog で十分
- JSON ライブラリ → encoding/json で十分
- CLI パーサー（cobra 等）→ 標準 flag で十分
- HTTP ライブラリ → sim に HTTP 機能は不要

### game で許可するライブラリ

| 用途 | ライブラリ | 備考 |
|------|-----------|------|
| ゲームエンジン | Ebitengine | Go 製 2D ゲームエンジン |
| その他 | game 開発時に判断 | |

### ライブラリ追加の判断基準

外部ライブラリを追加する前に以下を確認する：
1. 標準ライブラリで実現できないか？（できるなら使わない）
2. 追加先は core ではないか？（core への追加は禁止）
3. メンテナンスされているか？（直近1年以内のリリースがあるか）
4. 依存の連鎖が深くないか？（transitive dependency が少ないか）

## コーディング規約

### パッケージ設計

#### core のパッケージ依存方向

```
types ← world ← fengshui ← senju ← invasion ← economy ← scenario ← simulation
```

- `types` は他のどのパッケージにも依存しない
- 循環依存は絶対に作らない
- 各パッケージは自身のドメインに閉じたロジックのみ持つ

#### sim のパッケージ依存方向

```
server ← adapter/human, adapter/ai, adapter/batch
server ← metrics
balance ← adapter/batch, metrics
cmd/chaosseed-sim ← 全パッケージ
```

- adapter 間は互いに依存しない
- metrics は server から呼ばれるが、adapter には依存しない
- balance は batch と metrics を組み合わせる

### 命名規則

- 構造体: PascalCase（`Cave`, `Room`, `DragonVein`）
- インターフェース: 動詞 + er / 形容詞（`RNG`, `Evaluator`, `ActionProvider`）
- ファイル名: snake_case（`room_type.go`, `corridor_builder.go`）
- テストファイル: `*_test.go`（同一パッケージ内に配置）
- 定数: PascalCase（`Wood`, `North`）、グループは iota
- 壊れるサインメトリクス: `B` プレフィックス + 2桁番号（`B01`, `B02`, ...）

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
- Batch Mode の並列実行でも各ゲームに独立した RNG を割り当て、決定論性を保つ

### テスト

- テストカバレッジ目標: 80% 以上
- テーブル駆動テストを推奨
- テスト名は `Test<関数名>_<条件>` 形式（例: `TestCanPlaceRoom_OverlapRejected`）
- 外部テストフレームワークは使わない。標準 testing パッケージのみ
- ファイルI/Oテストは `t.TempDir()` を使う
- JSON定義ファイルのテストは `testdata/` ディレクトリまたは embed で埋め込み
- Human Mode: スクリプト化された入力（io.Reader）で E2E テスト
- AI Mode: パイプ経由の E2E テスト
- Batch Mode: 少数ゲーム（10〜100ゲーム）での統合テスト

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
# ルートから全体ビルド・テスト
make all          # core + sim のビルド・テスト

# 推奨: vet + lint + test-race を一括実行
make check        # ルートから全体チェック

# 個別パッケージ
cd core && make check
cd sim && make check

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

### core パッケージ

```
<package>/
├── doc.go              # パッケージドキュメント
├── <domain>.go         # 主要な型と基本操作
├── <domain>_ops.go     # 複合操作・ビジネスロジック
├── <domain>_test.go    # ユニットテスト
├── integration_test.go # 統合テスト（必要な場合）
└── testdata/           # テスト用データファイル
```

### sim パッケージ

```
sim/
├── server/             # Game Server（core の SimulationRunner をラップ）
├── adapter/
│   ├── human/          # Human Mode（ターミナルUI）
│   ├── ai/             # AI Mode（JSON Lines I/O）
│   └── batch/          # Batch Mode（ヘッドレス統計）
├── render/             # ASCII描画
├── metrics/            # 壊れるサイン検出（B01〜B11）
├── balance/            # バランス調整ダッシュボード
└── cmd/chaosseed-sim/  # エントリポイント
```

## コミット規約

- 1タスク = 1コミット
- Conventional Commits 形式（`feat:`, `fix:`, `ci:`, `refactor:`, `test:`, `docs:`）
- スコープで対象を明示: `feat(core):`, `fix(sim):`, `docs(root):`
- リリース時にタグ:
  - core: `core/v1.0.0`, `core/v1.1.0`
  - sim: `sim/v1.0.0`
  - セマンティックバージョニング

## やってはいけないこと

- **core に外部ライブラリを追加する**（標準ライブラリのみ）
- **sim に用途不明のライブラリを追加する**（上記の許可リスト以外は事前に判断基準を確認）
- グローバル変数の使用（テスト以外）
- init() 関数の使用
- `any` の多用（型安全を保つ）
- 未使用のコードを残す
- core と sim の間に循環依存を作る（sim → core の一方向のみ）