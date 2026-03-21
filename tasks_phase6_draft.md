# Phase 6 ドラフト — シナリオシステム（scenario/）

> PRD Phase 6: ゲームの進行管理、イベント、勝利条件。
> D002準拠: 3原則（不完全性の強制・時間圧力・トレードオフの連続）をシナリオレベルで統合する。
> 延期タスクの消化: 仙獣進化システム、仙獣敗北後処理をこのフェーズで実装する。

## 設計方針メモ

**シナリオシステムの責務:**
- ゲーム1回分の「ルール」を定義する（初期状態、勝利/敗北条件、イベント、侵入スケジュール）
- 各サブシステム（world, fengshui, senju, invasion, economy）を束ねるオーケストレーションは Phase 7 の simulation パッケージが担う
- scenario パッケージは「何をするか」の定義のみ。「どう動かすか」は simulation が担当

**パッケージ依存方向:**
- `scenario` は `economy`, `invasion`, `senju`, `fengshui`, `world`, `types` に依存する
- `simulation`（Phase 7）が `scenario` に依存する

## Phase 6-A: シナリオ基本型定義（scenario/）

- [ ] `scenario/doc.go`: パッケージドキュメント。シナリオ定義とゲーム進行ルールを管理するパッケージ
- [ ] `scenario/scenario.go`: Scenario 構造体（ID string, Name string, Description string, InitialState InitialState, WinConditions []Condition, LoseConditions []Condition, WaveSchedule []WaveScheduleEntry, Events []EventTrigger, Constraints ScenarioConstraints）
- [ ] `scenario/scenario.go`: InitialState 構造体（CaveWidth int, CaveHeight int, PrebuiltRooms []RoomPlacement, DragonVeinPaths [][]types.Pos, StartingChi float64, StartingBeasts []BeastPlacement）。RoomPlacement 構造体（RoomTypeID string, Pos types.Pos, Width int, Height int）。BeastPlacement 構造体（SpeciesID string, RoomID int）
- [ ] `scenario/scenario.go`: ScenarioConstraints 構造体（MaxRooms int, MaxBeasts int, ForbiddenRoomTypes []string, MaxTicks types.Tick）。ゲーム内の制約条件
- [ ] `scenario/scenario_test.go`: Scenario 構造体の基本テスト、InitialState のバリデーションテスト

## Phase 6-B: 勝利/敗北条件（scenario/）

- [ ] `scenario/condition.go`: Condition インターフェース（Evaluate(state GameState) bool, Description() string）。GameState 構造体（Tick types.Tick, Cave *world.Cave, ChiPool *economy.ChiPool, Beasts int, DefeatedWaves int, TotalWaves int, CoreHP int, FengShuiScore float64）
- [ ] `scenario/condition.go`: 具体的な条件実装:
  - SurviveUntilCondition: 指定ティックまで生存（CoreHP > 0）
  - DefeatAllWavesCondition: 全侵入波を撃退
  - FengShuiScoreCondition: 風水スコアが閾値以上に到達
  - ChiPoolCondition: ChiPool残高が閾値以上に到達
  - CoreDestroyedCondition（敗北）: 龍穴のHP が 0 以下
  - AllBeastsDefeatedCondition（敗北）: 全仙獣が敗北状態
  - BankruptCondition（敗北）: ChiPool が 0 で N ティック連続赤字
- [ ] `scenario/condition_test.go`: 各条件の Evaluate テスト、複数条件の AND/OR 組み合わせテスト

## Phase 6-C: イベントトリガーシステム（scenario/）

- [ ] `scenario/event.go`: EventTrigger 構造体（ID string, Condition Condition, Action EventAction, OneShot bool, Fired bool）。Condition が true になったとき Action を実行。OneShot なら1回のみ
- [ ] `scenario/event.go`: EventAction インターフェース（Execute(state *MutableGameState) []EventResult）。MutableGameState: 変更可能なゲーム状態へのアクセス
- [ ] `scenario/event.go`: 具体的なアクション実装:
  - SpawnWaveAction: 追加の侵入波を発生させる
  - ModifyChiAction: ChiPool に加算/減算
  - ModifyConstraintAction: 制約条件を変更（部屋数上限を増やす等）
  - MessageAction: メッセージを生成（UIレイヤーへの通知用）
- [ ] `scenario/event.go`: EventResult 構造体（Type EventResultType, Description string, Data any）。イベント実行結果のログ
- [ ] `scenario/event_test.go`: OneShot イベントの1回実行テスト、条件未達で実行されないテスト、SpawnWaveAction テスト、複数イベントの順序実行テスト

## Phase 6-D: 動的侵入波スケジュール（scenario/）

- [ ] `scenario/wave_schedule.go`: WaveScheduleEntry 構造体（TriggerTick types.Tick, Difficulty float64, MinInvaders int, MaxInvaders int, PreferredClasses []string, PreferredGoals []string）。Phase 4の WaveConfig を拡張し、クラス・目標の指定を可能にする
- [ ] `scenario/wave_schedule.go`: DynamicWaveScheduler 構造体。GenerateSchedule(scenario, rng) []invasion.WaveConfig — シナリオ定義からPhase 4の WaveConfig に変換。D007残作業: シナリオごとの波間隔チューニング
- [ ] `scenario/wave_schedule.go`: D002「時間圧力」の実装: 最初の侵入波が「構築が十分でないタイミング」に来るようにする。InitialState.StartingChi と最安部屋コストから「最低限の構築に必要なティック数」を算出し、その途中で最初の波を設定
- [ ] `scenario/wave_schedule_test.go`: スケジュール生成テスト、時間圧力の検証テスト（初波が構築完了前に来ること）、難易度エスカレーションテスト

## Phase 6-E: 地形テンプレートと制約（scenario/）

- [ ] `scenario/terrain.go`: TerrainTemplate 構造体（HardRockZones []Rect, SoftRockZones []Rect, WaterZones []Rect, UndiggableZones []Rect）。Rect 構造体（X, Y, W, H int）
- [ ] `scenario/terrain.go`: ApplyTerrain(cave *world.Cave, template TerrainTemplate) — 地形テンプレートをCaveに適用。D002原則1「不完全性の強制」: 掘削不可能な岩盤ゾーンにより理想配置を阻害
- [ ] `scenario/terrain.go`: GenerateRandomTerrain(width, height int, density float64, rng types.RNG) TerrainTemplate — ランダムに掘削不可能ゾーンを配置。density が高いほど制約が強い
- [ ] `scenario/terrain_test.go`: テンプレート適用テスト、UndiggableZone に部屋配置が拒否されるテスト、ランダム生成の決定論テスト

## Phase 6-F: 仙獣進化システム（senju/ 拡張）

- [ ] `senju/evolution.go`: EvolutionCondition 構造体（MinLevel int, MinFengShuiScore float64, RequiredElement types.Element, RequiredChiPool float64）。仙獣が進化するための条件
- [ ] `senju/evolution.go`: EvolutionPath 構造体（FromSpeciesID string, ToSpeciesID string, Condition EvolutionCondition, ChiCost float64）。進化経路の定義
- [ ] `senju/evolution.go`: EvolutionEngine 構造体。CheckEvolution(beast *Beast, roomChi *fengshui.RoomChi, fengShuiScore float64, chiPoolBalance float64) *EvolutionPath — 進化可能かチェック。Evolve(beast *Beast, path *EvolutionPath, registry *BeastSpeciesRegistry) error — 進化実行（ステータス変更、種族変更）
- [ ] `senju/evolution_data.json`: 進化経路データ（各基本種族に1つの進化先を定義）
- [ ] `senju/evolution_test.go`: 進化条件チェックテスト、進化実行テスト（ステータス変更確認）、条件未達で進化しないテスト、進化コストの確認テスト

## Phase 6-G: 仙獣敗北後処理（senju/ 拡張）

- [ ] `senju/defeat.go`: DefeatProcessor 構造体。ProcessDefeat(beast *Beast, tick types.Tick) DefeatResult — 仙獣敗北時の処理:
  - 仙獣のHPが0以下: Stunned 状態にする（Defeated → Stunned 遷移）
  - Stunned 状態は RecoveryTicks ティック後に Recovery 状態へ遷移
  - Recovery 状態の仙獣はHP1で復活、レベルが1下がる（最低1）
- [ ] `senju/defeat.go`: DefeatResult 構造体（BeastID int, NewState BeastState, RecoveryTick types.Tick, LevelPenalty int）
- [ ] `senju/defeat.go`: BeastState に Stunned/Recovering を追加。既存の行動エンジンはこれらの状態の仙獣をスキップ
- [ ] `senju/defeat_test.go`: 敗北→Stunned遷移テスト、Stunned→Recovering遷移テスト、復活後のステータス確認テスト、レベル低下テスト

## Phase 6-H: シナリオローダーとバリデーション（scenario/）

- [ ] `scenario/loader.go`: LoadScenario(data []byte) (*Scenario, error) — JSONからシナリオ定義をロード。条件・イベント・制約のバリデーション含む
- [ ] `scenario/loader.go`: ValidateScenario(s *Scenario) error — シナリオ整合性チェック:
  - 勝利条件が少なくとも1つ存在すること
  - InitialState の部屋配置が Cave サイズ内に収まること
  - WaveSchedule の TriggerTick が MaxTicks 以内であること
  - 参照する RoomTypeID/SpeciesID が存在すること
- [ ] `scenario/testdata/tutorial.json`: チュートリアルシナリオ定義（小マップ、弱い侵入波1つ、生存条件のみ）
- [ ] `scenario/testdata/standard.json`: 標準シナリオ定義（中マップ、5波の侵入、風水スコア + 全波撃退の勝利条件）
- [ ] `scenario/loader_test.go`: JSONロードテスト、バリデーション成功/失敗テスト、不正シナリオの拒否テスト

## Phase 6-I: シリアライズ（scenario/）

- [ ] `scenario/serialization.go`: MarshalScenarioProgress(progress ScenarioProgress) ([]byte, error) / UnmarshalScenarioProgress(data []byte) (ScenarioProgress, error)。ScenarioProgress 構造体（ScenarioID string, CurrentTick types.Tick, FiredEvents []string, WaveResults []WaveResult）
- [ ] `scenario/serialization_test.go`: 保存→復元→等価検証テスト、進行途中状態の保存/復元テスト

## Phase 6-J: ASCII可視化

- [ ] `scenario/ascii.go`: RenderScenarioStatus(scenario, progress, gameState) string — シナリオ進行状況のワンライン表示。`[Scenario: tutorial | Tick: 150/500 | Waves: 2/5 | Win: FengShui 45/80]` 形式
- [ ] `cmd/caveviz/main.go` 更新: `--scenario` フラグでシナリオ進行状況表示

## Phase 6-K: 統合検証

- [ ] `scenario/integration_test.go`: チュートリアルシナリオの完全シミュレーション（100ティック）:
  - シナリオロード → InitialState からCave構築 → 侵入波スケジュール生成
  - 勝利条件の到達可能性テスト（チュートリアルは適切にプレイすれば勝てる）
  - 敗北条件のトリガーテスト（何もしなければ負ける）
  - イベントトリガーの発火テスト
  - 仙獣進化の発生テスト（条件を満たすシナリオ設定で）
  - 仙獣敗北→復活のサイクルテスト
- [ ] `go vet ./...` と `go test -race ./...` がクリーンに通ることを確認
- [ ] Phase 6 完了。DECISIONS.md 更新、PHASE_COMPLETE 更新、次フェーズドラフトを `tasks_phase7_draft.md` として生成。**tasks.md には新しい未完了タスクを追加しない**
