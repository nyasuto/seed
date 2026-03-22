# chaosseed-game tasks.md

> Ralph Loop 用タスクファイル。1イテレーション1タスク。
> 上から順に消化する。各タスクの完了条件をすべて満たしてから次に進む。
> タスク完了時にチェックボックスを [x] にする。
>
> **GUI固有のルール**: 描画の「見た目の正しさ」はgo testで検証できない。
> タスクの完了条件は「ロジックテスト + コンパイル成功 + vetクリーン」に限定する。
> 目視確認はフェーズ末尾の統合タスク（Ralph Loopの外、チャットでぽんぽこが実施）で行う。

---

## Phase 0: Ebitengine習作 + プロジェクト初期化

- [ ] **Task 0-A: プロジェクト初期化とEbitengine導入**

1. game/ ディレクトリで作業する
2. Ebitengine v2 を `go get` で導入
3. game/main.go: 最小の Ebitengine アプリ（空ウィンドウ表示、1088x728、タイトル "ChaosForge"）
4. game/Makefile 作成（build, test, vet, clean）
5. ルート Makefile に game を追加
6. CLAUDE.md 更新（game の依存ポリシー追記）

**完了条件**:
- `go work sync` が成功
- `cd game && go build ./...` が成功
- `cd game && go vet ./...` がクリーン
- バイナリ実行でウィンドウが表示される（手動確認OK、ループ停止条件ではない）

- [ ] **Task 0-B: カラーパレットと仮タイルセット生成**

1. asset/palette.go: ゲーム全体のカラーパレット定義
   - 属性色（Fire=赤系, Water=青系, Wood=緑系, Metal=黄系, Earth=茶系）
   - 地形色（Wall=濃い灰, Floor=薄い灰, Corridor=灰, DragonHole=紫, HardRock=非常に濃い灰, WaterTerrain=深い青）
   - UI色（背景、テキスト、ボーダー）
2. asset/tileset.go: TilesetProvider インターフェース定義
   ```go
   type TilesetProvider interface {
       GetTile(cellType world.CellType, element types.Element) *ebiten.Image
       GetBeastSprite(species string, level int) *ebiten.Image
       GetInvaderSprite(class string) *ebiten.Image
   }
   ```
3. asset/placeholder.go: PlaceholderProvider 実装
   - 全CellType分の32x32色付き矩形を生成
   - HardRock: ✕模様、WaterTerrain: ～模様、DragonHole: ★マーク

**完了条件**:
- PlaceholderProvider が TilesetProvider を満たすコンパイル確認
- 全CellTypeに対して GetTile が non-nil の *ebiten.Image を返すテスト
- 全タイルが32x32ピクセルであるテスト

- [ ] **Task 0-C: Caveデータからタイルマップ描画**

1. view/mapview.go: MapView 構造体
   - Caveデータからタイルマップへの座標変換: `CellPos(x,y) → ScreenPos(px,py)`
   - `Draw(screen *ebiten.Image, cave *world.Cave, provider TilesetProvider)`
   - 全セルを走査し、CellType に応じたタイルを描画
2. coreのCaveデータを読み込み、MapViewで描画するデモ
   - 組み込みチュートリアルシナリオのCaveをそのまま使用

**完了条件**:
- CellPos→ScreenPos の座標変換テスト（境界値含む）
- ScreenPos→CellPos の逆変換テスト（マウスクリック→セル座標の基礎）
- `go build ./...` が成功

- [ ] **Task 0-D: マウスホバーとセル情報表示**

1. input/mouse.go: マウス座標→セル座標変換
   - ScreenPos→CellPos 変換（MapView の逆変換を利用）
   - ホバー中のセル座標を毎フレーム追跡
2. view/tooltip.go: ツールチップ描画
   - ホバー中のセルの CellType を文字表示
   - 部屋セルの場合は部屋ID、属性、レベルを表示
3. main.go に MapView + ツールチップを統合

**完了条件**:
- マウス座標(320, 160)→セル座標(10, 5)のような変換テスト
- マップ外の座標が無効として扱われるテスト
- `go build ./...` が成功

- [ ] **Task 0-E: Phase 0 統合と確認**

1. チュートリアルシナリオのCaveサイズが24x20タイルに収まることを確認
   - 収まらない場合は core のシナリオJSONを調整（DECISIONS.md に記録）
2. `go test ./... -count=1 -race` 全パス
3. `go vet ./...` クリーン
4. DECISIONS.md 棚卸し
5. LESSONS.md — Phase 0 の知見追記（Ebitengine の学び）

**完了条件**:
- `cd game && go test ./... -count=1 -race` 全パス
- `cd game && go vet ./...` クリーン
- PHASE_COMPLETE_0.md 生成

---

## Phase 1: Game Controller + ティック進行

- [ ] **Task 1-A: GameController — core接続とSnapshot管理**

1. controller/controller.go: GameController 構造体
   ```go
   type GameController struct {
       engine   *simulation.SimulationEngine
       snapshot *simulation.GameSnapshot
       pending  []simulation.PlayerAction
       state    GameState  // Playing, Paused, FastForward, GameOver
   }
   func NewGameController(scenarioDef *scenario.ScenarioDef, seed int64) (*GameController, error)
   func (gc *GameController) Snapshot() *simulation.GameSnapshot
   func (gc *GameController) AddAction(action simulation.PlayerAction)
   func (gc *GameController) AdvanceTick() (*simulation.StepResult, error)
   ```
2. シナリオ読み込み（組み込みシナリオ対応、sim/server の embed パターンを流用）
3. AdvanceTick: pending キューを Step() に渡し、Snapshot を更新

**完了条件**:
- NewGameController でチュートリアルシナリオ読み込み→AdvanceTick 1回→Snapshot が更新されるテスト
- AddAction → AdvanceTick で PlayerAction が engine に渡されるテスト
- 既存の core テストが影響を受けないことの確認

- [ ] **Task 1-B: ティック進行モード（手動/早送り/一時停止）**

1. controller/tick.go: ティック進行ロジック
   - Manual: AdvanceTick() の明示的呼び出しでのみ進行
   - FastForward: 毎フレーム N ティック自動進行（N は設定可能）
   - Paused: 進行停止（描画は継続）
2. TogglePause(), StartFastForward(speed int), StopFastForward()
3. FastForward 中に侵入波到達 or ゲーム終了で自動停止

**完了条件**:
- Manual モードで AdvanceTick 呼び出しなしではティック数が変わらないテスト
- FastForward で 10 ティック自動進行するテスト
- Paused → Resume でティック進行が再開するテスト
- FastForward 中にゲーム終了条件を満たすと停止するテスト

- [ ] **Task 1-C: GameSnapshot から描画データへの変換**

1. view/mapview.go の拡張: Snapshot から部屋の属性情報を読み取り、タイル色を属性に応じて変更
   - 部屋セルは部屋の Element に基づく色で描画
   - 龍穴は紫で描画
   - 空いている壁セルはデフォルト灰色
2. view/entity.go: 仙獣/侵入者のスプライト描画
   - Snapshot の仙獣リストから位置を読み取り、対応する部屋の中心に描画
   - Snapshot の侵入者リストから位置を読み取り、同様に描画
   - PlaceholderProvider 経由でスプライト取得

**完了条件**:
- Snapshot の部屋属性が Fire の場合、対応するセルが Fire 色タイルで描画される座標計算テスト
- 仙獣の部屋ID→描画座標の変換テスト
- 侵入者の部屋ID→描画座標の変換テスト

- [ ] **Task 1-D: Top Bar — ステータス表示**

1. view/ui.go: UIPanel 構造体
   - Top Bar: ChiPool（現在値/最大値）、CoreHP（バー + 数値）、Tick数
   - Snapshot から値を読み取り、毎フレーム描画
2. view/text.go: テキスト描画ヘルパー
   - Ebitengine のテキスト描画API（ebitenutil.DebugPrint or text パッケージ）
   - 位置指定、色指定のラッパー

**完了条件**:
- Snapshot の ChiPool=150, MaxChiPool=500 → "ChiPool: 150/500" のテキスト生成テスト
- CoreHP バーの幅計算テスト（80/100 → 80%幅）
- `go build ./...` が成功

- [ ] **Task 1-E: Phase 1 統合と確認**

1. main.go を統合: GameController + MapView + EntityView + TopBar
   - 起動 → チュートリアルシナリオ読み込み → マップ描画 → スペースキーでティック進行
   - 仙獣/侵入者がマップ上に表示される
2. `go test ./... -count=1 -race` 全パス
3. `go vet ./...` クリーン
4. DECISIONS.md 棚卸し
5. LESSONS.md 追記

**完了条件**:
- `cd game && go test ./... -count=1 -race` 全パス
- `cd game && go vet ./...` クリーン
- PHASE_COMPLETE_1.md 生成

---

## Phase 2: 操作システム

- [ ] **Task 2-A: 操作モードステートマシン**

1. input/action.go: ActionMode 定義
   ```go
   type ActionMode int
   const (
       ModeNormal ActionMode = iota
       ModeDigRoom
       ModeDigCorridor
       ModeSummon
       ModeUpgrade
   )
   ```
2. input/state.go: InputStateMachine
   - 現在のモード管理
   - モード切替（キーボード D/C/S/U、Escape でキャンセル→Normal に戻る）
   - モードに応じたマウスクリックの解釈を変更
3. モード表示（画面下部のAction Bar でアクティブモードをハイライト）

**完了条件**:
- キー 'D' で ModeDigRoom に遷移するテスト
- Escape で ModeNormal に戻るテスト
- ModeDigRoom 中にキー 'S' で ModeSummon に切り替わるテスト

- [ ] **Task 2-B: Action Bar 描画**

1. view/actionbar.go: Action Bar 描画
   - 5つのアクションボタン: [掘る D] [通路 C] [召喚 S] [強化 U] [待機 W]
   - 3つのティック制御ボタン: [▶ Space] [▶▶ F] [⏸ Esc]
   - アクティブモードのボタンをハイライト表示
   - マウスクリックでボタン押下判定
2. view/button.go: 汎用ボタン描画
   - 矩形 + テキストラベル + ホバー/アクティブ状態

**完了条件**:
- ボタン矩形内のクリック座標が「ボタン押下」と判定されるテスト
- ボタン矩形外のクリックが無視されるテスト
- `go build ./...` が成功

- [ ] **Task 2-C: DigRoom フロー**

1. ModeDigRoom の完全な操作フロー:
   - モード進入 → マップ上の壁セルをクリック → 部屋位置決定
   - 属性選択パネル表示（5属性のボタン一覧）→ クリックで選択
   - PlayerAction(DigRoom) を GameController.AddAction() に投入
   - ModeNormal に復帰
2. バリデーション:
   - 壁セル以外をクリック → エラー表示（赤テキスト）、モード継続
   - HardRock/Water セルをクリック → エラー表示
   - ChiPool 不足 → 属性選択パネルで警告表示

**完了条件**:
- 壁セル座標+Fire属性 → DigRoom PlayerAction が正しく生成されるテスト
- HardRock セルクリックが拒否されるテスト
- 属性選択後に ModeNormal に復帰するテスト

- [ ] **Task 2-D: DigCorridor フロー**

1. ModeDigCorridor の操作フロー:
   - モード進入 → 始点の部屋セルをクリック → 始点部屋ID記録 → ハイライト表示
   - 終点の部屋セルをクリック → DigCorridor PlayerAction 生成
   - ModeNormal に復帰
2. バリデーション:
   - 部屋セル以外をクリック → エラー表示
   - 始点と終点が同じ部屋 → エラー表示

**完了条件**:
- 始点部屋ID=1, 終点部屋ID=3 → DigCorridor PlayerAction が正しく生成されるテスト
- 部屋以外のセルクリックが拒否されるテスト
- 始点選択後にEscapeでキャンセル→ModeNormal復帰テスト

- [ ] **Task 2-E: SummonBeast フロー**

1. ModeSummon の操作フロー:
   - モード進入 → 部屋セルをクリック → 部屋ID記録
   - 種族選択パネル表示（利用可能な仙獣種族のボタン一覧）→ クリックで選択
   - SummonBeast PlayerAction 生成
   - ModeNormal に復帰
2. バリデーション:
   - 部屋以外をクリック → エラー表示
   - ChiPool 不足 → 種族選択パネルで警告
   - 部屋に既に仙獣がいる場合 → エラー表示

**完了条件**:
- 部屋ID=2, 種族="kirin" → SummonBeast PlayerAction が正しく生成されるテスト
- ChiPool 不足時に警告が返るバリデーションテスト

- [ ] **Task 2-F: UpgradeRoom フロー + Wait + ティック制御統合**

1. ModeUpgrade の操作フロー:
   - モード進入 → 部屋セルをクリック → UpgradeRoom PlayerAction 生成
   - ModeNormal に復帰
2. Wait: キー W → PlayerAction なしで AdvanceTick 呼び出し
3. ティック制御の統合:
   - スペースキー / [▶] ボタン → AdvanceTick（pending キュー実行）
   - キー F / [▶▶] ボタン → FastForward 開始
   - Escape / [⏸] ボタン → FastForward 停止 or モードキャンセル
4. pending キューに複数アクションが積まれている場合の表示（キュー内容を Action Bar 近くに表示）

**完了条件**:
- 部屋ID=1 → UpgradeRoom PlayerAction テスト
- Wait → ティック数が1増えるテスト
- FastForward 開始→10ティック進行→手動停止のテスト

- [ ] **Task 2-G: 操作フィードバック（エラー表示、モード表示）**

1. view/feedback.go: 操作フィードバック表示
   - エラーメッセージ: 赤テキストで画面下部に表示、3秒後にフェードアウト
   - モード表示: マウスカーソル近くにモード名テキスト（"掘削モード" 等）
   - 有効セルのハイライト: ModeDigRoom 中に壁セルを薄く光らせる
   - 無効セルの暗転: HardRock/Water を掘削モード中に暗く表示
2. view/selection.go: 選択パネルの汎用コンポーネント
   - 属性選択パネル（5つの色付きボタン）
   - 種族選択パネル（種族名のボタンリスト）
   - パネル表示中は他の操作をブロック

**完了条件**:
- エラーメッセージのタイマー（3秒）テスト
- 選択パネルの項目クリック判定テスト
- `go build ./...` が成功

- [ ] **Task 2-H: Phase 2 統合と確認**

1. 全操作フローの統合: main.go に ActionBar + InputStateMachine + 全フローを接続
2. キーボードショートカット一覧の最終確認: D/C/S/U/W/Space/F/Escape
3. `go test ./... -count=1 -race` 全パス
4. `go vet ./...` クリーン
5. DECISIONS.md 棚卸し
6. LESSONS.md 追記

**完了条件**:
- `cd game && go test ./... -count=1 -race` 全パス
- `cd game && go vet ./...` クリーン
- PHASE_COMPLETE_2.md 生成

---

## Phase 3: シーン管理 + UI

- [ ] **Task 3-A: シーンマネージャー**

1. scene/manager.go: SceneManager
   ```go
   type Scene interface {
       Update() error
       Draw(screen *ebiten.Image)
       OnEnter()
       OnExit()
   }
   type SceneManager struct { ... }
   func (sm *SceneManager) Switch(scene Scene)
   func (sm *SceneManager) Update() error
   func (sm *SceneManager) Draw(screen *ebiten.Image)
   ```
2. main.go の Game 構造体に SceneManager を組み込み、Update/Draw を委譲

**完了条件**:
- SceneA → Switch(SceneB) で OnExit(A) → OnEnter(B) が呼ばれるテスト
- Switch 後に Update/Draw が新シーンに委譲されるテスト

- [ ] **Task 3-B: タイトル画面とシナリオ選択画面**

1. scene/title.go: タイトル画面
   - ゲーム名 "ChaosForge — 風水回廊記" 表示
   - [新しいゲーム] ボタン → シナリオ選択画面へ
   - [ロード] ボタン → ファイル選択（Phase 4 で実装、ここではスタブ）
2. scene/select.go: シナリオ選択画面
   - 組み込みシナリオ一覧（tutorial, standard）
   - シナリオ選択 → InGame シーンへ遷移

**完了条件**:
- タイトル画面の [新しいゲーム] ボタンクリック → シナリオ選択画面への遷移テスト
- シナリオ選択 → InGame シーン生成テスト
- `go build ./...` が成功

- [ ] **Task 3-C: InGame シーンへの統合**

1. scene/ingame.go: InGameScene
   - Phase 1〜2 の全コンポーネント（GameController, MapView, EntityView, TopBar, ActionBar, InputStateMachine）を統合
   - Update() で入力処理→コントローラー更新→状態更新
   - Draw() で MapView→EntityView→UI→ツールチップ→フィードバックの順で描画
2. ゲーム終了検知: GameController が GameOver を返したら Result シーンへ遷移

**完了条件**:
- InGameScene の Update→Draw サイクルがエラーなく動くテスト（モック Snapshot で）
- ゲーム終了条件で Result シーンへの遷移が発生するテスト

- [ ] **Task 3-D: 結果画面**

1. scene/result.go: ResultScene
   - 勝利/敗北の大きなテキスト表示
   - 統計サマリー: 総ティック数、建設部屋数、撃退した侵入波数、最終CoreHP
   - [もう一度] ボタン → シナリオ選択画面へ
   - [タイトルへ] ボタン → タイトル画面へ

**完了条件**:
- 勝利時と敗北時で異なるテキストが表示されるテスト（文字列生成ロジック）
- [タイトルへ] ボタンでタイトル画面に遷移するテスト

- [ ] **Task 3-E: Info Panel — 部屋/仙獣/侵入者の詳細表示**

1. view/infopanel.go: InfoPanel
   - 部屋選択時: 部屋ID、属性、レベル、RoomChi（現在/最大）、配置仙獣リスト
   - 仙獣選択時: 種族、レベル、HP、ATK/DEF/SPD、行動状態（Guard/Patrol/Chase/Flee）
   - 侵入者選択時: クラス、HP、目標AI、現在の行動
   - 何も選択されていない場合: 全体のゲーム情報（シナリオ名、現在の侵入波状況）
2. 部屋クリック → Info Panel に詳細表示（ModeNormal 時）

**完了条件**:
- 部屋データから Info Panel 用のテキスト行が正しく生成されるテスト
- 仙獣データから Info Panel 用のテキスト行が正しく生成されるテスト

- [ ] **Task 3-F: Phase 3 統合と確認**

1. タイトル→シナリオ選択→InGame→Result の完全なシーンフローが動くことの確認
2. `go test ./... -count=1 -race` 全パス
3. `go vet ./...` クリーン
4. DECISIONS.md 棚卸し
5. LESSONS.md 追記

**完了条件**:
- `cd game && go test ./... -count=1 -race` 全パス
- `cd game && go vet ./...` クリーン
- PHASE_COMPLETE_3.md 生成

---

## Phase 4: セーブ/ロード + 仕上げ

- [ ] **Task 4-A: セーブ/ロード**

1. save/checkpoint.go: チェックポイント保存/復元
   - GameController の状態を JSON 保存（core の Checkpoint 機能を活用）
   - ファイルパス: `~/.chaosforge/saves/` ディレクトリ
   - ファイル名: `save_YYYYMMDD_HHMMSS.json`
2. save/config.go: ゲーム設定の保存/復元
   - ウィンドウサイズ、早送り速度
   - `~/.chaosforge/config.json`
3. InGame シーンにセーブ/ロードを統合
   - Ctrl+S → 自動ファイル名でセーブ → 確認メッセージ表示
   - Ctrl+L → 最新のセーブファイルからロード → ゲーム状態復元

**完了条件**:
- セーブ→ロードで GameController のティック数と Snapshot が一致するテスト
- セーブファイルが JSON として valid であるテスト
- 設定ファイルの保存/復元テスト

- [ ] **Task 4-B: タイトル画面のロード機能**

1. scene/title.go の [ロード] ボタン実装
   - `~/.chaosforge/saves/` 内のセーブファイル一覧を取得
   - 一覧表示（ファイル名、タイムスタンプ、シナリオ名）
   - 選択 → チェックポイント復元 → InGame シーンへ遷移
2. セーブファイルが存在しない場合: [ロード] ボタンをグレーアウト

**完了条件**:
- セーブファイル一覧の取得テスト（テスト用ディレクトリで）
- 存在しないディレクトリでもエラーにならないテスト
- `go build ./...` が成功

- [ ] **Task 4-C: 戦闘・侵入の視覚フィードバック**

1. 戦闘発生中の部屋: タイル枠線を赤く点滅（数フレーム周期で明滅）
2. 侵入波到達時: 画面上部に "Wave 3 incoming!" のような警告テキスト（数秒間表示）
3. CoreHP 減少時: CoreHP バーを赤く点滅
4. 仙獣敗北時: 対象仙獣のスプライトを点滅表示
5. ゲーム進行に影響する情報を視覚的に強調し、ASCIIの「気づけない」問題を解消

**完了条件**:
- 戦闘中フラグ ON の部屋が点滅対象として識別されるロジックテスト
- 侵入波到達の検知ロジックテスト（前ティックと現ティックの波数比較）
- `go build ./...` が成功

- [ ] **Task 4-D: Phase 4 統合と確認**

1. セーブ→終了→ロード→ゲーム続行のフロー確認
2. 戦闘/侵入フィードバックがゲーム進行中に機能することの確認
3. `go test ./... -count=1 -race` 全パス
4. `go vet ./...` クリーン
5. DECISIONS.md 棚卸し
6. LESSONS.md 追記

**完了条件**:
- `cd game && go test ./... -count=1 -race` 全パス
- `cd game && go vet ./...` クリーン
- PHASE_COMPLETE_4.md 生成

---

## Phase 5: 統合検証 + リリース

- [ ] **Task 5-A: チュートリアルシナリオ手動プレイテスト**

**このタスクはRalph Loopの外で、ぽんぽこがチャットで実施する。**

1. `chaosseed-game` を起動
2. タイトル → シナリオ選択（tutorial）→ ゲーム開始
3. 以下のD002体感チェックリストに沿ってプレイ:
   - [ ] 計画する: マップを見て「ここに部屋を掘りたい」と思えるか
   - [ ] 妥協する: 地形制約で理想が阻まれ、代替案を考える瞬間があるか
   - [ ] 中断される: 構築中に侵入波が来て「まだ準備できてない！」と焦るか
   - [ ] 噛み合う: 不完全な防衛で敵を撃退し「よし！」と感じるか
   - [ ] 完璧が来ない: クリア時に「次はもっとうまくやれる」と思えるか
4. 操作に関するフィードバック:
   - 全アクション（掘る/通路/召喚/強化/待機）が実行できるか
   - モード切替が直感的か
   - ティック進行が快適か
5. フィードバックをチャットに記録 → 修正が必要なら追加タスクを生成

**完了条件**:
- ぽんぽこがチュートリアルシナリオを完走（勝利 or 敗北）
- D002チェックリストの記入
- フィードバック記録

- [ ] **Task 5-B: 標準シナリオ手動プレイテスト + パフォーマンス確認**

**このタスクもRalph Loopの外。**

1. 標準シナリオの手動プレイ
2. パフォーマンス確認: 60FPS 維持（M4 Mac mini）
   - Ebitengine の FPS カウンターで確認
   - 仙獣・侵入者が多い状態（ゲーム中盤〜後半）でのフレームレート
3. 致命的なバグ・操作不能があれば追加タスク生成

**完了条件**:
- 標準シナリオを完走（勝利 or 敗北）
- 60FPS維持を確認
- フィードバック記録

- [ ] **Task 5-C: ドキュメント更新とリリース準備**

1. HANDOFF.md 更新
   - game の完了フェーズを追記
   - アーキテクチャ（GameController + SceneManager + View/Input 分離）の記載
   - 仮アセットシステム（TilesetProvider）の記載
   - D002体感フィードバックの結果を記載
   - 次のステップ（ドット絵アセット、BGM/SE、龍脈可視化等）の更新
2. DECISIONS.md 更新 — game全フェーズで生まれた設計判断の最終棚卸し
3. LESSONS.md 更新 — game全フェーズの知見の最終追記
4. game/README.md 作成（ビルド方法、操作方法、キーボードショートカット一覧）
5. v1.0.0 タグの準備

**完了条件**:
- HANDOFF.md, DECISIONS.md, LESSONS.md が最新
- game/README.md が存在し、操作方法をカバー
- `go test ./... -count=1 -race` 全パス（core + sim + game）
- `go vet ./...` クリーン（core + sim + game）
- PHASE_COMPLETE_5.md 生成