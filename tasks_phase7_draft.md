# Phase 7 ドラフト — 統合シミュレーション（simulation/）

> PRD Phase 7: 全システムを統合したゲームループ。
> ティックベースのメインループ、全サブシステムの更新順序管理、状態のスナップショットと巻き戻し、CLIシミュレーター向けインターフェース。
> D002準拠: 3原則をシミュレーションレベルで検証可能にする。
> D011準拠: EventCommand の実際の状態変更適用はこのフェーズで実装する。

## 設計方針メモ

**simulation パッケージの責務:**
- 全サブシステム（world, fengshui, senju, invasion, economy, scenario）を正しい順序でティック更新する
- EventCommand（D011）を受け取り、実際のゲーム状態に適用する
- ゲーム状態のスナップショット保存・巻き戻しを提供する
- CLIシミュレーター（chaosseed-sim）向けの高レベルAPIを公開する

**パッケージ依存方向:**
- `simulation` は `scenario`, `economy`, `invasion`, `senju`, `fengshui`, `world`, `types` に依存する
- 他のどのパッケージも `simulation` に依存しない（依存ツリーの頂点）

**更新順序（毎ティック）:**
1. 風水エンジン更新（ChiFlowEngine.Tick）— 気の供給・伝播・減衰
2. 仙獣成長更新（GrowthEngine.Tick）— 気を消費して成長
3. 仙獣行動AI更新（BehaviorEngine.Tick）— 移動・戦闘態勢
4. 侵入エンジン更新（InvasionEngine.Tick）— 侵入者の移動・戦闘・CoreHPダメージ
5. 経済エンジン更新（EconomyEngine.Tick）— 供給・維持費・トランザクション
6. イベントエンジン更新（EventEngine.Tick）— 条件評価 → EventCommand 生成
7. コマンド適用（CommandExecutor.Apply）— EventCommand を実際の状態に反映
8. 勝利/敗北条件評価 — ゲーム終了判定

## Phase 7-A: シミュレーション基本型定義（simulation/）

- [ ] `simulation/doc.go`: パッケージドキュメント。全サブシステムを統合したゲームループを管理するパッケージ
- [ ] `simulation/state.go`: GameState 構造体（Tick types.Tick, Cave *world.Cave, ChiEngine *fengshui.ChiFlowEngine, GrowthEngine *senju.GrowthEngine, BehaviorEngine *senju.BehaviorEngine, InvasionEngine *invasion.InvasionEngine, EconomyEngine *economy.EconomyEngine, EventEngine *scenario.EventEngine, Scenario *scenario.Scenario, Progress *scenario.ScenarioProgress, RNG types.RNG）
- [ ] `simulation/state.go`: GameStatus 型（iota: StatusRunning, StatusWon, StatusLost）。ゲームの進行状態
- [ ] `simulation/state.go`: GameResult 構造体（Status GameStatus, FinalTick types.Tick, Reason string）。ゲーム終了時の結果情報
- [ ] `simulation/state_test.go`: GameState の初期化テスト、GameStatus の文字列変換テスト

## Phase 7-B: コマンド実行器（simulation/）

- [ ] `simulation/executor.go`: CommandExecutor 構造体。EventCommand を受け取り GameState に適用する
- [ ] `simulation/executor.go`: Apply(state *GameState, cmds []scenario.EventCommand) error メソッド。各コマンド型を判別し適切な状態変更を実行:
  - SpawnWaveCommand → InvasionEngine に波を追加
  - ModifyChiCommand → EconomyEngine の ChiPool に加減算
  - ModifyConstraintCommand → Progress の制約を更新
  - MessageCommand → ログに記録（状態変更なし）
- [ ] `simulation/executor_test.go`: 各コマンド型の適用テスト、複数コマンドの順序実行テスト、不明なコマンド型のエラーテスト

## Phase 7-C: メインループ（simulation/）

- [ ] `simulation/engine.go`: SimulationEngine 構造体（State *GameState, Executor *CommandExecutor, TickLog []TickRecord）
- [ ] `simulation/engine.go`: NewSimulationEngine(scenario *scenario.Scenario, rng types.RNG) (*SimulationEngine, error) コンストラクタ。シナリオの InitialState から GameState を構築
- [ ] `simulation/engine.go`: TickRecord 構造体（Tick types.Tick, Commands []scenario.EventCommand, Events []string）。1ティック分のログ
- [ ] `simulation/engine.go`: Step() (GameResult, error) メソッド。1ティック分の更新を実行（設計方針メモの更新順序に従う）。ゲーム終了時は GameResult を返す
- [ ] `simulation/engine.go`: Run(maxTicks types.Tick) (GameResult, error) メソッド。Step を繰り返しゲーム終了または maxTicks に達するまで実行
- [ ] `simulation/engine_test.go`: 1ティック実行テスト（全サブシステムが正しい順序で更新されること）、勝利条件到達テスト、敗北条件到達テスト（CoreHP=0）、maxTicks 制限テスト

## Phase 7-D: GameSnapshot 構築（simulation/）

- [ ] `simulation/snapshot.go`: BuildSnapshot(state *GameState) scenario.GameSnapshot 関数。GameState から読み取り専用の GameSnapshot を構築し、EventEngine と条件評価器に提供する
- [ ] `simulation/snapshot_test.go`: GameState の各フィールドが正しく GameSnapshot にマッピングされるテスト

## Phase 7-E: 状態スナップショットと巻き戻し（simulation/）

- [ ] `simulation/checkpoint.go`: Checkpoint 構造体（Tick types.Tick, CaveData []byte, ProgressData []byte, ChiPoolData []byte）。シリアライズ済みの状態スナップショット
- [ ] `simulation/checkpoint.go`: CreateCheckpoint(state *GameState) (*Checkpoint, error) 関数。現在の状態をシリアライズしてスナップショットを作成
- [ ] `simulation/checkpoint.go`: RestoreCheckpoint(checkpoint *Checkpoint, scenario *scenario.Scenario, rng types.RNG) (*GameState, error) 関数。スナップショットから状態を復元
- [ ] `simulation/checkpoint_test.go`: 保存→復元の往復テスト（復元後の状態が保存時と一致すること）、複数チェックポイントの管理テスト

## Phase 7-F: CLIシミュレーター向けインターフェース（simulation/）

- [ ] `simulation/runner.go`: SimulationRunner 構造体。高レベルAPIとしてシナリオのロード→実行→結果出力を提供
- [ ] `simulation/runner.go`: RunScenario(scenarioJSON []byte, seed int64) (*RunResult, error) メソッド。JSON からシナリオをロードし、RNG seed で実行
- [ ] `simulation/runner.go`: RunResult 構造体（GameResult GameResult, TickCount int, TickLog []TickRecord, Statistics RunStatistics）
- [ ] `simulation/runner.go`: RunStatistics 構造体（PeakChi float64, TotalWavesDefeated int, FinalFengShuiScore float64, BeastEvolutions int, TotalDamageDealt int, TotalDamageReceived int）
- [ ] `simulation/runner_test.go`: シナリオJSON→実行→結果取得のエンドツーエンドテスト、同一seed で同一結果の決定論性テスト

## Phase 7-G: 統合テスト（simulation/）

- [ ] `simulation/integration_test.go`: フルゲームシナリオ統合テスト:
  - チュートリアルシナリオ（scenario/testdata のサンプル利用）をロード
  - SimulationEngine で最大500ティック実行
  - 全サブシステムが連携動作すること（風水→成長→行動→侵入→経済→イベント）
  - 勝利または敗北で正常終了すること
  - TickLog に全ティックの記録があること
  - 同一 RNG seed で2回実行し、同一の GameResult になること（決定論性検証）
- [ ] `simulation/integration_test.go`: ストレステスト:
  - 大規模Cave（64x64）、部屋10個、仙獣8体、侵入波10回のシナリオ
  - 1000ティック以内に完了すること（無限ループ検出）
  - メモリリークなし（チェックポイント作成→復元を繰り返す）
- [ ] `simulation/integration_test.go`: チェックポイント復元テスト:
  - 100ティック実行→チェックポイント保存→さらに100ティック実行
  - チェックポイントから復元→同じ100ティック実行
  - 復元後の実行結果が元の実行と一致すること（決定論性 + シリアライズ正確性）
- [ ] `go vet ./...` と `go test -race ./...` がクリーンに通ることを確認
- [ ] Phase 7 完了。DECISIONS.md 更新（必要な設計判断を記録）。全フェーズ完了を確認し、v1.0.0 タグの準備。
