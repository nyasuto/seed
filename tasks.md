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

- [x] `world/serialization.go`: Cave.MarshalJSON() ([]byte, error)、UnmarshalCave(data []byte) (*Cave, error) — Grid全セル + Rooms + Corridors の完全保存/復元
- [x] `world/serialization_test.go`: 部屋と通路を含むCaveを保存→復元→元と等価であることを検証。空のCaveの保存/復元テスト

## Phase 1-G: 統合検証

- [x] `world/integration_test.go`: 中規模マップ（32x32）に部屋5つ配置→全接続→風水フェーズへの受け渡し用データ構造が正しく取れることを確認
- [x] testutil に `testutil/rng.go` を作成: FixedRNG（常に同じ値を返すモックRNG）と NewTestRNG(seed) ヘルパー
- [x] `go vet ./...` と `go test -race ./...` がクリーンに通ることを確認
- [x] Phase 1 完了。tasks.md の Phase 2 タスクを追記する（PRD Phase 2 を参照して fengshui/ のタスクを展開）

## Phase 2-A: 風水基本型定義（fengshui/）

- [ ] `fengshui/doc.go`: パッケージドキュメント。龍脈・気・風水評価を扱うパッケージであることを記述
- [ ] `fengshui/dragon_vein.go`: DragonVein 構造体（ID int, SourcePos types.Pos, Element types.Element, FlowRate float64, Path []types.Pos）。龍脈は洞窟の入口から内部へ気を運ぶ経路
- [ ] `fengshui/chi.go`: RoomChi 構造体（RoomID int, Current float64, Capacity float64, ConsumptionRate float64）。部屋ごとの気の状態を管理。IsFull() bool, IsEmpty() bool, Ratio() float64 メソッド
- [ ] `fengshui/chi_test.go`: RoomChi の基本メソッドテスト（IsFull/IsEmpty/Ratio の境界値テスト）

## Phase 2-B: 龍脈の経路計算（fengshui/）

- [ ] `fengshui/dragon_vein_builder.go`: BuildDragonVein(cave *world.Cave, sourcePos types.Pos, element types.Element, flowRate float64) (*DragonVein, error) — 入口位置から通路・部屋を通って気を運ぶ経路をBFSで計算。RoomsOnPath(cave *world.Cave) []int で龍脈上にある部屋IDリストを返す
- [ ] `fengshui/dragon_vein_test.go`: 龍脈が入口から通路を通って部屋に到達するテスト、到達不能ケースのテスト、複数部屋を経由する龍脈のテスト

## Phase 2-C: 気の蓄積・消費モデル（fengshui/）

- [ ] `fengshui/chi_flow.go`: ChiFlowState 構造体（DragonVeins []DragonVein, RoomChi map[int]*RoomChi）。NewChiFlowState(cave *world.Cave, veins []DragonVein) *ChiFlowState で初期化（各部屋のCapacityはRoomTypeのBaseChiCapacityから取得）。Tick() で1ティック分の気の流れを計算: 龍脈上の部屋にFlowRateに応じて気を分配、各部屋のConsumptionRateを減算、Capacityを超えない・0を下回らないようクランプ
- [ ] `fengshui/chi_flow_test.go`: 1部屋への気の蓄積テスト、容量上限でのクランプテスト、消費による減少テスト、複数ティック経過後の状態テスト、龍脈上にない部屋には気が流れないことのテスト

## Phase 2-D: 風水評価スコア（fengshui/）

- [ ] `fengshui/score.go`: FengShuiScore 構造体（RoomID int, BaseScore float64, AdjacencyBonus float64, DragonVeinBonus float64, TotalScore float64）。部屋単位のスコア内訳を保持
- [ ] `fengshui/evaluator.go`: Evaluator 構造体。NewEvaluator(cave *world.Cave, registry *world.RoomTypeRegistry) *Evaluator。EvaluateRoom(roomID int, veins []DragonVein) (FengShuiScore, error) — 個別部屋の風水スコアを算出。EvaluateAll(veins []DragonVein) ([]FengShuiScore, error) — 全部屋のスコアを一括算出。スコアリングルール: (1) BaseScore = 部屋の気の充填率 × 100, (2) AdjacencyBonus = 隣接部屋との五行相性（相生: +20, 相克: -15, 同属性: +5, その他: 0）の合計, (3) DragonVeinBonus = 龍脈に接続されていれば +30
- [ ] `fengshui/evaluator_test.go`: 単独部屋のスコア計算テスト、相生隣接ボーナステスト、相克隣接ペナルティテスト、龍脈接続ボーナステスト、全部屋一括評価テスト

## Phase 2-E: 風水シリアライズ（fengshui/）

- [ ] `fengshui/serialization.go`: ChiFlowState.MarshalJSON() / UnmarshalChiFlowState() — 龍脈と部屋の気の状態を保存/復元
- [ ] `fengshui/serialization_test.go`: 龍脈と気の状態を含むChiFlowStateの保存→復元→等価検証テスト

## Phase 2-F: 統合検証

- [ ] `fengshui/integration_test.go`: 中規模Cave（32x32）に部屋5つ配置→龍脈2本設定→10ティック気の流れをシミュレーション→風水スコア評価→スコアが相生/相克の配置に応じて正しく変動することを検証
- [ ] `go vet ./...` と `go test -race ./...` がクリーンに通ることを確認
- [ ] Phase 2 完了。tasks.md の Phase 3 タスクを追記する（PRD Phase 3 を参照して senju/ のタスクを展開）