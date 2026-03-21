# Phase 5 ドラフト — 経済システム（economy/）

> PRD Phase 5: リソースの収支管理、建設コスト。
> D002準拠: 原則3「トレードオフの連続」をこのフェーズで構造的に実装する。気の総量は常に不足する設計。

## Phase 5-A: 経済基本型定義（economy/）

- [ ] `economy/doc.go`: パッケージドキュメント。気の収支バランス管理、建設コスト、侵入報酬、経済イベントを扱うパッケージであることを記述
- [ ] `economy/chi_pool.go`: ChiPool 構造体（Total float64, Income float64, Expense float64, History []ChiTransaction）。Deposit(amount, reason) / Withdraw(amount, reason) (bool, error)。Balance() float64。CanAfford(amount) bool
- [ ] `economy/chi_pool.go`: ChiTransaction 構造体（Tick types.Tick, Amount float64, Type TransactionType, Reason string）。TransactionType 型（Supply/Consumption/Reward/Theft/Construction/BeastCost）
- [ ] `economy/chi_pool_test.go`: Deposit/Withdraw テスト、残高不足時の Withdraw 失敗テスト、トランザクション履歴テスト

## Phase 5-B: 気の供給計算（economy/）

- [ ] `economy/supply.go`: SupplyCalculator 構造体。CalcTickSupply(chiFlowEngine, fengShuiScore) float64 — 龍脈からの1ティック分の気供給量を計算。供給量 = 全龍脈のFlowRate合計 × 風水スコアボーナス倍率
- [ ] `economy/supply.go`: FengShuiBonus(score float64) float64 — 風水スコアを経済ボーナス倍率に変換。スコア0.0→倍率0.8（ペナルティ）、0.5→1.0（等倍）、1.0→1.3（ボーナス）。線形補間
- [ ] `economy/supply_test.go`: 龍脈なしで供給0テスト、龍脈複数の合算テスト、風水スコアによるボーナス/ペナルティテスト

## Phase 5-C: 維持コストモデル（economy/）

- [ ] `economy/cost_params.go`: CostParams 構造体（RoomMaintenancePerTick map[string]float64, BeastMaintenancePerTick float64, BeastGrowthCostPerLevel float64, TrapMaintenancePerTick float64）。DefaultCostParams()。LoadCostParams(data []byte)
- [ ] `economy/cost_params_data.json`: デフォルトコストパラメータ。部屋維持（龍穴: 0.5, 蓄気室: 0.3, 仙獣部屋: 0.2, 罠部屋: 0.4, 回復室: 0.3, 倉庫: 0.1）、仙獣維持: 0.3/tick、仙獣成長: 5.0/level、罠維持: 0.2/tick
- [ ] `economy/maintenance.go`: MaintenanceCalculator 構造体。CalcTickMaintenance(rooms, beasts, traps, params) float64 — 1ティック分の維持コスト合計を計算。部屋数・仙獣数・罠数に比例
- [ ] `economy/maintenance_test.go`: 部屋0で維持コスト0テスト、部屋タイプ別コストテスト、仙獣追加で維持コスト増加テスト

## Phase 5-D: 建設コストモデル（economy/）

- [ ] `economy/construction.go`: ConstructionCost 構造体（RoomCost map[string]float64, CorridorCostPerCell float64, RoomUpgradeCostPerLevel map[string]float64）。DefaultConstructionCost()。LoadConstructionCost(data []byte)
- [ ] `economy/construction_data.json`: デフォルト建設コスト。部屋建設（龍穴: 50.0, 蓄気室: 20.0, 仙獣部屋: 15.0, 罠部屋: 25.0, 回復室: 20.0, 倉庫: 10.0）、通路: 2.0/セル、部屋強化（レベルごとに基本コスト × レベル倍率）
- [ ] `economy/construction.go`: CalcRoomCost(roomTypeID string, costs) float64。CalcCorridorCost(pathLength int, costs) float64。CalcUpgradeCost(roomTypeID string, currentLevel int, costs) float64
- [ ] `economy/construction_test.go`: 部屋建設コストテスト、通路コストテスト、部屋強化コストのレベルスケーリングテスト

## Phase 5-E: 仙獣コストモデル（economy/）

- [ ] `economy/beast_cost.go`: BeastCost 構造体（SummonCostByElement map[types.Element]float64, GrowthCostPerLevel float64, EvolutionCost float64）。DefaultBeastCost()。LoadBeastCost(data []byte)
- [ ] `economy/beast_cost_data.json`: デフォルト仙獣コスト。召喚（木: 30.0, 火: 35.0, 土: 25.0, 金: 40.0, 水: 30.0）、成長: 5.0/level、進化: 50.0
- [ ] `economy/beast_cost.go`: CalcSummonCost(element) float64。CalcGrowthCost(currentLevel) float64
- [ ] `economy/beast_cost_test.go`: 属性別召喚コストテスト、レベル別成長コストテスト

## Phase 5-F: 侵入報酬と損失（economy/）

- [ ] `economy/invasion_economy.go`: InvasionEconomyProcessor 構造体。ProcessInvasionResult(events []invasion.InvasionEvent, chiPool *ChiPool, tick types.Tick) InvasionEconomySummary — 侵入結果を経済に反映:
  1. CollectRewards で得た RewardChi を ChiPool に Deposit（TransactionType: Reward）
  2. CollectStolenChi で失った StolenChi を ChiPool から Withdraw（TransactionType: Theft）
  3. 仙獣敗北による復活コスト（将来拡張用スタブ）
- [ ] `economy/invasion_economy.go`: InvasionEconomySummary 構造体（RewardChi float64, StolenChi float64, NetChi float64, BeastRevivalCost float64）
- [ ] `economy/invasion_economy_test.go`: 侵入者撃破で報酬獲得テスト、盗賊逃走で気損失テスト、報酬と損失の差し引きテスト

## Phase 5-G: 経済ティックエンジン（economy/）

- [ ] `economy/engine.go`: EconomyEngine 構造体（ChiPool, SupplyCalculator, MaintenanceCalculator, CostParams）。NewEconomyEngine(chiPool, params)
- [ ] `economy/engine.go`: EconomyEngine.Tick(tick, chiFlowEngine, fengShuiScore, rooms, beasts, traps) EconomyTickResult — 1ティック分の経済処理:
  1. 気供給を計算して ChiPool に Deposit
  2. 維持コストを計算して ChiPool から Withdraw
  3. 収支バランスを記録
- [ ] `economy/engine.go`: EconomyTickResult 構造体（Tick types.Tick, Supply float64, Maintenance float64, Balance float64, IsDeficit bool）。IsDeficit = Maintenance > Supply の場合 true
- [ ] `economy/engine.go`: EconomyEngine.CanBuildRoom(roomTypeID) bool。CanSummonBeast(element) bool。CanUpgradeRoom(roomTypeID, level) bool — 建設可否判定（ChiPool の残高チェック）
- [ ] `economy/engine.go`: EconomyEngine.ExecuteBuild(roomTypeID, tick) error。ExecuteSummon(element, tick) error。ExecuteUpgrade(roomTypeID, level, tick) error — 建設実行（コスト引き落とし + トランザクション記録）
- [ ] `economy/engine_test.go`: 供給→維持の収支テスト、赤字判定テスト、建設可否テスト、建設実行でChiPool減少テスト、残高不足で建設失敗テスト

## Phase 5-H: シリアライズ（economy/）

- [ ] `economy/serialization.go`: MarshalEconomyState(chiPool, engine) ([]byte, error) / UnmarshalEconomyState(data []byte) (*ChiPool, error) — ChiPool とトランザクション履歴の保存/復元
- [ ] `economy/serialization_test.go`: 保存→復元→等価検証テスト、トランザクション履歴の保存/復元テスト

## Phase 5-I: 経済バランス検証

- [ ] `economy/balance_test.go`: 標準的な構成（部屋6つ、仙獣3体、龍脈1本）での100ティック経済シミュレーション。D002原則3「リソースが常に不足」を検証:
  - 維持コストのみで供給の70%以上が消費されること
  - 新規建設をすると一時的に赤字になること
  - 侵入報酬が建設コストの部分回復になること（全額はカバーしない）
  - 100ティック時点で「全部屋MAX + 全仙獣MAX」に到達不可能なバランスであること

## Phase 5-J: 統合検証

- [ ] `economy/integration_test.go`: Cave + ChiFlowEngine + 仙獣 + 侵入波 + 経済エンジンの50ティックフルシミュレーション:
  - 気供給が毎ティック ChiPool に入ること
  - 維持コストが毎ティック引き落とされること
  - 侵入波の撃退報酬が ChiPool に加算されること
  - 盗賊逃走の損失が ChiPool から減算されること
  - 赤字状態で建設不可になること
  - 全トランザクションが記録されていること
- [ ] `go vet ./...` と `go test -race ./...` がクリーンに通ることを確認
- [ ] Phase 5 完了。DECISIONS.md 更新、PHASE_COMPLETE 更新、次フェーズドラフトを `tasks_phase6_draft.md` として生成。**tasks.md には新しい未完了タスクを追加しない**
