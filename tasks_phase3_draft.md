# tasks_phase3_draft.md — Phase 3: 仙獣システム（senju/）

<!-- Phase: 3 (senju/) — 仙獣システム -->

## Phase 3-A: 仙獣基本型定義（senju/）

- [ ] `senju/doc.go`: パッケージドキュメント。仙獣の定義・配置・成長・進化を扱うパッケージであることを記述
- [ ] `senju/species.go`: Species 構造体（ID string, Name string, Element types.Element, BaseHP int, BaseATK int, BaseDEF int, BaseSPD int, GrowthRate float64, Description string）。仙獣の種族定義
- [ ] `senju/species_registry.go`: SpeciesRegistry（map管理、JSONから一括ロード）。LoadSpecies(data []byte) error、Get(id string) (*Species, error)
- [ ] `senju/species_data.json`: 初期仙獣種族5種の定義（各五行属性に1種ずつ）。木: 翠龍（バランス型）、火: 炎鳳（攻撃型）、土: 岩亀（防御型）、金: 金狼（速度型）、水: 水蛇（回復型）
- [ ] `senju/species_test.go`: SpeciesRegistryのJSONロードテスト、存在しないIDのエラーテスト

## Phase 3-B: 仙獣インスタンスと配置（senju/）

- [ ] `senju/senju.go`: Senju 構造体（ID int, SpeciesID string, Element types.Element, RoomID int, Level int, EXP int, HP int, ATK int, DEF int, SPD int, BornTick types.Tick）。NewSenju(id, species, roomID, tick) *Senju。CurrentStats() でレベルに応じたステータスを返す
- [ ] `senju/placement.go`: CanPlace(senju, room, cave) error — 配置可能判定（部屋が存在するか、部屋に空きがあるか）。Place(senju, room) — 配置実行（RoomIDセット）。Remove(senju) — 部屋から除去
- [ ] `senju/affinity.go`: Affinity(senjuElement, roomElement) float64 — 仙獣と部屋の属性相性倍率を返す（相生: 1.3, 同属性: 1.1, 中立: 1.0, 相克: 0.7）。成長速度・戦闘力に影響
- [ ] `senju/senju_test.go`: Senju生成テスト、配置テスト（正常/部屋不存在/重複配置）、属性相性倍率テスト

## Phase 3-C: 成長システム（senju/）

- [ ] `senju/growth_params.go`: GrowthParams 構造体（BaseEXPPerTick int, LevelUpEXP func(level int) int, ChiConsumptionPerTick float64, MaxLevel int）。DefaultGrowthParams() で初期値。JSON読み込み対応
- [ ] `senju/growth_params_data.json`: デフォルト成長パラメータ（基本EXP/tick: 10, レベルアップEXP: level*100, 気消費/tick: 2.0, 最大レベル: 50）
- [ ] `senju/growth.go`: GrowthEngine 構造体。NewGrowthEngine(params, registry) で生成。Tick(senjus []*Senju, roomChi map[int]*fengshui.RoomChi) — 1ティック分の成長処理:
  1. 各仙獣の部屋から気をChiConsumptionPerTick分消費（気が足りなければ成長しない）
  2. 属性相性倍率をEXP獲得量に適用
  3. EXPがLevelUpEXP(currentLevel)に到達したらレベルアップ（ステータス再計算）
  4. MaxLevelクランプ
- [ ] `senju/growth_test.go`: 基本成長テスト（1ティックでEXP獲得）、気不足で成長停止テスト、レベルアップテスト、属性相性による成長速度変化テスト、最大レベルクランプテスト

## Phase 3-D: 進化システム（senju/）

- [ ] `senju/evolution.go`: Evolution 構造体（FromSpeciesID string, ToSpeciesID string, RequiredLevel int, RequiredElement types.Element（部屋属性条件、ゼロ値は条件なし）、RequiredChiRatio float64（部屋の気充填率条件））
- [ ] `senju/evolution_registry.go`: EvolutionRegistry（進化ルートの管理）。LoadEvolutions(data []byte) error。GetEvolutions(speciesID string) []Evolution — 該当種族の進化先一覧。CheckEvolution(senju, room, roomChi) *Evolution — 進化条件を満たす最初の進化を返す（nil=条件未達）
- [ ] `senju/evolution_data.json`: 初期進化ルート定義（各種族に1段階進化を1つ）。例: 翠龍Lv15→蒼龍、炎鳳Lv15→朱雀
- [ ] `senju/evolution.go`: Evolve(senju, evolution, registry) error — 進化実行（SpeciesID変更、ステータス再計算、レベルは維持）
- [ ] `senju/evolution_test.go`: 進化条件チェックテスト（レベル条件/属性条件/気充填率条件）、進化実行テスト、条件未達で進化しないテスト、進化後のステータス変更テスト

## Phase 3-E: 仙獣シリアライズ（senju/）

- [ ] `senju/serialization.go`: MarshalSenjus(senjus []*Senju) ([]byte, error) / UnmarshalSenjus(data []byte) ([]*Senju, error) — 全仙獣の状態を保存/復元
- [ ] `senju/serialization_test.go`: 保存→復元→等価検証テスト、空リストの保存/復元テスト

## Phase 3-F: 統合検証

- [ ] `senju/integration_test.go`: Cave（部屋3つ）+ ChiFlowEngine（気供給あり）+ 仙獣3体（異なる属性）を用意→20ティック成長シミュレーション→属性相性のよい部屋の仙獣ほどレベルが高いことを検証→進化条件を満たした仙獣が進化することを検証
- [ ] `go vet ./...` と `go test -race ./...` がクリーンに通ることを確認
- [ ] Phase 3 完了タスク（DECISIONS.md更新、PHASE_COMPLETE更新、次フェーズドラフト生成）
