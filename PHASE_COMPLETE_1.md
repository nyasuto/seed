# PHASE_COMPLETE_1 — Game Controller + ティック進行

## 実装した内容

### Task 1-A: GameController — core接続とSnapshot管理
- controller/controller.go: GameController 構造体（SimulationEngine ラップ、Snapshot/AddAction/AdvanceTick API）
- シナリオ JSON から SimulationEngine を初期化、Snapshot を自動構築

### Task 1-B: ティック進行モード（手動/早送り/一時停止）
- controller/tick.go: TogglePause, StartFastForward, StopFastForward, UpdateTick
- Manual/FastForward/Paused/GameOver の4状態遷移
- FastForward 中にゲーム終了で自動停止

### Task 1-C: GameSnapshot から描画データへの変換
- view/mapview.go 拡張: BuildRoomRenderMap で部屋属性に基づく色分けタイル描画
- view/entity.go: EntityRenderer（仙獣/侵入者のスプライト描画、部屋中心座標計算）

### Task 1-D: Top Bar — ステータス表示
- view/ui.go: TopBar（ChiPool、CoreHP バー + 数値、Tick 数）
- view/text.go: DrawText/TextWidth テキスト描画ヘルパー

### Task 1-E: Phase 1 統合と確認
- main.go を GameController ベースに統合
- Space キーで手動ティック進行、F キーで早送り、Escape で早送り停止
- MapView + EntityRenderer + TopBar + Tooltip の全コンポーネント統合
- ゲーム終了時のオーバーレイ表示

## 未解決の課題・技術的負債

- **Ebitengine GLFW テスト問題**: ディスプレイアクセスがない環境で asset/view/input パッケージのテストが GLFW init panic で実行不可。controller パッケージのテストは常にパス。ディスプレイ環境での全テスト実行が必要
- **game/testdata/tutorial.json の重複**: core の組み込みシナリオと同一内容を個別に embed。core の LoadBuiltinScenario への統一は Phase 2 以降で検討
- **BuildRoomRenderMap の毎フレーム呼び出し**: 現在 Draw() で毎フレーム再構築。Cave 変更時のみ再構築する最適化は Phase 2 以降で検討
- **MaxCoreHP の取得方法**: Snapshot に MaxCoreHP が含まれていないため、龍穴部屋の RoomType から動的に計算。Snapshot の拡張を検討すべき

## 次フェーズへの申し送り事項

- **Phase 2: 操作システム**: InputStateMachine + ActionBar で D/C/S/U/W キーの操作モード切替と PlayerAction 生成を実装
- **入力処理の分離**: 現在の main.go 内のキーボード処理を input パッケージに移動し、InputStateMachine と統合
- **DECISIONS.md**: Phase 1 game で新規設計判断は発生せず。D014/D015/D016 は RESOLVED 確認済み

## LESSONS.md から特に重要な知見

- GameController の3つのAPI（Snapshot/AddAction/AdvanceTick）で game と core の境界を明確化
- TopBarData の MaxCoreHP は Snapshot に含まれないため龍穴部屋から動的取得が必要
- Ebitengine v2.9.9 は GLFW init がパッケージ import 時に実行されるため、ディスプレイなし環境ではテスト不可
- `inpututil.IsKeyJustPressed` は長押し連続発火を防げる
