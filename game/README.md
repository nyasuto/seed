# ChaosForge — 風水回廊記

カオスシード（風水回廊記）インスパイアのダンジョン経営シミュレーション GUI クライアント。
Ebitengine ベースの 2D ゲーム。

## ビルド方法

```bash
# game ディレクトリでビルド
cd game
go build -o chaosforge .

# 実行
./chaosforge
```

または、ルートから:

```bash
make all    # core + sim + game の全体ビルド・テスト
```

## テスト

```bash
# headless-safe なテスト（CI 対応）
cd game && go test ./controller/... ./save/... -count=1 -race

# ディスプレイ環境での全テスト
cd game && go test ./... -count=1 -race

# vet
cd game && go vet ./...
```

> **注意**: asset/view/input/scene パッケージは Ebitengine の GLFW 初期化に依存するため、headless 環境（SSH、CI等）ではテスト実行不可。controller と save パッケージは常にテスト可能。

## 操作方法

### シーンフロー

```
タイトル → [新しいゲーム] → シナリオ選択 → InGame → 結果画面
         → [ロード]      → セーブ一覧   → InGame
```

### キーボードショートカット

| キー | 操作 |
|------|------|
| **D** | 部屋掘削モード（壁セルをクリック → 属性選択） |
| **C** | 通路掘削モード（始点部屋クリック → 終点部屋クリック） |
| **S** | 仙獣召喚モード（部屋クリック → 属性選択） |
| **U** | 部屋強化モード（部屋クリック） |
| **W** | 待機（アクションなしで 1 ティック進行） |
| **Space** | ティック進行（pending キューのアクションを実行） |
| **F** | 早送り開始（毎フレーム自動ティック進行） |
| **Escape** | 早送り停止 / 操作モードキャンセル / パネル閉じ |
| **Ctrl+S** | セーブ（`~/.chaosforge/saves/` に保存） |
| **Ctrl+L** | ロード（最新のセーブファイルから復元） |

### マウス操作

- **マップ上ホバー**: セル情報をツールチップ表示
- **マップ上クリック（Normal モード）**: 部屋情報を InfoPanel に表示
- **マップ上クリック（操作モード）**: モードに応じた操作（部屋掘削、通路始点/終点選択等）
- **Action Bar クリック**: 操作モード切替 / ティック制御
- **パネル上クリック**: 属性選択等

## アーキテクチャ

```
game/
├── main.go           # エントリポイント、SceneManager 駆動
├── asset/            # カラーパレット、TilesetProvider、PlaceholderProvider
├── controller/       # GameController（core.SimulationEngine ラッパー）
├── input/            # 入力処理（MouseTracker、ActionMode ステートマシン、操作フロー）
├── view/             # 描画（MapView、EntityRenderer、TopBar、ActionBar、InfoPanel、フィードバック）
├── scene/            # シーン管理（Title、Select、InGame、Load、Result）
├── save/             # セーブ/ロード（チェックポイント、ゲーム設定）
└── docs/PRD.md       # プロダクト要件定義
```

依存方向: `game → core`（sim とは独立）

## シナリオ

| シナリオ | 難易度 | 説明 |
|---------|--------|------|
| チュートリアル | easy | 基本操作を学ぶための簡単なシナリオ |
| 標準シナリオ | normal | 中規模マップでの本格的な洞窟経営シナリオ |

## セーブデータ

- セーブファイル: `~/.chaosforge/saves/save_YYYYMMDD_HHMMSS.json`
- 設定ファイル: `~/.chaosforge/config.json`
