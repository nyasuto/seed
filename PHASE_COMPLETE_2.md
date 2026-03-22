# PHASE_COMPLETE_2 — Phase 2: 操作システム

## 実装した内容

Phase 2 では game クライアントの全操作フローを実装した。

### Task 2-A: 操作モードステートマシン
- `input/action.go`: ActionMode 定義（Normal, DigRoom, DigCorridor, Summon, Upgrade）
- `input/state.go`: InputStateMachine — キーボード D/C/S/U/Escape でモード切替

### Task 2-B: Action Bar 描画
- `view/actionbar.go`: 5つのアクションボタン + 3つのティック制御ボタン
- `view/button.go`: 汎用ボタン描画（矩形 + ラベル + ホバー/アクティブ状態）
- D018: input→view 循環依存を CellConverter インターフェースで解消

### Task 2-C: DigRoom フロー
- `input/digroom.go`: 壁セル選択 → 属性選択 → DigRoom PlayerAction 生成
- バリデーション: 壁以外/HardRock/Water の拒否

### Task 2-D: DigCorridor フロー
- `input/corridor.go`: 始点部屋選択 → 終点部屋選択 → DigCorridor PlayerAction 生成
- バリデーション: 部屋セル以外の拒否、同一部屋の拒否

### Task 2-E: SummonBeast フロー
- `input/summon.go`: 部屋選択 → 属性選択 → SummonBeast PlayerAction 生成
- D019: 種族選択ではなく Element 選択方式（core API に忠実）

### Task 2-F: UpgradeRoom フロー + Wait + ティック制御統合
- `input/upgrade.go`: 部屋選択 → UpgradeRoom PlayerAction 生成
- Wait（W キー）: アクションなしでティック進行
- ティック制御: Space（手動）/ F（早送り）/ Escape（停止）

### Task 2-G: 操作フィードバック
- `view/feedback.go`: FeedbackOverlay — エラーメッセージ（赤、3秒フェード）+ モードラベル
- `view/selection.go`: SelectionPanel — 属性選択/種族選択の汎用コンポーネント
- `view/highlight.go`: セルハイライト（有効セル明転、無効セル暗転）

### Task 2-H: Phase 2 統合と確認
- main.go に全フローを統合済み
- キーボードショートカット: D/C/S/U/W/Space/F/Escape 全確認
- `go test ./... -count=1 -race` 全パス
- `go vet ./...` クリーン

## キーボードショートカット一覧

| キー | 機能 |
|------|------|
| D | 掘削モード（DigRoom） |
| C | 通路モード（DigCorridor） |
| S | 召喚モード（SummonBeast） |
| U | 強化モード（UpgradeRoom） |
| W | 待機（Wait — ティック進行） |
| Space | ティック手動進行 |
| F | 早送り開始（10ティック/フレーム） |
| Escape | モードキャンセル / 早送り停止 |

## 未解決の課題

- `game/testdata/tutorial.json` の embed 重複（core の `LoadBuiltinScenario` への統一は Phase 3 以降で検討）
- `BuildRoomRenderMap` の毎フレーム再構築（Phase 3 以降で最適化検討）

## 次フェーズへの申し送り

- Phase 3 はシーン管理（SceneManager）+ UI 強化。main.go の Game 構造体を InGameScene に移行し、タイトル画面/シナリオ選択/結果画面を追加
- Phase 2 で構築した全操作フロー（InputStateMachine + 4つの Flow + FeedbackOverlay）はそのまま InGameScene に統合可能
- D019（SummonBeast の Element 選択方式）は継続有効

## LESSONS.md から特に重要な知見

- `input` と `view` の循環依存は implicit interface で解消するパターンが有効（D018）
- core の API 設計に忠実に従うことで、不要な UI を作らずに済む（D019: 種族選択不要）
- FeedbackOverlay のタイマーはフレームカウントベースが決定論的テストに適している
