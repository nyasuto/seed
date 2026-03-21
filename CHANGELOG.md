# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] — 2026-03-21

chaosseed-core v1.0.0 — 全7フェーズ完了。ダンジョン経営シミュレーションのコアエンジン初版リリース。

### Phase 0: プロジェクト初期化・エコシステム整備

- Go モジュール初期化、ディレクトリ構造作成
- CI/CD（GitHub Actions）、golangci-lint、Makefile、カバレッジ計測
- PRD、README、LICENSE（MIT）

### Phase 1: マップシステム（types/, world/）

- 共有型定義: Pos, Direction, Element（五行）, RNG インターフェース, Tick
- 五行の相生・相克判定
- グリッドベースの洞窟マップ: Grid, Cell, CellType
- 部屋の定義・配置・重複チェック: Room, RoomType, RoomTypeRegistry（JSON駆動）
- BFS ベースの通路生成: BuildCorridor
- 洞窟全体管理: Cave, AdjacencyGraph
- JSON シリアライズ/デシリアライズ
- ASCII 可視化ツール（caveviz）

### Phase 2: 風水システム（fengshui/）

- 龍脈（DragonVein）: BFS 経路計算、動的再計算
- 気の蓄積・伝播モデル: ChiFlowEngine（供給・伝播・減衰）
- 風水評価スコア: FengShuiScore, Evaluator（ScoreParams JSON 外出し）
- シリアライズ、ASCII 可視化レイヤー

### Phase 3: 仙獣システム（senju/）

- 仙獣種族定義: Species, SpeciesRegistry（JSON駆動）
- 仙獣インスタンス・配置: Beast, Placement, CombatStats, Affinity
- 成長システム: GrowthEngine（EXP、レベルアップ、気消費）
- 行動AI: Guard/Patrol/Chase/Flee パターン、BehaviorEngine
- AI パラメータ JSON 外出し
- シリアライズ、ASCII 可視化レイヤー

### Phase 4: 侵入システム（invasion/）

- 侵入者定義: InvaderClass, InvaderClassRegistry（JSON駆動）
- 目標指向経路探索: PathFinder（コア部屋への最短経路）
- 戦闘解決: ResolveRoomCombat（SPD 順ペアリング、1ティック1ラウンド）
- 撤退ロジック、罠効果システム
- 侵入波管理: WaveSchedule, WaveSpawner
- 侵入ティックエンジン: InvasionEngine
- シリアライズ、ASCII 可視化レイヤー

### Phase 5: 経済システム（economy/）

- ChiPool（プレイヤー通貨）: 物理層（ChiFlowEngine）と経済層の二層構造（D009）
- 気の供給計算: SupplyCalculator（龍脈 × 充填率 × 風水倍率）
- 維持コストモデル: MaintenanceCalculator（部屋・仙獣・罠）
- 赤字処理: DeficitProcessor（軽度/中度/重度の段階的ペナルティ）
- 建設・仙獣コストモデル
- 侵入報酬・損失: InvasionEconomyProcessor
- 経済ティックエンジン: EconomyEngine
- シリアライズ、ASCII 可視化レイヤー

### Phase 6: シナリオシステム（scenario/）

- 仙獣進化システム: EvolutionEngine（Phase 3 延期分）
- 仙獣敗北後処理: DefeatProcessor（スタン→復活、HP 0 で消滅）
- 地形バリエーション: HardRock/Water セル、TerrainGenerator、ValidateTerrain（詰み防止）
- シナリオ定義: Scenario, GameConstraints, InitialState
- 勝利/敗北条件: ConditionEvaluator（survive_until, defeat_all_waves, chi_pool_above, core_hp_zero 等）
- イベントシステム: EventEngine + EventCommand パターン（D011）
- 動的侵入波スケジュール: DynamicWaveScheduler, CalcFirstWaveTiming
- 地形テンプレート生成: TerrainTemplate
- シナリオローダー・バリデーション（JSON駆動）
- CoreHP（龍穴耐久値）: D010
- シリアライズ、ASCII 可視化レイヤー

### Phase 7: 統合シミュレーション（simulation/）

- PlayerAction 型定義（DigRoom, DigCorridor, SummonBeast, UpgradeRoom）
- コマンド実行器: CommandExecutor（EventCommand → 状態変更）
- メインループ: SimulationEngine（12ステップ更新順序 D012）
- GameState, GameSnapshot（読み取り専用ビュー）
- スナップショット・チェックポイント: CreateCheckpoint / RestoreCheckpoint
- リプレイ: RecordReplay / PlayReplay（JSON シリアライズ）
- AIプレイヤー: SimpleAIPlayer, RandomAIPlayer
- D001 プロファイリング: OnCaveChanged 線形スケール確認、差分更新不要と結論
- D002 定量検証: 原則1（不完全性強制）、原則2（時間圧力）、原則3（トレードオフ）全検証通過
- SimulationRunner: RunWithAI, RunInteractive, BatchRun
- ASCII 統合表示: RenderFullStatus
- 統合テスト: 決定論性、チェックポイント復元、リプレイ再生、大規模ストレス（64x64, 10部屋, 8仙獣, 10波）

### 設計判断

- D001: OnCaveChanged 全再計算（差分更新不要）— RESOLVED
- D002: コアゲーム体験「常に不完全な状態での判断」— 3原則すべて実装・検証完了
- D003: 仙獣の気消費は GrowthEngine 側で RoomChi.Current を直接減算 — ACTIVE
- D004: 戦闘ステータスはクエリ時計算 — RESOLVED
- D005: 仙獣AI行動衝突は先着順 — ACTIVE
- D006: 侵入者位置は map[int][]int で受け渡し — RESOLVED
- D007: 時間圧力の段階的実装 — RESOLVED（Phase 6 完了）
- D008: 戦闘マッチングは SPD 順ペアリング・1ティック1ラウンド — ACTIVE
- D009: ChiPool（経済層）と ChiFlowEngine（物理層）の二層構造 — ACTIVE
- D010: CoreHP（龍穴コア耐久値）— ACTIVE
- D011: EventCommand パターン — ACTIVE
- D012: ティック更新順序の厳密な定義（12ステップ）— ACTIVE
- D013: PlayerAction は Step() の引数として毎ティック注入 — ACTIVE
