# Phase 4 タスクドラフト — 侵入システム（invasion/）

> PRD Phase 4: 侵入者の生成、経路探索、戦闘解決。

## Phase 4-A: 侵入者基本型定義（invasion/）

- [ ] `invasion/doc.go`: パッケージドキュメント。侵入者の種族定義・生成・経路探索・戦闘解決を扱うパッケージであることを記述
- [ ] `invasion/invader_type.go`: InvaderType 構造体（ID string, Name string, Element types.Element, BaseHP int, BaseATK int, BaseDEF int, BaseSPD int, RewardChi float64, Description string）。InvaderTypeRegistry（map管理、JSONから一括ロード）
- [ ] `invasion/invader_type_data.json`: 初期侵入者タイプ5種の定義（各五行属性に1種）。木: 木行の修行者（バランス型）、火: 火行の武闘家（攻撃型）、土: 土行の鎧武者（防御型）、金: 金行の剣士（速度型）、水: 水行の道士（回復型）。各タイプに基本ステータスとRewardChi（撃退時の気報酬）を設定
- [ ] `invasion/invader.go`: Invader 構造体（ID int, TypeID string, Name string, Element types.Element, RoomID int, HP int, MaxHP int, ATK int, DEF int, SPD int, Level int, State InvaderState, EntryTick types.Tick）。InvaderState 型（Exploring/Fighting/Retreating/Defeated）。NewInvader(id, invaderType, level, tick) *Invader でレベルに応じたステータス計算
- [ ] `invasion/invader_test.go`: InvaderTypeRegistryのJSONロードテスト、全5種が取得できるテスト、レベルスケーリングテスト

## Phase 4-B: 侵入者の経路探索AI（invasion/）

- [ ] `invasion/pathfinder.go`: InvaderPathfinder 構造体。FindNextRoom(invader, cave, adjacencyGraph, rng) int — 侵入者の次の移動先部屋IDを返す。入口から未探索の部屋を優先的に探索する貪欲アルゴリズム。行き止まりに到達したらバックトラック
- [ ] `invasion/pathfinder.go`: ExplorationMemory 構造体（VisitedRooms map[int]bool）— 侵入者ごとの探索記憶。未訪問の隣接部屋を優先、全訪問済みならランダムに選択
- [ ] `invasion/pathfinder_test.go`: 直線経路の探索テスト、分岐での未訪問優先テスト、行き止まりからのバックトラックテスト、全部屋訪問後のランダム移動テスト

## Phase 4-C: 戦闘解決システム（invasion/）

- [ ] `invasion/combat_params.go`: CombatParams 構造体（ATKMultiplier float64, DEFReduction float64, ElementAdvantage float64, ElementDisadvantage float64, TrapDamageBase int, CriticalChance float64, CriticalMultiplier float64）。DefaultCombatParams()。LoadCombatParams(data []byte)
- [ ] `invasion/combat_params_data.json`: デフォルト戦闘パラメータ（ATK倍率: 1.0, DEF減算率: 0.5, 属性有利: 1.5, 属性不利: 0.7, 罠基本ダメージ: 20, クリティカル率: 0.1, クリティカル倍率: 2.0）
- [ ] `invasion/combat.go`: CombatEngine 構造体。NewCombatEngine(params, rng)。ResolveCombat(beast *senju.Beast, invader *Invader, roomChi *fengshui.RoomChi) CombatResult — 1回の戦闘ラウンドを解決:
  1. 仙獣の実効ステータス = beast.CalcCombatStats(roomChi)
  2. ダメージ計算: ATK × ATKMultiplier - 相手DEF × DEFReduction
  3. 属性相性倍率の適用（仙獣Element vs 侵入者Element）
  4. クリティカル判定（RNG経由）
  5. 両者のHP減算
  6. 勝敗判定（HP 0以下で敗北）
- [ ] `invasion/combat.go`: CombatResult 構造体（BeastDamage int, InvaderDamage int, BeastHP int, InvaderHP int, IsBeastDefeated bool, IsInvaderDefeated bool, WasCritical bool）
- [ ] `invasion/combat_test.go`: 基本ダメージ計算テスト、属性有利/不利テスト、クリティカルヒットテスト（FixedRNG使用）、DEFが高い場合のダメージ下限テスト（最低1ダメージ保証）、両者生存の継続戦闘テスト

## Phase 4-D: 罠効果システム（invasion/）

- [ ] `invasion/trap.go`: TrapEffect 構造体（RoomID int, DamagePerTick int, SlowTicks int, Element types.Element）。ApplyTrap(invader, trap, combatParams) TrapResult — 罠の効果を侵入者に適用。属性相性で罠ダメージに倍率。SlowTicks 中は侵入者の移動を1ティック遅延
- [ ] `invasion/trap.go`: TrapResult 構造体（Damage int, IsSlowed bool, RemainingSlowTicks int）
- [ ] `invasion/trap.go`: BuildTrapEffects(cave, rooms, roomTypes) []TrapEffect — 罠部屋からTrapEffectリストを構築。罠部屋のElementを罠のElementとする
- [ ] `invasion/trap_test.go`: 罠ダメージ計算テスト、属性相性倍率テスト、スロー効果テスト、罠部屋でない部屋は効果なしテスト

## Phase 4-E: 侵入波管理（invasion/）

- [ ] `invasion/wave.go`: InvasionWave 構造体（ID int, TriggerTick types.Tick, Invaders []*Invader, State WaveState）。WaveState 型（Pending/Active/Completed/Failed）。IsActive() bool, IsCompleted() bool
- [ ] `invasion/wave_generator.go`: WaveGenerator 構造体。NewWaveGenerator(typeRegistry, rng)。GenerateWave(waveNumber int, difficulty float64, tick types.Tick) *InvasionWave — 波番号と難易度に応じた侵入者グループを生成。難易度に応じて侵入者数・レベルがスケール
- [ ] `invasion/wave_generator.go`: WaveSchedule 構造体（Waves []WaveConfig）。WaveConfig（TriggerTick types.Tick, Difficulty float64, InvaderCount int）。LoadWaveSchedule(data []byte) で JSON からスケジュール読み込み
- [ ] `invasion/wave_schedule_data.json`: テスト用の侵入波スケジュール（3波: tick 50 で弱い3体、tick 150 で中程度5体、tick 300 で強い7体）
- [ ] `invasion/wave_test.go`: 波の生成テスト、難易度スケーリングテスト、WaveScheduleのJSONロードテスト

## Phase 4-F: 侵入ティックエンジン（invasion/）

- [ ] `invasion/engine.go`: InvasionEngine 構造体（CombatEngine, WaveGenerator, cave, adjacencyGraph, rng）。NewInvasionEngine(cave, adjacencyGraph, combatEngine, behaviorEngine, rng)
- [ ] `invasion/engine.go`: InvasionEngine.Tick(currentTick, waves, beasts, rooms, roomTypes, roomChi) []InvasionEvent — 1ティック分の侵入処理:
  1. TriggerTick に到達した Pending 波を Active に
  2. Active 波の各侵入者を移動（FindNextRoom）
  3. 仙獣と同じ部屋にいる侵入者は戦闘（ResolveCombat）
  4. 罠部屋にいる侵入者に罠効果を適用
  5. HP 0 以下の侵入者を Defeated に
  6. 全侵入者が Defeated/Retreating なら波を Completed に
  7. InvasionEvent リストを返す
- [ ] `invasion/engine.go`: InvasionEvent 構造体（Type InvasionEventType, Tick types.Tick, InvaderID int, BeastID int, RoomID int, Details string）。InvasionEventType（WaveStarted/InvaderMoved/CombatOccurred/TrapTriggered/InvaderDefeated/InvaderRetreated/WaveCompleted）
- [ ] `invasion/engine.go`: InvasionEngine.BuildInvaderPositions(waves) map[int][]int — アクティブな侵入者の位置マップを構築。senju.BehaviorEngine.Tick に渡す用
- [ ] `invasion/engine_test.go`: 波のアクティベーションテスト、侵入者移動テスト、戦闘発生テスト、罠効果テスト、波完了判定テスト、InvaderPositions構築テスト

## Phase 4-G: シリアライズ（invasion/）

- [ ] `invasion/serialization.go`: MarshalInvasionState(waves []*InvasionWave) ([]byte, error) / UnmarshalInvasionState(data []byte, typeRegistry) ([]*InvasionWave, error) — 全侵入波の状態を保存/復元
- [ ] `invasion/serialization_test.go`: 保存→復元→等価検証テスト、空リストの保存/復元テスト

## Phase 4-H: ASCII可視化への侵入レイヤー追加

- [ ] `invasion/ascii.go`: RenderInvasionOverlay(cave, waves) string — CaveのASCII表示に侵入者の位置をオーバーレイ。侵入者は `!!` で表示。戦闘中の部屋は `⚔️` または `XX` で表示
- [ ] `cmd/caveviz/main.go` 更新: `--invasion` フラグで侵入レイヤー表示
- [ ] `invasion/ascii_test.go`: 小さなCaveで侵入オーバーレイの出力テスト

## Phase 4-I: 統合検証

- [ ] `invasion/integration_test.go`: Cave（部屋5つ、通路接続）+ ChiFlowEngine + 仙獣3体（Guard/Patrol/Chase）+ 侵入波1つ（侵入者3体）→ 50ティック侵入シミュレーション→侵入者が入口から部屋を探索することを検証→仙獣が侵入者を追跡・戦闘することを検証→罠部屋でダメージを受けることを検証→侵入波が完了することを検証→気の報酬が発生することを検証
- [ ] `go vet ./...` と `go test -race ./...` がクリーンに通ることを確認
- [ ] Phase 4 完了。DECISIONS.md 更新、PHASE_COMPLETE 更新、次フェーズドラフトを `tasks_phase5_draft.md` として生成。**tasks.md には新しい未完了タスクを追加しない**
