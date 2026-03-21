# chaosseed-sim tasks.md

> RALF Loop 用タスクファイル。1イテレーション1タスク。
> 上から順に消化する。各タスクの完了条件をすべて満たしてから次に進む。
> タスク完了時にチェックボックスを [x] にする。

---

## Phase 0: プロジェクト初期化

- [x] **Task 0-A: simスケルトン作成**

モノレポ構成（go.work, core/, sim/, sim/go.mod, sim/docs/PRD.md）は既に存在する。
以下の残作業を実施する：

1. sim/go.mod に core への require を追加（`require github.com/nyasuto/seed/core v0.0.0` + replace ディレクティブ）
2. `go work sync` で整合確認
3. sim/cmd/chaosseed-sim/main.go のスケルトン作成（`--human`, `--ai`, `--batch`, `--balance`, `--version` フラグのパース、未実装モードは "not implemented" で終了）

**完了条件**:
- `go work sync` が成功
- `cd sim && go build ./...` が成功
- `cd sim && go vet ./...` が成功
- `chaosseed-sim --version` がバージョン文字列を出力
- `chaosseed-sim --human` が "not implemented" で終了

- [x] **Task 0-B: sim用Makefileとプロジェクト基盤**

ルートMakefileは既に存在する（test, test-race, vet, lint, cover, check, clean ターゲット）。
以下を実施する：

1. ルートMakefile に `build` と `all` ターゲットを追加（`all: build check`、`build` は sim バイナリ生成）
2. sim/Makefile 作成（ターゲット: build, test, vet, lint, clean）
3. sim/.golangci.yml 作成（ルートの .golangci.yml と同じルールセット）

**完了条件**:
- `cd sim && make build` が sim バイナリを生成
- `cd sim && make test` が（まだテストはないが）エラーなく完了
- `cd sim && make vet` がクリーン
- ルートの `make all` で core と sim が両方ビルド・テストされる

---

## Phase 1: core修正 + Game Server

- [x] **Task 1-A: core修正 — MaxRooms制約チェック追加**

core/simulation/action.go の validateDigRoom に GameConstraints.MaxRooms チェックを追加する。
現在の部屋数が MaxRooms 以上なら DigRoom アクションを拒否する。

1. validateDigRoom に MaxRooms チェック追加（MaxRooms > 0 の場合のみ制約を適用）
2. エラーメッセージ: "max rooms reached: %d/%d"
3. SimpleAIPlayer が MaxRooms に達した場合に DigRoom を試行しないよう修正

**完了条件**:
- MaxRooms=5 のシナリオで 5部屋建設後に DigRoom が拒否されるテスト
- SimpleAIPlayer が MaxRooms 制約を尊重するテスト
- 既存テストが全パス

- [x] **Task 1-B: core修正 — SimpleAIPlayerコリドー戦略**

core/simulation/ai_player.go の SimpleAIPlayer に、新部屋建設後に隣接部屋への通路を掘るロジックを追加する。

1. SimpleAIPlayer の DecideActions メソッドで DigRoom 後の次ティックに DigCorridor を発行
2. 対象: 新部屋と既存の隣接部屋の間（龍穴からの気の伝播経路を確保）
3. 通路が掘れない場合（壁、距離等）はスキップ

**完了条件**:
- SimpleAIPlayer が部屋建設後に通路を掘ることを確認するテスト
- 通路経由で気が新部屋に伝播することを確認するテスト（ChiFlowEngine.Tick 後に新部屋の RoomChi.Current > 0）
- 既存テストが全パス

- [x] **Task 1-C: core修正 — caveScore正規化（呼び出し側）**

core/simulation/engine.go の Step メソッドで、`ev.CaveTotal()` の生値（部屋スコア合計）をそのまま CalcTickSupply に渡している問題を修正する。CalcTickSupply は caveScore を [0,1] の範囲として扱う設計のため、呼び出し側で正規化が必要。

1. simulation/engine.go で CaveTotal を正規化してから CalcTickSupply に渡す
2. 正規化方法: CaveTotal / MaxPossibleScore（全部屋の理論最大スコア合計）、0.0〜1.0 にクランプ
3. MaxPossibleScore が 0 の場合（部屋なし）は caveScore = 0.0

**完了条件**:
- CalcTickSupply に渡される caveScore が 0.0〜1.0 の範囲に収まるテスト
- 部屋数を増やしても供給量が指数的に膨らまないテスト（線形またはサブリニア）
- 既存テストの期待値を正規化後の値に更新し、全パス

- [x] **Task 1-D: core v1.1.0 タグ + DECISIONS.md更新**

1. Task 1-A〜1-C の修正が core の全テストをパスすることを確認
2. DECISIONS.md に以下を追記:
   - D014: MaxRooms制約のバリデーション追加（RESOLVED）
   - D015: SimpleAIPlayerコリドー戦略（RESOLVED）
   - D016: caveScore正規化（RESOLVED）
3. core に v1.1.0 タグを打つ準備（タグ打ち自体はチャットで確認後に実施）

**完了条件**:
- `cd core && go test ./... -count=1` 全パス
- `cd core && go vet ./...` クリーン
- DECISIONS.md に D014, D015, D016 が記載

- [x] **Task 1-E: Game Server — ActionProviderインターフェースとGameServer構造体**

sim/server/ パッケージを作成する。

1. server/provider.go: ActionProvider インターフェース定義
   ```go
   type ActionProvider interface {
       ProvideActions(snapshot *simulation.GameSnapshot) ([]simulation.PlayerAction, error)
       OnTickComplete(snapshot *simulation.GameSnapshot)
       OnGameEnd(result *simulation.RunResult)
   }
   ```
2. server/server.go: GameServer 構造体
   ```go
   type GameServer struct { ... }
   func NewGameServer(scenario *scenario.ScenarioDef, seed int64) (*GameServer, error)
   func (gs *GameServer) RunGame(provider ActionProvider) (*simulation.RunResult, error)
   ```
3. GameServer.RunGame は内部で core の SimulationRunner.Run を呼び、ActionProvider をブリッジする

**完了条件**:
- GameServer が NewGameServer → RunGame の流れでゲームを1ゲーム完走するテスト（core の SimpleAIPlayer をラップした ActionProvider で）
- ActionProvider の各メソッドが適切なタイミングで呼ばれることを検証するモックテスト

- [x] **Task 1-F: Game Server — セッション管理とシナリオ読み込み**

1. server/session.go: セッションライフサイクル管理
   - LoadScenario(path string) でシナリオJSON読み込み
   - 組み込みシナリオ（"tutorial", "standard"）のサポート
2. GameServer にセッション操作を委譲するメソッド追加

**完了条件**:
- シナリオJSONファイルの読み込み → ゲーム開始テスト
- 組み込みシナリオ名（"tutorial"）での読み込みテスト

- [x] **Task 1-G: Game Server — チェックポイント保存/復元**

1. server/checkpoint.go: チェックポイント機能
   - SaveCheckpoint(path string) — 現在のゲーム状態をファイルに保存
   - LoadCheckpoint(path string) — ファイルからゲーム状態を復元
2. GameServer にチェックポイント操作を委譲するメソッド追加

**完了条件**:
- チェックポイント保存 → 復元 → 続行が元実行と同一結果になるテスト

- [x] **Task 1-H: Game Server — リプレイ保存/再生**

1. server/replay.go: リプレイ機能
   - SaveReplay(path string) — アクション履歴をファイルに保存
   - LoadReplay(path string) — ファイルからアクション履歴を読み込み再生
2. GameServer にリプレイ操作を委譲するメソッド追加

**完了条件**:
- リプレイ保存 → 再生が元実行と同一結果になるテスト

- [x] **Task 1-I: Metrics Collector スケルトン**

sim/metrics/ パッケージのスケルトンを作成する。

1. metrics/collector.go: Collector 構造体と基本インターフェース
   ```go
   type Collector struct { ... }
   func NewCollector() *Collector
   func (c *Collector) OnTick(snapshot *simulation.GameSnapshot, actions []simulation.PlayerAction)
   func (c *Collector) OnGameEnd(result *simulation.RunResult) *GameSummary
   ```
2. metrics/summary.go: GameSummary 構造体（勝敗、総ティック数、建設部屋数、最終CoreHP等の基本統計）
3. GameServer に Collector を組み込み、毎ティック OnTick を呼ぶ

**完了条件**:
- Collector が GameServer 経由でティックごとに呼ばれるテスト
- GameSummary に基本統計が正しく集計されるテスト（1ゲーム完走後）
- `go vet ./...` クリーン

- [x] **Task 1-J: Phase 1 統合テストと完了**

1. GameServer + SimpleAIPlayer(ActionProviderラップ) + Collector で1ゲーム完走の統合テスト
2. チュートリアルシナリオと標準シナリオの両方で完走確認
3. GameSummary の出力を確認（勝敗、ティック数等が妥当な値か）
4. DECISIONS.md 棚卸し — Phase 1 で生まれた設計判断があれば追記
5. LESSONS.md — Phase 1 の知見追記

**完了条件**:
- 統合テスト2件（tutorial, standard）がパス
- `cd sim && go test ./... -count=1 -race` 全パス
- `cd sim && go vet ./...` クリーン
- PHASE_COMPLETE_1.md 生成

---

## Phase 2: Human Mode（対話メニュー）

- [x] **Task 2-A: ASCII描画の拡張**

sim/render/ パッケージを作成する。

1. render/ascii.go: core の caveviz/RenderFullStatus をベースに拡張
   - ANSIカラーコード対応（属性ごとの色: 火=赤, 水=青, 木=緑, 金=黄, 土=茶）
   - 仙獣の状態表示（名前、Lv、行動状態）
   - 侵入者の位置・状態表示
   - 経済情報（ChiPool残量、供給量/ティック、維持コスト/ティック）
   - CoreHP バー表示
2. render/format.go: 共通フォーマット関数（HP バー、プログレスバー、カラー出力）

**完了条件**:
- テスト用の GameSnapshot を渡して期待通りの ASCII 文字列が生成されるテスト
- ANSIカラーが正しく出力されるテスト（エスケープシーケンスの検証）
- ターミナル幅80文字以内に収まるレイアウト

- [x] **Task 2-B: Human Mode — メインメニューとアクション選択**

sim/adapter/human/ パッケージを作成する。

1. adapter/human/menu.go: メインメニュー表示と入力受付
   - 1. 部屋を掘る → サブメニューへ
   - 2. 通路を掘る → サブメニューへ
   - 3. 仙獣を召喚する → サブメニューへ
   - 4. 部屋をアップグレードする → サブメニューへ
   - 5. 何もしない（1ティック進める）
   - 6. 早送り（Nティック）→ ティック数入力
   - s. セーブ / l. ロード / r. リプレイ保存 / q. 終了
2. adapter/human/input.go: 入力ヘルパー（数字入力、Yes/No、座標入力）
3. 無効入力時のエラー表示と再入力ループ

**完了条件**:
- io.Reader からスクリプト化された入力を渡してメニュー遷移をテスト
- 無効入力（範囲外の数字、非数値文字列）でエラー表示後に再入力になるテスト

- [x] **Task 2-C: Human Mode — サブメニュー（部屋掘削・通路）**

1. adapter/human/submenu_build.go: 建設系サブメニュー
   - 部屋を掘る: 座標入力 → 属性選択（fire/water/wood/metal/earth） → PlayerAction生成
   - 通路を掘る: 始点部屋ID → 終点部屋ID → PlayerAction生成
2. valid_actions に基づく選択肢の動的生成
   - 建設不可セルは表示しない、コスト不足の場合は警告表示
3. "戻る" 操作でメインメニューに復帰

**完了条件**:
- 部屋掘削サブメニューからPlayerActionが正しく生成されるテスト
- 通路掘削サブメニューからPlayerActionが正しく生成されるテスト
- コスト不足時に警告が表示されるテスト
- "戻る" でメインメニューに戻るテスト

- [x] **Task 2-D: Human Mode — サブメニュー（召喚・アップグレード）**

1. adapter/human/submenu_unit.go: 仙獣系サブメニュー
   - 仙獣を召喚する: 部屋ID → 種族選択 → PlayerAction生成
   - 部屋をアップグレードする: 部屋ID → PlayerAction生成
2. valid_actions に基づく選択肢の動的生成
   - コスト不足の場合は警告表示
3. "戻る" 操作でメインメニューに復帰

**完了条件**:
- 召喚サブメニューからPlayerActionが正しく生成されるテスト
- アップグレードサブメニューからPlayerActionが正しく生成されるテスト
- コスト不足時に警告が表示されるテスト

- [x] **Task 2-E: Human Mode — ActionProvider実装とティック結果表示**

1. adapter/human/provider.go: ActionProvider インターフェース実装
   - ProvideActions: メニュー表示 → 入力受付 → PlayerAction返却
   - OnTickComplete: ティック結果のサマリー表示（戦闘発生、経済変化、イベント通知）
   - OnGameEnd: 勝敗結果とGameSummaryの表示
2. adapter/human/display.go: ティック結果の表示ロジック
   - 戦闘が発生した場合: 部屋ごとの戦闘結果（ダメージ、倒した/倒された）
   - 侵入波が到達した場合: 波の情報（敵数、種類）
   - CoreHPが減少した場合: 警告表示
   - 仙獣が成長/進化した場合: 通知
3. 早送り中は描画スキップ（完了後にサマリー表示）

**完了条件**:
- スクリプト入力で1ゲーム完走するE2Eテスト（wait連打でゲーム終了まで進む）
- 早送り（ff 50）で50ティック分スキップされるテスト
- OnTickComplete で戦闘ログが出力されるテスト

- [x] **Task 2-F: Human Mode — CLI統合とチェックポイント操作**

1. cmd/chaosseed-sim/main.go の `--human` フラグ実装
   - `--scenario` 引数でシナリオ指定
   - `--scenario tutorial` で組み込みチュートリアル
2. Human Mode のセーブ/ロード統合
   - 's' キー → ファイルパス入力 → チェックポイント保存
   - 'l' キー → ファイルパス入力 → チェックポイント復元 → ゲーム再開
   - 'r' キー → ファイルパス入力 → リプレイ保存
3. 'q' キーで確認ダイアログ → 終了

**完了条件**:
- `chaosseed-sim --human --scenario tutorial` でゲームが起動するテスト
- セーブ → 終了 → ロードでゲームが続行されるテスト
- リプレイ保存 → `chaosseed-sim --replay` で再生されるテスト

- [x] **Task 2-G: Phase 2 統合テストと完了**

1. チュートリアルシナリオのスクリプト実行E2Eテスト
   - 部屋掘る → 仙獣召喚 → wait × N → ゲーム終了 の一連の流れ
2. 全サブメニューの遷移テスト
3. ASCII描画の目視確認用スクリプト（手動確認用、テストではない）
4. DECISIONS.md 棚卸し
5. LESSONS.md 追記

**完了条件**:
- E2Eテストがパス
- `cd sim && go test ./... -count=1 -race` 全パス
- `cd sim && go vet ./...` クリーン
- PHASE_COMPLETE_2.md 生成

---

## Phase 3: AI Mode（JSON I/O）

- [x] **Task 3-A: AI Modeプロトコル定義**

sim/adapter/ai/ パッケージを作成する。

1. adapter/ai/protocol.go: メッセージ型定義
   ```go
   // サーバー → クライアント
   type StateMessage struct {
       Type         string                     `json:"type"`          // "state"
       Tick         int                        `json:"tick"`
       Snapshot     json.RawMessage            `json:"snapshot"`
       ValidActions []ValidAction              `json:"valid_actions"`
   }
   type GameEndMessage struct {
       Type    string                          `json:"type"`          // "game_end"
       Result  string                          `json:"result"`        // "victory" | "defeat"
       Summary json.RawMessage                 `json:"summary"`
       Metrics json.RawMessage                 `json:"metrics"`
   }
   type ErrorMessage struct {
       Type    string                          `json:"type"`          // "error"
       Message string                          `json:"message"`
   }

   // クライアント → サーバー
   type ActionMessage struct {
       Type    string                          `json:"type"`          // "action"
       Actions []ActionDef                     `json:"actions"`
   }

   type ValidAction struct {
       Kind   string                           `json:"kind"`          // "dig_room", "dig_corridor", ...
       Params map[string]interface{}            `json:"params"`        // アクション固有のパラメータ
   }
   ```
2. JSON Lines 形式: 1行1メッセージ、改行区切り

**完了条件**:
- 全メッセージ型の JSON シリアライズ/デシリアライズのラウンドトリップテスト
- ValidAction が正しく生成されるテスト（建設可能座標、召喚可能種族等）

- [x] **Task 3-B: AI Mode — valid_actions生成とSnapshot変換**

1. adapter/ai/serializer.go:
   - SnapshotToJSON: GameSnapshot を JSON に変換
   - BuildValidActions: 現在のゲーム状態から実行可能なアクション一覧を生成
     - dig_room: 建設可能な全座標と選択可能な属性
     - dig_corridor: 接続可能な部屋ペア
     - summon_beast: 配置可能な部屋と召喚可能な種族（コスト条件含む）
     - upgrade_room: アップグレード可能な部屋（コスト条件含む）
     - wait: 常に利用可能
2. コスト不足のアクションは valid_actions に含めない（LLM の無駄な試行を防ぐ）

**完了条件**:
- チュートリアルシナリオの初期状態から valid_actions が正しく生成されるテスト
- ChiPool 不足時に召喚が valid_actions に含まれないテスト
- MaxRooms 到達後に dig_room が valid_actions に含まれないテスト

- [x] **Task 3-C: AI Mode — ActionProvider実装とエラーハンドリング**

1. adapter/ai/provider.go: ActionProvider実装
   - ProvideActions: StateMessage を stdout に書き出し → stdin から ActionMessage を読み取り → パース → バリデーション → PlayerAction変換
   - OnTickComplete: 何もしない（AI Mode では StateMessage に全情報を含めるため）
   - OnGameEnd: GameEndMessage を stdout に書き出し
2. エラーハンドリング:
   - 不正なJSON → ErrorMessage を返却 → 再度 StateMessage を送信して再入力待ち
   - valid_actions にないアクション → ErrorMessage → 再入力待ち
   - stdin がEOF → ゲームを "quit" で終了
   - タイムアウト（`--timeout` フラグ）→ 自動 wait アクション

**完了条件**:
- パイプ経由で ActionMessage を送り、次の StateMessage を受け取るテスト
- 不正JSON送信 → ErrorMessage 受信 → 再送信で正常続行するテスト
- EOF → ゲーム終了のテスト

- [x] **Task 3-D: AI Mode — CLI統合とE2Eテスト**

1. cmd/chaosseed-sim/main.go の `--ai` フラグ実装
   - `--scenario` でシナリオ指定
   - `--timeout` でアクション入力タイムアウト指定（デフォルト: なし）
2. E2Eテスト: Go テストから chaosseed-sim プロセスをパイプで起動し、JSON Lines で対話して1ゲーム完走
   - 全ティック wait で完走（最低限の動作確認）
   - 数回の dig_room + summon_beast + wait で完走（基本戦略）
3. adapter/ai/docs/PROTOCOL.md: AI Mode プロトコル仕様書
   - LLMのプロンプトにそのまま含められる形式
   - メッセージ型、valid_actions の使い方、エラー処理

**完了条件**:
- `echo '...' | chaosseed-sim --ai --scenario tutorial` で動作するテスト
- E2Eテスト2件（wait完走、基本戦略完走）がパス
- PROTOCOL.md が存在し、プロトコルの全メッセージ型をカバー

- [x] **Task 3-E: Phase 3 統合テストと完了**

1. AI Mode の全機能の統合テスト
2. Human Mode との共存テスト（同一バイナリで `--human` と `--ai` が切り替え可能）
3. DECISIONS.md 棚卸し
4. LESSONS.md 追記

**完了条件**:
- `cd sim && go test ./... -count=1 -race` 全パス
- `cd sim && go vet ./...` クリーン
- PHASE_COMPLETE_3.md 生成

---

## Phase 4: Batch Mode + 壊れるサイン検出

- [x] **Task 4-A: Metrics — 壊れるサイン検出メトリクス B01〜B05**

metrics/collector.go に以下のメトリクス収集ロジックを実装する。

| ID | メトリクス | 収集タイミング |
|---|---|---|
| B01 | TicksBeforeFirstWave — 初波到達ティック | 初回侵入波到達時に記録 |
| B02 | ActionsBeforeFirstWave — 初波前のPlayerAction数 | 初回侵入波到達時に記録 |
| B03 | TerrainBlockRate — 地形制約による建設阻止率 | DigRoom失敗時に加算 |
| B04 | ZeroBuildableRate — 建設可能セル極小ゲームの割合 | ゲーム開始時に判定 |
| B05 | WaveOverlapRate — 建設中に侵入波が到達した割合 | 侵入波到達時に直前Nティックの建設有無を確認 |

**完了条件**:
- 各メトリクスが正しく収集されるユニットテスト（5件）
- B01: 初波がtick 30で来るシナリオ → B01=30
- B02: 初波前にDigRoom 3回 → B02=3
- B03: 10回のDigRoom試行で3回地形阻止 → B03=0.3
- B05: 侵入波到達の直前5ティック以内にDigRoomあり → overlap記録

- [x] **Task 4-B: Metrics — 壊れるサイン検出メトリクス B06〜B09**

| ID | メトリクス | 収集タイミング |
|---|---|---|
| B06 | StompRate — CoreHP80%以上残して勝利の割合 | ゲーム終了時に判定 |
| B07 | EarlyWipeRate — 全ティック50%以内に敗北の割合 | ゲーム終了時に判定 |
| B08 | PerfectionRate — 全部屋MaxLv到達ゲームの割合 | ゲーム終了時（勝利時のみ）に判定 |
| B09 | AvgRoomLevelRatio — クリア時のLv平均/MaxLv | ゲーム終了時（勝利時のみ）に計算 |

**完了条件**:
- 各メトリクスが正しく収集されるユニットテスト（4件）
- B06: CoreHP 90/100 で勝利 → stomp=true
- B07: max_ticks=200 のシナリオで tick 80 に敗北 → early_wipe=true
- B08: 全部屋Lv5/MaxLv5 で勝利 → perfection=true
- B09: 3部屋でLv 2,3,4、MaxLv=5 → ratio=0.6

- [x] **Task 4-C: Metrics — 壊れるサイン検出メトリクス B10〜B11**

| ID | メトリクス | 収集タイミング |
|---|---|---|
| B10 | LayoutEntropy — 異なるseed間の部屋配置エントロピー | バッチ完了後に全ゲームの配置を比較 |
| B11 | ResourceSurplusRate — ゲーム後半でChiPool余剰のティック割合 | 毎ティック ChiPool/MaxObserved を追跡 |

**完了条件**:
- B10: 異なるseed 10個でのゲーム結果から配置エントロピーが計算されるテスト
- B11: ChiPool が常に余裕あるゲーム → surplus_rate > 0.5 のテスト

- [x] **Task 4-D: Metrics — BreakageReport（閾値判定・アラート生成）**

1. metrics/breakage.go:
   - BreakageAlert 構造体（MetricID, BrokenSign, Value, Threshold, Direction）
   - BreakageReport 構造体（Alerts, Clean）
   - DetectBreakage() — 全メトリクスの閾値判定
2. 閾値はD002テーブルから直接転記:
   - B01: < シナリオ定義の最低猶予ティック
   - B02: < 3
   - B03: < 0.05
   - B04: > 0.10
   - B05: < 0.30
   - B06: > 0.30
   - B07: > 0.20
   - B08: > 0.05
   - B09: > 0.80
   - B10: < 閾値（要シナリオ依存の調整）
   - B11: > 0.50

**完了条件**:
- DetectBreakage() が閾値違反のメトリクスだけをアラートに含めるテスト
- アラート0件のケースで Clean に全IDが含まれるテスト

- [x] **Task 4-E: Batch Mode — 並列バッチ実行**

sim/adapter/batch/ パッケージを作成する。

1. adapter/batch/runner.go:
   - BatchRunner 構造体（シナリオ、ゲーム数、AIタイプ、並列数）
   - Run() → []GameSummary + BreakageReport
   - goroutine プール（runtime.NumCPU() ベース）
   - 進捗表示（stderr に "100/1000 games completed..." 形式）
2. 各ゲームは独立したRNGシード（ベースシード + ゲーム番号）

**完了条件**:
- 100ゲームのバッチ実行が成功するテスト
- 並列実行でも決定論性が保たれるテスト（同一ベースシード → 同一結果）
- GameSummary が全ゲーム分集約されるテスト

- [x] **Task 4-F: Batch Mode — レポート生成とパラメータスイープ**

1. metrics/report.go:
   - GenerateJSON(summaries, breakageReport) → JSON文字列
   - GenerateCSV(summaries) → CSV文字列
   - 出力フォーマットはPRDセクション3.3のBatch Mode出力に準拠
2. adapter/batch/sweep.go:
   - パラメータスイープ実行（"key=v1,v2,v3" 形式のパース）
   - シナリオJSONの指定パラメータを動的に書き換えてバッチ実行
   - 各パラメータ値ごとの BreakageReport を比較出力

**完了条件**:
- JSON レポートが PRD のフォーマットに準拠するテスト
- CSV レポートが生成されるテスト
- パラメータスイープ "economy.supply_multiplier=0.5,1.0,2.0" で3回のバッチ実行が行われるテスト

- [x] **Task 4-G: Batch Mode — CLI統合とD002検証**

1. cmd/chaosseed-sim/main.go の `--batch` フラグ実装
   - `--scenario`, `--games`, `--ai`, `--output`, `--format`, `--sweep` オプション
2. D002検証: core Phase 7-H の検証をBreakageReportで再現
   - チュートリアルシナリオ × SimpleAI × 1,000ゲーム
   - BreakageReport のアラートが0件であることを確認
3. パフォーマンステスト: 1,000ゲーム × SimpleAI が5分以内

**完了条件**:
- `chaosseed-sim --batch --scenario tutorial --games 1000 --ai simple --output results.json` が成功
- results.json の breakage_report.alerts が空配列
- 1,000ゲームが5分以内に完了（テストでタイムアウト設定）

- [ ] **Task 4-H: Phase 4 統合テストと完了**

1. Batch Mode の全機能統合テスト
2. Human Mode / AI Mode / Batch Mode の共存テスト
3. DECISIONS.md 棚卸し
4. LESSONS.md 追記

**完了条件**:
- `cd sim && go test ./... -count=1 -race` 全パス
- `cd sim && go vet ./...` クリーン
- PHASE_COMPLETE_4.md 生成

---

## Phase 5: バランス調整ダッシュボード

- [ ] **Task 5-A: ダッシュボード — ベースライン実行と壊れるサイン表示**

sim/balance/ パッケージを作成する。

1. balance/dashboard.go:
   - Dashboard 構造体（シナリオ、ゲーム数、AIタイプ）
   - Run() — ベースラインのバッチ実行 → BreakageReport の整形表示
   - 表示フォーマット: ✅/🔴 + メトリクスID + 名前 + 値 + 閾値
2. stdin からの対話操作（スイープ実行の確認等）

**完了条件**:
- ダッシュボード表示が PRD セクション4 のフォーマットに準拠するテスト
- アラート0件の場合に "No breakage detected" が表示されるテスト
- アラートありの場合にメトリクスID + 壊れるサイン + 調整の方向が表示されるテスト

- [ ] **Task 5-B: ダッシュボード — スイープ提案と比較**

1. balance/suggest.go:
   - SuggestSweep(alert BreakageAlert) → パラメータ名と値の候補
   - D002テーブルの「調整の方向」に基づくルールベースの提案
   - 例: B06(StompRate) → invasion.base_attack_power のスイープ提案
2. balance/compare.go:
   - CompareResults(baseline, sweepResults) → 比較表の文字列
   - 新たなアラートが発生したパラメータ値には警告マーク
   - "Best" 判定: アラート解消 + 新規アラートなし

**完了条件**:
- B06 アラートから invasion.base_attack_power のスイープが提案されるテスト
- スイープ結果の比較表が生成されるテスト
- 新規アラート発生時に警告が表示されるテスト

- [ ] **Task 5-C: ダッシュボード — パラメータ適用とCLI統合**

1. balance/apply.go:
   - ApplyParameter(scenarioPath, key, value) — シナリオJSONの指定パラメータを書き換え
   - バックアップファイルの自動生成（.bak）
2. cmd/chaosseed-sim/main.go の `--balance` フラグ実装
   - `--scenario`, `--games` オプション
3. E2Eテスト: ダッシュボード起動 → スイープ → 適用 の一連の流れ

**完了条件**:
- `chaosseed-sim --balance --scenario tutorial --games 100` が起動するテスト
- パラメータ適用後にシナリオJSONが更新されるテスト
- バックアップファイルが生成されるテスト

- [ ] **Task 5-D: Phase 5 統合テストと完了**

1. ダッシュボードの全機能統合テスト
2. 全モード（Human / AI / Batch / Balance）の共存テスト
3. DECISIONS.md 棚卸し
4. LESSONS.md 追記

**完了条件**:
- `cd sim && go test ./... -count=1 -race` 全パス
- `cd sim && go vet ./...` クリーン
- PHASE_COMPLETE_5.md 生成

---

## Phase 6: 統合検証 + リリース

- [ ] **Task 6-A: 全モード統合テスト**

1. Human Mode: スクリプト入力でチュートリアルシナリオ完走
2. AI Mode: パイプ経由でチュートリアルシナリオ完走
3. Batch Mode: 1,000ゲーム × SimpleAI → BreakageReport アラート0件
4. Balance: ダッシュボード起動 → BreakageReport 表示
5. リプレイ: Batch で記録 → `--replay` で再生 → 決定論性確認
6. チェックポイント: Human Mode で保存 → `--checkpoint` で復元 → 続行
7. `go test ./... -count=1 -race` 全パス（core + sim）
8. `go vet ./...` クリーン（core + sim）

**完了条件**:
- 上記8項目すべてパス

- [ ] **Task 6-B: ドキュメント更新とリリース準備**

1. HANDOFF.md 作成（現在未作成のため新規作成）
   - sim の完了フェーズを追記
   - アーキテクチャ（Game Server + 3アダプター）の記載
   - Breakageメトリクス（B01〜B11）の記載
   - 次のステップ（chaosseed-game）の更新
2. DECISIONS.md 更新 — sim全フェーズで生まれた設計判断の最終棚卸し
3. LESSONS.md 更新 — sim全フェーズの知見の最終追記
4. sim/README.md の最終更新（全モードの使い方、プロトコル仕様へのリンク）
5. v1.0.0 タグの準備

**完了条件**:
- HANDOFF.md, DECISIONS.md, LESSONS.md が最新
- README.md が全モードをカバー
- PHASE_COMPLETE_6.md 生成
