# tasks.md — chaosseed-core

<!-- Phase: 1 (world/) — 洞窟マップシステム -->

## Phase 0: プロジェクト初期化

- [x] `go mod init github.com/ponpoko/chaosseed-core` でモジュール初期化
- [x] ディレクトリ構造を作成: `types/`, `world/`, `testutil/`, `docs/`
- [x] PRD を `docs/PRD.md` に配置（参照用コピー）

## Phase 0-B: エコシステム整備

- [x] `.gitignore` 作成: バイナリ(`*.exe`,`*.out`), カバレッジ(`coverage.out`,`coverage.html`), ログ(`logs/`), IDE設定(`.idea/`,`.vscode/`), OS生成ファイル(`.DS_Store`)
- [x] `LICENSE` 作成: MIT License、著作者名とYearを記入
- [x] `README.md` 作成: プロジェクト概要、ビルド方法（`go test ./...`）、アーキテクチャ図（テキストベースのディレクトリツリー）、ライセンス表記。最小限でよい、Phase進行に合わせて育てる
- [x] `Makefile` 作成: ターゲット `test`(`go test ./...`), `test-race`(`go test -race ./...`), `vet`(`go vet ./...`), `lint`(`golangci-lint run`), `cover`(カバレッジHTML生成), `check`(`vet` + `lint` + `test-race` を順に実行), `clean`(生成物削除)
- [x] `.golangci.yml` 作成: 有効linter — `govet`, `errcheck`, `staticcheck`, `unused`, `gosimple`, `ineffassign`, `typecheck`。タイムアウト3分。`testdata/` を除外
- [x] `.github/workflows/ci.yml` 作成: on push/PR → Go setup → `make check` 実行。Go バージョンは matrix で 1.22.x。runs-on ubuntu-latest
- [x] `.github/workflows/coverage.yml` 作成: on push to main → カバレッジ計測 → 80%未満で warning（fail はしない、初期は厳しすぎるので）
- [x] `CHANGELOG.md` 作成: Keep a Changelog 形式、`## [Unreleased]` セクションのみで開始

## Phase 1-A: 共有型定義（types/）

- [x] `types/pos.go`: Pos 構造体（X, Y int）、Add/Sub/Distance メソッド、Neighbors() で上下左右の隣接Posを返す
- [x] `types/direction.go`: Direction 型（North/South/East/West）、Opposite() メソッド、Delta() で移動量Posを返す
- [x] `types/element.go`: Element 型（Wood/Fire/Earth/Metal/Water）、String() メソッド
- [x] `types/element_relation.go`: Generates(from, to) bool（相生判定: 木→火→土→金→水→木）、Overcomes(from, to) bool（相克判定: 木→土→水→火→金→木）
- [x] `types/rng.go`: RNG インターフェース定義（Intn, Float64）、NewSeededRNG(seed int64) で deterministic な実装を返す
- [x] `types/tick.go`: Tick 型（uint64）の定義
- [x] `types/types_test.go`: 相生・相克の全組み合わせテスト、Pos演算テスト、Direction.Opposite テスト

## Phase 1-B: グリッドとセル管理（world/）

- [x] `world/cell.go`: CellType 定義（Rock/Corridor/RoomFloor/Entrance）、Cell 構造体（Type, RoomID（部屋の一部なら所属ID, それ以外は0））
- [x] `world/grid.go`: Grid 構造体（Width, Height int, cells [][]Cell）、NewGrid(w, h) コンストラクタ、At(pos)/Set(pos, cell) メソッド、InBounds(pos) bool
- [x] `world/grid_test.go`: グリッド生成、範囲外アクセスのエラー、セル読み書きのテスト

## Phase 1-C: 部屋の定義と配置（world/）

- [x] `world/room_type.go`: RoomType 構造体（ID string, Name string, Element types.Element, BaseChiCapacity int, Description string）、RoomTypeRegistry（map管理、JSONから一括ロード）
- [x] `world/room_type_data.json`: 初期部屋タイプ6種の定義（龍穴/蓄気室/仙獣部屋/罠部屋/回復室/倉庫）
- [x] `world/room.go`: Room 構造体（ID int, TypeID string, Pos types.Pos, Width int, Height int, Level int, Entrances []RoomEntrance）、RoomEntrance（Pos types.Pos, Dir types.Direction）
- [x] `world/room_placement.go`: CanPlaceRoom(grid, room) bool（範囲内チェック、重複チェック、岩盤上のみ）、PlaceRoom(grid, room) error（セルをRoomFloorに書き換え、RoomIDをセット）
- [x] `world/room_test.go`: 正常配置テスト、範囲外配置の拒否テスト、重複配置の拒否テスト、RoomTypeRegistryのJSONロードテスト

## Phase 1-D: 通路の生成（world/）

- [x] `world/corridor.go`: Corridor 構造体（ID int, FromRoomID int, ToRoomID int, Path []types.Pos）
- [x] `world/corridor_builder.go`: BuildCorridor(grid, fromPos, toPos) (Corridor, error) — BFSベースで岩盤を掘って最短経路を生成。既存の通路/部屋床は通過可、他部屋の内部は回避
- [x] `world/corridor_builder_test.go`: 隣接部屋間の直線通路テスト、障害物を迂回する通路テスト、到達不能ケースのエラーテスト

## Phase 1-E: 洞窟全体の管理（world/）

- [x] `world/cave.go`: Cave 構造体（Grid, Rooms []Room, Corridors []Corridor, nextRoomID/nextCorridorID の自動採番）、NewCave(w, h) コンストラクタ
- [x] `world/cave_ops.go`: Cave.AddRoom(roomType, pos, w, h) (Room, error) — バリデーション→配置→登録を一括実行。Cave.ConnectRooms(roomID1, roomID2) (Corridor, error) — 最寄りの入口同士を通路接続
- [x] `world/adjacency.go`: AdjacencyGraph 構造体（部屋IDをノード、通路をエッジとするグラフ）、Cave.BuildAdjacencyGraph() AdjacencyGraph、Neighbors(roomID) []int、PathExists(from, to) bool（BFS）
- [x] `world/cave_test.go`: Cave生成→部屋2つ追加→通路接続→隣接グラフ確認の結合テスト

## Phase 1-F: シリアライズ（world/）

- [x] `world/serialization.go`: Cave.MarshalJSON() ([]byte, error)、UnmarshalCave(data []byte) (\*Cave, error) — Grid全セル + Rooms + Corridors の完全保存/復元
- [x] `world/serialization_test.go`: 部屋と通路を含むCaveを保存→復元→元と等価であることを検証。空のCaveの保存/復元テスト

## Phase 1-G: 統合検証

- [x] `world/integration_test.go`: 中規模マップ（32x32）に部屋5つ配置→全接続→風水フェーズへの受け渡し用データ構造が正しく取れることを確認
- [x] testutil に `testutil/rng.go` を作成: FixedRNG（常に同じ値を返すモックRNG）と NewTestRNG(seed) ヘルパー
- [x] `go vet ./...` と `go test -race ./...` がクリーンに通ることを確認
- [x] Phase 1 完了。tasks.md の Phase 2 タスクを追記する（PRD Phase 2 を参照して fengshui/ のタスクを展開）

## Phase 1-H: ASCII可視化ツール（world確認用）

- [x] `cmd/caveviz/main.go` 作成: ハードコードでCaveを生成（部屋3〜4個＋通路接続）し、ASCIIグリッドを標準出力に表示。凡例: `██`=岩盤, `..`=通路, `[]`=部屋床, `><`=入口, 部屋IDを部屋内に表示
- [x] `world/ascii.go`: Cave.RenderASCII() string メソッド。Gridの各セルを文字にマッピング。部屋内にはRoomIDを1桁で表示（10以上はA,B,C...）
- [x] `world/ascii_test.go`: 小さなCave（8x8）でRenderASCIIの出力が期待文字列と一致するテスト
- [x] Makefileに `viz` ターゲット追加: `go run ./cmd/caveviz`

## Phase 2-A: 風水基本型定義（fengshui/）

- [x] `fengshui/doc.go`: パッケージドキュメント。龍脈・気・風水評価を扱うパッケージであることを記述
- [x] `fengshui/dragon_vein.go`: DragonVein 構造体（ID int, SourcePos types.Pos, Element types.Element, FlowRate float64, Path []types.Pos）。龍脈は洞窟の入口から内部へ気を運ぶ経路。RoomsOnPath(cave) []int で龍脈経路上の部屋IDリストを返す
- [x] `fengshui/chi.go`: RoomChi 構造体（RoomID int, Current float64, Capacity float64, Element types.Element）。IsFull() bool, IsEmpty() bool, Ratio() float64 メソッド。ElementはRoomTypeから引き継ぐ（気の属性＝部屋の属性）
- [x] `fengshui/chi_test.go`: RoomChi の基本メソッドテスト（IsFull/IsEmpty/Ratio の境界値、ゼロ容量のエッジケース）

## Phase 2-B: 龍脈の経路計算と動的再計算（fengshui/）

- [x] `fengshui/dragon_vein_builder.go`: BuildDragonVein(cave, sourcePos, element, flowRate) (\*DragonVein, error) — 入口から通路・部屋床を通る経路をBFSで計算。岩盤は通過不可
- [x] `fengshui/dragon_vein_builder.go`: RebuildDragonVein(cave, existingVein) (\*DragonVein, error) — 既存龍脈をCaveの現在の地形で再計算。部屋追加/通路掘削で経路が変化する。元のSourcePos/Element/FlowRateは維持
- [x] `fengshui/dragon_vein_test.go`: 龍脈が入口から部屋に到達するテスト、到達不能ケースのエラーテスト、部屋追加後にRebuildで経路が伸びることのテスト、通路がない部屋には到達しないテスト

## Phase 2-C: 気の蓄積・伝播モデル（fengshui/）

- [x] `fengshui/flow_params.go`: FlowParams 構造体（GeneratesMultiplier float64, OvercomesMultiplier float64, SameElementMultiplier float64, NeutralMultiplier float64, BaseDecayRate float64）。DefaultFlowParams() で初期値を返す。JSON読み込み対応: LoadFlowParams(path) (\*FlowParams, error)
- [x] `fengshui/flow_params_data.json`: デフォルトパラメータ定義（相生: 1.3, 相克: 0.6, 同属性: 1.1, 中立: 1.0, 基本減衰率: 0.02）
- [x] `fengshui/chi_flow.go`: ChiFlowEngine 構造体（Veins []DragonVein, RoomChi map[int]*RoomChi, Params *FlowParams, cave \*world.Cave）。NewChiFlowEngine(cave, veins, registry, params) で初期化。Tick() で1ティック分の気の流れを計算:
  1. 龍脈上の各部屋にFlowRate分の気を供給（龍脈のElementと部屋のElementの相性でFlowRateに倍率適用）
  2. 隣接部屋間の気の伝播: 気が多い部屋→少ない部屋へ差分の一定割合が移動。移動量に相生/相克の倍率を適用
  3. 全部屋に基本減衰（BaseDecayRate）を適用
  4. Capacityクランプ（0〜Capacity）
- [x] `fengshui/chi_flow.go`: ChiFlowEngine.OnCaveChanged(cave) — Cave変更時に龍脈を全再計算し、新しい部屋のRoomChiを追加。差分更新の最適化は後回し（DECISIONS.mdに記録）
- [x] `fengshui/chi_flow_test.go`:
  - 龍脈上の部屋への気の供給テスト
  - 龍脈Elementと部屋Elementが相生のとき供給量が増えるテスト
  - 龍脈Elementと部屋Elementが相克のとき供給量が減るテスト
  - 隣接部屋間の気の伝播テスト（高→低へ流れる）
  - 隣接部屋間の相生/相克倍率テスト
  - 基本減衰テスト
  - Capacityクランプテスト（上限・下限）
  - 龍脈上にない部屋には直接供給されないテスト
  - 10ティック経過後に定常状態に近づくテスト
  - OnCaveChanged: 部屋追加後に新部屋が気の計算に含まれるテスト

## Phase 2-D: 風水評価スコア（fengshui/）

- [x] `fengshui/score_params.go`: ScoreParams 構造体（GeneratesBonus float64, OvercomesPenalty float64, SameElementBonus float64, DragonVeinBonus float64, ChiRatioWeight float64）。DefaultScoreParams() で初期値。JSON読み込み対応: LoadScoreParams(path)
- [x] `fengshui/score_params_data.json`: デフォルトスコアパラメータ（相生: +20, 相克: -15, 同属性: +5, 龍脈接続: +30, 気充填率ウェイト: 100）
- [x] `fengshui/score.go`: FengShuiScore 構造体（RoomID int, ChiScore float64, AdjacencyScore float64, DragonVeinScore float64, Total float64）。内訳を保持
- [x] `fengshui/evaluator.go`: Evaluator 構造体。NewEvaluator(cave, registry, params) で生成。EvaluateRoom(roomID, flowEngine) FengShuiScore — 個別部屋のスコア算出。EvaluateAll(flowEngine) []FengShuiScore — 全部屋一括。CaveTotal(flowEngine) float64 — 洞窟全体の風水スコア合計。スコアリング:
  1. ChiScore = 部屋の気充填率 × ChiRatioWeight
  2. AdjacencyScore = 隣接部屋との五行相性ボーナス/ペナルティの合計（Paramsから取得）
  3. DragonVeinScore = 龍脈に接続されていれば DragonVeinBonus
  4. Total = ChiScore + AdjacencyScore + DragonVeinScore
- [x] `fengshui/evaluator_test.go`: 単独部屋のスコアテスト、相生隣接ボーナステスト、相克隣接ペナルティテスト、龍脈接続ボーナステスト、全部屋一括評価テスト、パラメータ変更がスコアに反映されるテスト

## Phase 2-E: 風水シリアライズ（fengshui/）

- [x] `fengshui/serialization.go`: ChiFlowEngine.MarshalJSON() / UnmarshalChiFlowEngine(data, cave, registry, params) — 龍脈・各部屋の気の状態を保存/復元。龍脈のPathは保存するが、復元時にcaveとの整合性を検証
- [x] `fengshui/serialization_test.go`: 保存→復元→等価検証テスト、空の状態の保存/復元テスト

## Phase 2-F: ASCII可視化への風水レイヤー追加

- [x] `fengshui/ascii.go`: RenderChiOverlay(cave, flowEngine) string — Caveの ASCII表示に気の充填率をオーバーレイ。充填率に応じて部屋内の表示を変える（0%: `__`, 1-33%: `░░`, 34-66%: `▒▒`, 67-99%: `▓▓`, 100%: `██`）。龍脈の経路を `~~` で表示
- [x] `cmd/caveviz/main.go` 更新: 風水レイヤー付き表示を追加。コマンドライン引数 `--chi` で切り替え
- [x] `fengshui/ascii_test.go`: 小さなCaveで風水オーバーレイの出力テスト

## Phase 2-G: 統合検証

- [x] `fengshui/integration_test.go`: 中規模Cave（32x32）に部屋5つ配置（意図的に相生ペアと相克ペアを含む）→龍脈2本設定→20ティックシミュレーション→相生ペアの部屋は気が多く、相克ペアは少ないことを検証→風水スコアが配置に応じて正しく変動することを検証
- [x] `go vet ./...` と `go test -race ./...` がクリーンに通ることを確認
- [x] Phase 2 完了。DECISIONS.md に差分更新を後回しにした判断を記録。次フェーズのタスクドラフトを `tasks_phase3_draft.md` として生成。プロジェクトルートに `PHASE_COMPLETE` ファイルを作成し、以下を記載: (1) 実装した内容の要約, (2) 未解決の課題や技術的負債, (3) 次フェーズへの申し送り事項, (4) LESSONS.md から特に重要な知見。**tasks.md には新しい未完了タスクを追加しない**（Ralph Loopは未完了タスク消滅により自動停止する。次フェーズのタスク作成はチャットでのレビューを経て行う）

# Phase 3 & 3.5 タスク（Phase 2-G 以降に配置）

## Phase 3-A: 仙獣基本型定義（senju/）

- [x] `senju/doc.go`: パッケージドキュメント。仙獣の種族定義・配置・成長・行動AIを扱うパッケージであることを記述
- [x] `senju/species.go`: Species 構造体（ID string, Name string, Element types.Element, BaseHP int, BaseATK int, BaseDEF int, BaseSPD int, GrowthRate float64, MaxBeasts int（この種族が1部屋に配置できる上限のヒント値）, Description string）
- [x] `senju/species_registry.go`: SpeciesRegistry（map管理）。LoadSpecies(data []byte) error（JSONから一括ロード）、Get(id string) (*Species, error)、All() []*Species
- [x] `senju/species_data.json`: 初期仙獣種族5種の定義（各五行属性に1種）。木: 翠龍（バランス型）、火: 炎鳳（攻撃型）、土: 岩亀（防御型）、金: 金狼（速度型）、水: 水蛇（回復型）。各種族に基本ステータスとGrowthRateを設定
- [x] `senju/species_test.go`: SpeciesRegistryのJSONロードテスト、全5種が取得できるテスト、存在しないIDのエラーテスト

## Phase 3-B: 部屋の仙獣容量（world/ 拡張）

- [x] `world/room_type.go` 拡張: RoomType に `MaxBeasts int` フィールドを追加（仙獣部屋: 3, 龍穴: 1, 罠部屋: 2, その他: 0）
- [x] `world/room_type_data.json` 更新: 全部屋タイプに MaxBeasts 値を追加
- [x] `world/room.go` 拡張: Room に `BeastIDs []int` フィールドを追加。BeastCount() int、HasBeastCapacity(roomType) bool メソッド
- [x] `world/room_test.go` 追加: BeastCapacity 関連テスト（容量内追加OK、容量超過NG、仙獣配置不可の部屋タイプNG）

## Phase 3-C: 仙獣インスタンスと配置（senju/）

- [x] `senju/beast.go`: Beast 構造体（ID int, SpeciesID string, Name string, Element types.Element, RoomID int, Level int, EXP int, HP int, MaxHP int, ATK int, DEF int, SPD int, BornTick types.Tick, State BeastState）。BeastState 型（Idle/Patrolling/Chasing/Fighting/Recovering）。NewBeast(id, species, tick) *Beast で基本ステータスをSpeciesから初期化
- [x] `senju/combat_stats.go`: CombatStats 構造体（HP int, ATK int, DEF int, SPD int）。Beast.CalcCombatStats(roomChi *fengshui.RoomChi) CombatStats — 基本ステータスに属性相性バフを適用した実効戦闘ステータスを返す。相生部屋: ATK/DEF/SPD ×1.3、同属性: ×1.1、中立: ×1.0、相克: ×0.7
- [x] `senju/affinity.go`: RoomAffinity(beastElement, roomElement types.Element) float64 — 仙獣と部屋の属性相性倍率を返す。GrowthAffinity(beastElement, roomElement types.Element) float64 — 成長速度への倍率（RoomAffinityと同じ値だが将来独立調整できるよう分離）
- [x] `senju/placement.go`: PlaceBeast(beast, room, roomType) error — 配置可能判定（部屋存在、MaxBeasts未達、仙獣配置可能な部屋タイプ）→配置実行（beast.RoomID設定、room.BeastIDs追加）。RemoveBeast(beast, room) error — 除去。MoveBeast(beast, fromRoom, toRoom, toRoomType) error — 部屋間移動
- [x] `senju/placement_test.go`: 正常配置テスト、容量超過拒否テスト、配置不可部屋タイプ拒否テスト、移動テスト、CalcCombatStatsの相性倍率テスト

## Phase 3-D: 成長システム（senju/）

- [x] `senju/growth_params.go`: GrowthParams 構造体（BaseEXPPerTick int, LevelUpThreshold func(level int) int をJSON対応のため LevelUpBase int + LevelUpPerLevel int に分解、ChiConsumptionPerTick float64, MaxLevel int）。DefaultGrowthParams()。LoadGrowthParams(data []byte) (*GrowthParams, error)
- [x] `senju/growth_params_data.json`: デフォルト成長パラメータ（基本EXP/tick: 10, レベルアップ基礎値: 100, レベルアップ係数: 50, 気消費/tick: 2.0, 最大レベル: 50）。レベルアップ必要EXP = LevelUpBase + LevelUpPerLevel × currentLevel
- [x] `senju/growth.go`: GrowthEngine 構造体。NewGrowthEngine(params, speciesRegistry)。Tick(beasts []*Beast, roomChi map[int]*fengshui.RoomChi, rooms map[int]*world.Room) []GrowthEvent — 1ティック分の成長処理:
  1. 各仙獣の部屋の RoomChi から ChiConsumptionPerTick 分の気を消費試行
  2. 気が足りなければ GrowthEvent{Type: ChiStarved} を返して成長スキップ
  3. BaseEXPPerTick × GrowthAffinity × Species.GrowthRate のEXPを獲得
  4. EXP が LevelUpThreshold に到達したらレベルアップ → GrowthEvent{Type: LevelUp}
  5. MaxLevel クランプ
- [x] `senju/growth.go`: GrowthEvent 構造体（BeastID int, Type GrowthEventType, OldLevel int, NewLevel int, EXPGained int）。GrowthEventType: EXPGained/LevelUp/ChiStarved
- [x] `senju/growth_test.go`: 基本成長（1ティックでEXP獲得）テスト、気不足で成長停止テスト、レベルアップテスト、属性相性による成長速度変化テスト、MaxLevelクランプテスト、気の実消費量がRoomChiに反映されるテスト、GrowthEvent生成テスト

## Phase 3-E: 仙獣シリアライズ（senju/）

- [x] `senju/serialization.go`: MarshalBeasts(beasts []*Beast) ([]byte, error) / UnmarshalBeasts(data []byte, speciesRegistry) ([]*Beast, error) — 全仙獣の状態を保存/復元。SpeciesIDからElementやName等を復元
- [x] `senju/serialization_test.go`: 保存→復元→等価検証テスト、空リストの保存/復元テスト、存在しないSpeciesIDのエラーハンドリングテスト

## Phase 3-F: ASCII可視化への仙獣レイヤー追加

- [x] `senju/ascii.go`: RenderBeastOverlay(cave, beasts) string — CaveのASCII表示に仙獣の配置をオーバーレイ。部屋内に仙獣の属性を示す文字を表示（木:W, 火:F, 土:E, 金:M, 水:A）。仙獣数が2以上の部屋は数字+属性（例: `2F`）
- [x] `cmd/caveviz/main.go` 更新: `--beasts` フラグで仙獣レイヤー表示。`--all` で全レイヤー（通常+気+仙獣）を重ねて表示
- [x] `senju/ascii_test.go`: 小さなCaveで仙獣オーバーレイの出力テスト

## Phase 3-G: 統合検証

- [x] `senju/integration_test.go`: Cave（部屋3つ、属性: 木/火/水）+ ChiFlowEngine（気供給あり）+ 仙獣3体（翠龍を木部屋、炎鳳を火部屋、水蛇を土部屋（相克））を配置→30ティック成長シミュレーション→相性のよい木・火部屋の仙獣はレベルが高く、相克の水蛇はレベルが低いことを検証→気の消費がRoomChiに反映されていることを検証→CalcCombatStatsが配置部屋に応じて変動することを検証
- [x] `go vet ./...` と `go test -race ./...` がクリーンに通ることを確認
- [x] Phase 3 完了。DECISIONS.md 更新、PHASE_COMPLETE 更新、次フェーズドラフトを `tasks_phase3_5_draft.md` として生成。**tasks.md には新しい未完了タスクを追加しない**

---

## Phase 3.5-A: 行動パターン定義（senju/）

- [x] `senju/behavior.go`: BehaviorType 型（Guard/Patrol/Chase/Flee）。Behavior インターフェース — DecideAction(beast, context BehaviorContext) Action。BehaviorContext 構造体（Beast の現在位置、部屋情報、隣接部屋の仙獣/侵入者情報、RoomChi）
- [x] `senju/action.go`: Action 構造体（Type ActionType, TargetRoomID int, TargetBeastID int）。ActionType 型（Stay/MoveToRoom/Attack/Retreat）。仙獣が1ティックで取れる行動を表現
- [x] `senju/behavior_guard.go`: GuardBehavior — 定点防衛。自分の配置部屋に留まる。侵入者が同じ部屋にいればAttack、いなければStay。最もシンプルなAI
- [x] `senju/behavior_guard_test.go`: 侵入者なし→Stay、侵入者あり→Attack のテスト

## Phase 3.5-B: 巡回と追跡AI（senju/）

- [x] `senju/behavior_patrol.go`: PatrolBehavior — 巡回。隣接部屋を順番に移動する。巡回経路は配置部屋を起点に隣接グラフから生成。侵入者を発見したらChaseに遷移
- [x] `senju/behavior_chase.go`: ChaseBehavior — 追跡。発見した侵入者の方向へ隣接部屋を移動。侵入者と同じ部屋に入ったらAttack。一定ティック追跡して見失ったらPatrolに戻る
- [x] `senju/behavior_flee.go`: FleeBehavior — 逃走。HPが一定割合（25%）以下になったら発動。侵入者から最も遠い隣接部屋へ移動。回復室に到達したら回復状態（Recovering）に遷移
- [x] `senju/behavior_test.go`: Patrol の巡回経路テスト、Chase の追跡方向テスト、Flee のHP閾値判定テスト、Flee が回復室を目指すテスト

## Phase 3.5-C: 行動エンジン（senju/）

- [x] `senju/behavior_engine.go`: BehaviorEngine 構造体。NewBehaviorEngine(cave, adjacencyGraph)。AssignBehavior(beast, behaviorType) — 仙獣に行動パターンを割り当て。Tick(beasts, invaderPositions map[int][]int, roomChi) []BeastAction — 全仙獣の行動を1ティック分一括決定:
  1. 各仙獣の現在Behaviorに基づいてDecideAction呼び出し
  2. HP閾値チェック → Flee への自動遷移
  3. 行動の衝突解決（同じ部屋への移動は先着順）
  4. BeastAction リストを返す（実際の移動・攻撃の適用は呼び出し側）
- [x] `senju/behavior_engine.go`: BeastAction 構造体（BeastID int, Action Action, PreviousRoomID int, ResultingState BeastState）。行動の結果を記録
- [x] `senju/behavior_engine.go`: ApplyActions(beasts, rooms, actions) error — BeastAction リストを適用して仙獣の位置・状態を更新
- [x] `senju/behavior_engine_test.go`: 全仙獣Guard→侵入者なしで全員Stay テスト、Patrol仙獣が隣接部屋を巡回するテスト、侵入者発見→Chase遷移テスト、HP低下→Flee遷移テスト、行動の衝突解決テスト

## Phase 3.5-D: 仙獣AI パラメータ外出し

- [x] `senju/behavior_params.go`: BehaviorParams 構造体（FleeHPThreshold float64, ChaseTimeoutTicks int, PatrolRestTicks int）。DefaultBehaviorParams()。LoadBehaviorParams(data []byte)
- [x] `senju/behavior_params_data.json`: デフォルト行動パラメータ（逃走HP閾値: 0.25, 追跡タイムアウト: 10ティック, 巡回時の部屋滞在: 3ティック）
- [x] `senju/behavior_params_test.go`: JSONロードテスト、デフォルト値テスト

## Phase 3.5-E: ASCII可視化への行動レイヤー追加

- [x] `senju/ascii.go` 更新: 仙獣の行動状態を表示に反映。Guard: `[G]`, Patrol: `[P]`, Chase: `[!]`, Flee: `[←]`, Recovering: `[+]`。侵入者位置のプレースホルダー表示（`??` — Phase 4で本実装）
- [x] `cmd/caveviz/main.go` 更新: `--ai` フラグで行動状態レイヤー表示

## Phase 3.5-F: 統合検証

- [x] `senju/ai_integration_test.go`: Cave（部屋5つ、通路接続）+ 仙獣3体（Guard×1, Patrol×1, Chase×1）+ 疑似侵入者位置（map[int][]int で手動設定）→20ティック行動シミュレーション→Guard仙獣は配置部屋に留まることを検証→Patrol仙獣は複数部屋を巡回することを検証→侵入者位置を設定するとChase仙獣が追跡方向に移動することを検証→HP低下でFleeに遷移することを検証
- [x] `go vet ./...` と `go test -race ./...` がクリーンに通ることを確認
- [x] Phase 3.5 完了。DECISIONS.md 更新、PHASE_COMPLETE 更新、次フェーズドラフトを `tasks_phase4_draft.md` として生成。**tasks.md には新しい未完了タスクを追加しない**