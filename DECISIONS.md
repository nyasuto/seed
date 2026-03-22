# DECISIONS.md — 設計判断の記録

> ステータス凡例:
> - **OPEN**: 未解決。将来のフェーズで対応が必要
> - **ACTIVE**: 現在有効な設計原則。変更予定なし
> - **RESOLVED**: 後続フェーズで対応済み

---

## D001: OnCaveChanged の差分更新は不要

**ステータス**: RESOLVED
**日付**: 2026-03-21
**フェーズ**: Phase 2-C → Phase 7-G で解決

**判断**: `ChiFlowEngine.OnCaveChanged` は龍脈を全再計算する実装とした。差分更新（変更された部分のみ再計算）は行わない。

**理由**:
- 現時点では部屋数・龍脈数が少なく（数十オーダー）、全再計算でもパフォーマンスに問題がない
- 差分更新は「どの龍脈が影響を受けたか」の追跡が必要で、複雑さが大幅に増す

**影響範囲**: `fengshui/chi_flow.go` の `OnCaveChanged` メソッド

**Phase 7-G ベンチマーク結果** (Apple M4 Pro, Go 1.25):

| ケース | 実行時間 | メモリ | アロケーション |
|--------|---------|--------|--------------|
| OnCaveChanged/Rooms5 | 4.5μs | 6.7KB | 37 |
| OnCaveChanged/Rooms10 | 9.6μs | 13.5KB | 70 |
| OnCaveChanged/Rooms20 | 20.1μs | 27.2KB | 133 |
| OnCaveChanged/Rooms50 | 55.3μs | 85.4KB | 318 |
| FullTick (5部屋, 100ticks) | 663μs | 888KB | 13906 |

**結論**: OnCaveChanged は部屋数に対して線形（O(n)）でスケールし、50部屋でも55μs。ゲームの想定最大部屋数（20〜30程度）では20μs以下であり、60FPSのフレーム予算（16.7ms）の0.1%にも満たない。差分更新の導入は不要と判断する。全再計算のまま維持する

---

## D002: コアゲーム体験の定義 — 「常に不完全な状態での判断」

**ステータス**: ACTIVE（全フェーズに影響する設計原則）
**日付**: 2026-03-21
**フェーズ**: Phase 3〜（全フェーズ横断）

**判断**: カオスシードの面白さの核心は「完璧なダンジョンを完成させること」ではなく、「常に不完全な状態で判断を迫られ、80%の妥協で噛み合わせるプロセス」にある。

### 3原則と実装状況

| 原則 | 内容 | 実装状況 |
|---|---|---|
| 原則1: 不完全性の強制 | 地形のランダム制約で理想配置が不可能 | **IMPLEMENTED** — Phase 6-C/6-H で HardRock/Water セル、TerrainGenerator、ValidateTerrain（詰み防止）を実装 |
| 原則2: 時間圧力 | 侵入波がプレイヤーの準備完了を待たない | **IMPLEMENTED** — Phase 4 で WaveSchedule 基盤、Phase 6-G で DynamicWaveScheduler + CalcFirstWaveTiming を実装 |
| 原則3: トレードオフの連続 | リソースが常に不足 | **IMPLEMENTED & VERIFIED** — Phase 5 で経済構造実装。Phase 7-H で定量検証完了（200ティックシミュレーションで全MAX到達率0%、供給2倍でもアンチパターン非発生を確認） |

### アンチパターン（避けるべき設計）
- 十分な時間があれば完璧なダンジョンが完成する → ゲームの死
- 最適解が1つに収束し、毎回同じ配置が最強 → リプレイ性の喪失
- 侵入者が弱すぎて構築の邪魔にならない → 判断の必要性が消滅
- リソースが潤沢で全部に投資できる → トレードオフの消滅

### 検証方法（Phase 7-H で実施）
1. AIプレイヤーによる数千ゲームの自動実行で、「全部屋MAX」到達率がゲームクリア前に0%に近いこと
2. 同じシナリオでも地形RNGの違いで最適配置が変わること
3. 侵入波の50%以上が「構築中」に到達すること

### 面白さの構造分解（プレイヤー体験の観点）

| 面白さの種類 | 内容 | 壊れるサイン | 調整の方向 |
|---|---|---|---|
| 1. 計画する面白さ | 地形と龍脈を見て理想の構想を練る | 計画前に敵が来る | 最初の侵入波までの猶予を伸ばす |
| 2. 妥協する面白さ | 制約の中で80点の解を見つける | 常に理想配置が可能 / 常に詰む | 地形制約の密度を調整 |
| 3. 中断される面白さ | 構築中に防衛が割り込む | 構築と防衛が完全に交互 | 侵入波を構築中に重ねる |
| 4. 噛み合う面白さ | 不完全な防衛が創発的に成功する | 常に圧勝 / 常に惨敗 | 侵入者の強さとリソース供給を調整 |
| 5. 完璧が来ない面白さ | クリア時もダンジョンは未完成 | クリア前に全MAX到達 | リソース総量を絞る / クリア条件を早める |

### 面白さの連鎖モデル

```
計画する → 制約に出会う → 妥協する → 構築を始める
    ↑                                       ↓
    └──── 「次こそは」のリプレイ動機 ←── 防衛で中断される
                                            ↓
                                    不完全な状態で噛み合わせる
                                            ↓
                                    撃退成功のカタルシス
                                            ↓
                                    でも完璧ではない → 次の計画へ
```

---

## D003: 仙獣の気消費は GrowthEngine 側で RoomChi.Current を直接減算

**ステータス**: ACTIVE（D009で文脈が明確化）
**日付**: 2026-03-21
**フェーズ**: Phase 3-D

**判断**: 仙獣の成長時の気消費は、ChiFlowEngine に消費APIを追加するのではなく、GrowthEngine が `RoomChi.Current` を直接減算する方式とした。

**理由**:
- ChiFlowEngine は気の「流れ」（供給・伝播・減衰）に責務を持ち、消費は別レイヤーの関心事
- RoomChi.Current は公開フィールドであり、外部からの操作を許容する設計
- 将来的に仙獣以外（罠の維持、部屋の強化等）も気を消費するため、消費元ごとに自分で減算する方式が拡張しやすい

**D009との関係**: これは「物理層」（部屋レベルの気）の消費。「経済層」（ChiPool）の仙獣維持コストとは別。仙獣は「部屋の気を食べて成長する（物理層、本D003）」と「存在するだけでChiPoolから維持費がかかる（経済層、Phase 5）」の二重構造。

**影響範囲**: `senju/growth.go` の Tick メソッド

---

## D004: 仙獣の戦闘ステータスは配置時ではなくクエリ時に計算

**ステータス**: RESOLVED（Phase 4で確認済み）
**日付**: 2026-03-21
**フェーズ**: Phase 3-C

**判断**: `Beast.CalcCombatStats(roomChi)` はキャッシュせず、呼び出し時に毎回計算する方式とした。

**理由**:
- RoomChi は毎ティック変動するため、キャッシュの無効化タイミングが複雑になる
- 戦闘ステータスの計算は乗算数回で非常に軽量
- 戦闘解決時にのみ呼ばれるため、頻度も低い

**解決確認**: Phase 4のResolveRoomCombatで実際に使用され、パフォーマンス問題なし。設計として定着。

**影響範囲**: `senju/combat_stats.go`

---

## D005: 仙獣AIの行動衝突は先着順で解決

**ステータス**: ACTIVE
**日付**: 2026-03-21
**フェーズ**: Phase 3.5-C

**判断**: BehaviorEngine の Tick メソッドで、複数の仙獣が同じ部屋に同時移動しようとした場合、先着順（beasts スライスの順序）で解決する方式とした。

**理由**:
- 優先度ベースは決定論的再現が難しくなる
- 先着順はシンプルで決定論的
- 衝突が頻発する状況は実際のゲームプレイでは稀

**Phase 4での再評価結果**: Phase 4の侵入者移動はInvasionEngine側で独立に処理されるため、仙獣と侵入者の移動衝突は発生しない。D005は仙獣同士の衝突にのみ適用され、当初の想定通り問題なし。

**影響範囲**: `senju/behavior_engine.go` の Tick メソッド

---

## D006: 仙獣の行動AIは侵入者位置をmap[int][]intで受け取る

**ステータス**: RESOLVED（Phase 4で実装完了）
**日付**: 2026-03-21
**フェーズ**: Phase 3.5-A

**判断**: BehaviorEngine.Tick の侵入者位置引数は `map[int][]int`（部屋ID → 侵入者IDリスト）とした。

**理由**:
- senju パッケージが invasion パッケージに依存しない（循環依存防止）
- 侵入者の「どの部屋にいるか」だけが行動判定に必要

**解決確認**: Phase 4で `InvasionEngine.BuildInvaderPositions()` が実装され、設計意図通りに機能。

**影響範囲**: `senju/behavior_engine.go`, `invasion/engine.go`

---

## D007: D002時間圧力のPhase 4段階の実装

**ステータス**: RESOLVED（Phase 6で残作業消化）
**日付**: 2026-03-21
**フェーズ**: Phase 4, Phase 6

**判断**: D002の「時間圧力」原則をPhase 4で基盤実装し、Phase 6でシナリオレベルの調整を完了。

**Phase 4での実装**:
1. WaveSchedule による固定タイミングの侵入波
2. テスト用スケジュールの早期侵入（tick 50）
3. 難易度のエスカレーション

**Phase 6での解決**:
- DynamicWaveScheduler: シナリオごとの波間隔チューニング
- CalcFirstWaveTiming: 初期リソースと建設コストから「構築が十分でないタイミング」を自動算出
- D002「面白さ3: 中断される面白さ」のシナリオレベル設計基盤が確立

**影響範囲**: `invasion/wave_schedule.go`, `scenario/wave_schedule.go`

---

## D008: 戦闘マッチングは素早さ順ペアリング・1ティック1ラウンド

**ステータス**: ACTIVE
**日付**: 2026-03-21
**フェーズ**: Phase 4-C

**判断**:
1. 全戦闘参加者をSPD降順でソート→順序ペアリング
2. 余剰はそのラウンドでは戦闘なし（フリー）
3. 1ティック1ラウンド

**理由**:
- O(N log N)で決定論的
- 数的優位が「攻撃を受けない」メリットになり防衛配置の戦略性が生まれる
- ティック間で移動・撤退の判断ができ、D002の「判断の連続」体験を維持

**影響範囲**: `invasion/combat.go` の `ResolveRoomCombat`

---

## D009: ChiPool（経済層）と ChiFlowEngine（物理層）の二層構造

**ステータス**: ACTIVE
**日付**: 2026-03-21
**フェーズ**: Phase 5

**判断**: 気のシステムを2つのレイヤーに分離する。

| レイヤー | 管轄 | 実体 | 用途 |
|---|---|---|---|
| 物理層 | ChiFlowEngine (Phase 2) | 各部屋の RoomChi.Current | 龍脈からの供給、隣接部屋間の伝播、減衰、仙獣の成長消費（D003） |
| 経済層 | ChiPool (Phase 5) | プレイヤーの「通貨」 | 建設・召喚・強化の支払い、侵入報酬、維持コスト |

**レイヤー間の接続**:
- SupplyCalculator が毎ティック、物理層の状態（全部屋の気充填率 + 風水スコア）を読み取り、経済層の ChiPool への供給量に変換する
- 物理層が豊かであるほど経済層も潤う、という間接的な関係
- 物理層の消費（D003: 仙獣が部屋の気を食べる）と経済層の消費（仙獣の維持費）は独立

**理由**:
- 物理層はシミュレーションの正確性が重要（気の流れ、属性相性）
- 経済層はゲームバランスの調整容易性が重要（JSON外出しのパラメータで調整）
- 二層に分離することで、物理層のパラメータ変更が直接経済を破壊せず、SupplyCalculatorの変換式で吸収できる
- D002原則3「トレードオフの連続」は経済層で保証する

**影響範囲**: `fengshui/` (物理層), `economy/` (経済層), `economy/supply.go` (接続点)

---

## D010: CoreHP — 龍穴コアの耐久値

**ステータス**: ACTIVE
**日付**: 2026-03-21
**フェーズ**: Phase 6

**判断**: 龍穴（コア部屋）にHP概念を追加する。侵入者がコア部屋に到達して攻撃するとCoreHPが減少し、0以下でゲーム敗北。

**設計**:
1. `RoomType.BaseCoreHP` にベース値を定義（龍穴のみ非ゼロ: 100）
2. `RoomType.CoreHPAtLevel(level)` で `BaseCoreHP * level` を計算
3. `Room.CoreHP` に現在値を保持
4. `GameSnapshot.CoreHP` で条件評価器に読み取り専用で提供

**理由**:
- 一定ティック滞在ではなく攻撃ベースにすることで、仙獣による防衛の意味が増す（侵入者を殴って追い返せばCoreHPは減らない）
- 部屋のレベルアップとCoreHPが自然に連動する

**影響範囲**: `world/room_type.go`, `world/room.go`, `scenario/progress.go`, `scenario/conditions.go`

---

## D011: EventCommand パターン — イベントは状態変更ではなくコマンドを返す

**ステータス**: ACTIVE
**日付**: 2026-03-21
**フェーズ**: Phase 6

**判断**: イベントシステムは直接ゲーム状態を変更せず、`EventCommand` インターフェースを返す。実際の状態変更は Phase 7 の simulation 層の CommandExecutor が担当する。

**設計**:
1. `EventCommand` インターフェース: `Execute() string` で人間可読な説明を返す
2. 4つのコマンド型: SpawnWaveCommand, ModifyChiCommand, ModifyConstraintCommand, MessageCommand
3. `NewCommand(def CommandDef)` ファクトリでJSON定義からコマンドを生成
4. `EventEngine.Tick()` は `GameSnapshot`（読み取り専用）を受け取り、条件を評価して `[]EventCommand` を返す

**理由**:
- 関心の分離: イベントは「何をすべきか」を宣言し、「どう実行するか」は simulation 層に委ねる
- 決定論性: 条件評価は純粋関数（GameSnapshot → bool）で副作用なし
- 監査性: 全コマンドの人間可読ログを提供
- 拡張性: ファクトリにコマンド型を追加するだけで新しいイベントアクションを追加可能
- データ駆動: イベント定義はJSONから読み込み、コード変更なしにシナリオを追加可能

**影響範囲**: `scenario/command.go`, `scenario/event_engine.go`, `simulation/executor.go`

---

## D012: ティック更新順序の厳密な定義

**ステータス**: ACTIVE
**日付**: 2026-03-21
**フェーズ**: Phase 7-C

**判断**: SimulationEngine.Step() の1ティック内の更新順序を以下の12ステップで固定する。

```
1.  PlayerAction 実行     — プレイヤーの操作（部屋掘削、仙獣配置等）
2.  ChiFlowEngine.Tick    — 気の供給・伝播・減衰
3.  GrowthEngine.Tick     — 仙獣成長（気を消費）
4.  DefeatProcessor       — Stunned仙獣の復活チェック
5.  EvolutionEngine       — 進化条件チェック → 実行
6.  BehaviorEngine.Tick   — 仙獣行動AI
7.  InvasionEngine.Tick   — 侵入者移動・戦闘・CoreHPダメージ
8.  InvasionEconomyProcessor — 報酬・損失のChiPool反映
9.  EconomyEngine.Tick    — 供給・維持・赤字処理
10. EventEngine.Tick      — 条件評価 → EventCommand生成
11. CommandExecutor.Apply  — EventCommandの状態反映
12. 勝利/敗北条件評価     — ゲーム終了判定
```

**理由**:
- PlayerAction が最初: プレイヤーの操作結果がそのティックの全サブシステムに即座に反映される（「部屋を掘った直後に龍脈が変わる」体験）
- 気→成長→行動→侵入の順: 因果連鎖の自然な流れ
- 経済が侵入の後: 侵入の報酬/損失が同ティックの経済計算に含まれる
- イベント→条件評価が最後: そのティックの全結果を反映した上で判定
- **順序を変更すると決定論的再現が壊れるため、この順序は不変とする**

**影響範囲**: `simulation/engine.go` の Step メソッド

---

## D013: PlayerAction は Step() の引数として毎ティック注入

**ステータス**: ACTIVE
**日付**: 2026-03-21
**フェーズ**: Phase 7-A

**判断**: プレイヤー操作は SimulationEngine 内部に状態として持たず、Step(actions []PlayerAction) の引数として毎ティック外部から注入する。

**理由**:
- AIプレイヤーと人間入力が同じインターフェースで動く
- リプレイ時はアクション列をファイルから読んでStep()に渡すだけで再現可能
- SimulationEngine 自体はプレイヤーの意思決定を一切持たない（純粋なシミュレーター）
- Run() メソッドは actionProvider 関数を受け取り、毎ティック GameSnapshot を渡してアクションを得る

**代替案の棄却**:
- ActionQueue方式（事前にアクションを積む）: ティックごとの状態を見てから判断できない
- Observer方式（イベントでアクションをフック）: 更新順序との整合性が取りにくい

**影響範囲**: `simulation/engine.go`, `simulation/action.go`, `simulation/runner.go`

---

## D014: MaxRooms制約のバリデーション追加

**ステータス**: RESOLVED
**日付**: 2026-03-22
**フェーズ**: Phase 1-A

**判断**: `simulation/action.go` の `validateDigRoom` に `GameConstraints.MaxRooms` チェックを追加。現在の部屋数が MaxRooms 以上なら DigRoom アクションを拒否する。MaxRooms が 0 の場合は制約なし（無制限）として扱う。

**理由**:
- シナリオごとに部屋数上限を設定できるようにすることで、D002原則1「不完全性の強制」を強化
- SimpleAIPlayer も MaxRooms を参照し、上限到達時は DigRoom を試行しないよう修正

**影響範囲**: `simulation/action.go`, `simulation/ai_player.go`

---

## D015: SimpleAIPlayerコリドー戦略

**ステータス**: RESOLVED
**日付**: 2026-03-22
**フェーズ**: Phase 1-B

**判断**: SimpleAIPlayer の DecideActions で、DigRoom 成功後の次ティックに隣接部屋への DigCorridor を自動発行する戦略を追加。龍穴からの気の伝播経路を確保する。

**理由**:
- AI が部屋を掘るだけで通路を掘らないと、気が伝播せず新部屋が活用されない
- 通路掘りロジックにより、AIプレイヤーが最低限の機能するダンジョンを構築できるようになる
- 通路が掘れない場合（壁、距離等）はスキップし、エラーにしない

**影響範囲**: `simulation/ai_player.go`

---

## D016: caveScore正規化（呼び出し側）

**ステータス**: RESOLVED
**日付**: 2026-03-22
**フェーズ**: Phase 1-C

**判断**: `simulation/engine.go` の Step メソッドで、`CaveTotal()` の生値をそのまま `CalcTickSupply` に渡していた問題を修正。CaveTotal を MaxPossibleScore（全部屋の理論最大スコア合計）で割って [0,1] に正規化してから渡す。

**理由**:
- `CalcTickSupply` は `caveScore` を [0,1] の範囲として扱う設計だが、CaveTotal は部屋数に比例して増加する生値であり、部屋数が増えると供給量が指数的に膨らむ問題があった
- 正規化により、部屋数が増えても供給量が適切な範囲に収まる
- MaxPossibleScore が 0 の場合（部屋なし）は caveScore = 0.0 として安全に処理

**影響範囲**: `simulation/engine.go`

---

## D017a: AI Mode プロトコル設計 — JSON Lines + valid_actions ホワイトリスト

**ステータス**: ACTIVE
**日付**: 2026-03-22
**フェーズ**: sim Phase 3

**判断**: AI Mode は JSON Lines プロトコルで外部プログラムと通信する。各ティックで StateMessage（ゲーム状態 + valid_actions）を送信し、クライアントは valid_actions に含まれるアクションのみ実行できる。

**設計**:
1. **valid_actions ホワイトリスト**: サーバーが毎ティック実行可能なアクションを列挙し、クライアントはその中から選択するのみ。不正なアクションは拒否
2. **エラーリトライ**: 不正入力時は ErrorMessage + StateMessage 再送で最大3回リトライ。超過時は wait にフォールバック
3. **タイムアウト**: `--timeout` オプションでアクション入力のタイムアウトを設定可能。タイムアウト時は wait
4. **GameEndMessage**: ゲーム終了時に result (victory/defeat) + summary (統計情報) を送信

**理由**:
- LLM等の外部プログラムが正しいアクションのみ実行できるよう、ホワイトリスト方式を採用
- JSON Lines は1行1メッセージで、パイプ経由の通信に最適
- リトライ機構によりクライアントの軽微なエラーに耐性を持つ

**影響範囲**: `sim/adapter/ai/`, `sim/server/`

---

## D017b: D002検証 — チュートリアルシナリオのBreakageReport

**ステータス**: RESOLVED
**日付**: 2026-03-22
**フェーズ**: Phase 4-G

**判断**: チュートリアルシナリオ × SimpleAI × 1,000ゲームの BreakageReport で、以下3件のアラートが発生する。これらはチュートリアルの意図的な easy 設計に起因し、バランス上の問題ではない。

| メトリクス | 値 | 閾値 | 原因 |
|---|---|---|---|
| B03 (terrain block rate) | 0.0 | > 0.05 | terrain_density=0.05、意図的に制約が弱い |
| B05 (wave overlap) | 0.0 | > 0.30 | 波が1回（tick 100）のみ、建設と重ならない |
| B11 (surplus rate) | ~0.92 | < 0.50 | starting_chi=200 + 低コスト、リソース潤沢 |

**critical アラート（B04, B06, B07, B08）は0件**であり、ゲームの基本的な健全性は確認済み。

**パフォーマンス**: 1,000ゲーム × SimpleAI が約1〜2秒で完了（5分制限を大幅にクリア）。

**影響範囲**: `sim/adapter/batch/integration_test.go`

---

## 未解決課題サマリー

| ID | 内容 | ステータス |
|---|---|---|
| ~~D001~~ | ~~OnCaveChanged差分更新~~ | **RESOLVED** — 差分更新不要（50部屋55μs、線形スケール） |
| ~~D002原則3~~ | ~~経済バランスの定量検証~~ | **RESOLVED** — Phase 7-H で定量検証完了 |
| ~~D017b~~ | ~~D002検証（チュートリアル）~~ | **RESOLVED** — critical アラート0件、基本的な健全性確認済み |
| 罠の盗賊回避率 | 盗賊のSPDによる罠回避 | 将来拡張（v1.0.0 スコープ外） |
| 侵入者AI高度化 | 複数ステップ先読み、仙獣回避 | 将来拡張（v1.0.0 スコープ外） |
| standardスイープ | standard シナリオ向けスイープ値チューニング | 将来拡張（v1.0.0 スコープ外） |

## Phase 6 最終棚卸し（2026-03-22）

全 D001〜D017b を確認済み。Phase 6 で新たな設計判断は発生せず。
- D001〜D016: ステータス変更なし
- D017a (AI Mode プロトコル): ACTIVE — 設計原則として継続有効
- D017b (D002検証): RESOLVED — Phase 4-G で確認済み
- 未解決2件（罠の盗賊回避率、侵入者AI高度化）は v1.0.0 スコープ外で変更なし

## game Phase 1 棚卸し（2026-03-22）

全 D001〜D017b を確認済み。game Phase 1 で新たな設計判断は発生せず。
- D014 (MaxRooms制約): RESOLVED — Phase 1-A で実装済み
- D015 (AIコリドー戦略): RESOLVED — Phase 1-B で実装済み
- D016 (caveScore正規化): RESOLVED — Phase 1-C で実装済み
- PHASE_COMPLETE_0 の申し送り2件は Phase 1 で解消:
  - GameController 導入 → Phase 1-A で完了
  - main.go の SimulationEngine 直接参照 → Phase 1-E で GameController 経由に移行完了
- game/testdata/tutorial.json の embed 重複は残存（core の LoadBuiltinScenario への統一は Phase 2 以降で検討）