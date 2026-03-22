# chaosseed-game PRD v1.0.0 — Ebitengine GUIクライアント

> chaosseed-core のコアメカニクスエンジンを活用し、
> D002の面白さを体感できる最小のGUIを構築する。

---

## 1. プロジェクトの目的

### 1.1 なぜ game が必要か

simのASCIIプレイで判明した教訓：

**ASCIIでは、D002「計画する面白さ」が発揮される前に認知負荷で潰れる。**

マップの空間関係、気の流れ、仙獣と侵入者の位置関係を文字列から読み取るコストが高すぎ、
「地形を見て理想の構想を練る」というD002の第一歩が機能しない。

GUIが解決すること：

- タイルマップで部屋・通路・龍脈の空間関係が一目で分かる
- 仙獣と侵入者の位置・状態がスプライトで直感的に把握できる
- マウスクリックで部屋を掘る場所を指定でき、操作のフリクションが消える

### 1.2 成功基準

1. チュートリアルシナリオを最初から最後までGUIでプレイでき、D002の「計画→妥協→中断→噛み合う→完璧が来ない」の体験サイクルを体感できる
2. マップを見た瞬間に「ここに部屋を掘りたい」「ここは地形制約で掘れない」が分かる
3. 侵入波到達時に「どの部屋が危険か」がマップ上で即座に判別できる
4. 全PlayerAction（DigRoom, DigCorridor, SummonBeast, UpgradeRoom, Wait）がGUIから実行できる

### 1.3 v1.0.0 スコープ

**含める:**

- 32x32 タイルマップ描画（仮アセット: 色付き矩形+文字ラベル）
- 仙獣/侵入者のスプライト表示（仮アセット）
- マウス/キーボード操作で全アクション実行
- ティック進行制御（1ティックステップ、早送り、一時停止）
- 最小限のUI（ChiPool、CoreHP、ティック数の数値表示）
- シナリオ選択（組み込みシナリオ）
- セーブ/ロード（チェックポイント）

**含めない（v1.1.0以降）:**

- ドット絵アセット（v1.0.0は仮アセットで体感検証に集中）
- BGM・SE
- リアルタイム風水スコア可視化（龍脈の気の流れアニメーション等）
- 経済グラフ/履歴表示
- リプレイ再生UI
- マップスクロール/ズーム（v1.0.0はCaveサイズを画面に収まる範囲に制限）

---

## 2. アーキテクチャ

### 2.1 レイヤー構成

```
┌──────────────────────────────────────────────┐
│  Ebitengine Game Loop                        │
│  Update() → Draw() @ 60FPS                  │
├──────────────────────────────────────────────┤
│  Scene Manager                               │
│  Title → ScenarioSelect → InGame → Result   │
├──────────────────────────────────────────────┤
│  InGame Scene                                │
│  ┌────────────┐ ┌──────────┐ ┌────────────┐ │
│  │ Map View   │ │ UI Panel │ │ Input      │ │
│  │ タイル描画  │ │ ステータス│ │ マウス/キー │ │
│  └────┬───────┘ └────┬─────┘ └────┬───────┘ │
│       │              │            │          │
├───────┴──────────────┴────────────┴──────────┤
│  Game Controller                             │
│  - ティック進行管理（手動/自動/早送り）         │
│  - PlayerAction の組み立て                    │
│  - GameSnapshot → 描画データ変換              │
├──────────────────────────────────────────────┤
│  chaosseed-core                              │
│  SimulationEngine.Step(actions)              │
│  GameSnapshot（読み取り専用）                  │
└──────────────────────────────────────────────┘
```

### 2.2 core との接続

HANDOFF.mdに記載の通り、coreのSimulationEngineはGUIからも同じインターフェースで利用可能。

```go
// game/controller/controller.go
type GameController struct {
    engine   *simulation.SimulationEngine
    snapshot *simulation.GameSnapshot
    pending  []simulation.PlayerAction  // 今ティックの操作キュー
}

// Update はEbitengineのフレームごとに呼ばれる
func (gc *GameController) Update() {
    if gc.shouldAdvanceTick() {
        result := gc.engine.Step(gc.pending)
        gc.snapshot = gc.engine.BuildSnapshot()
        gc.pending = nil
        // result から勝敗判定
    }
}
```

**重要な設計判断: フレームレートとティックレートの分離**

- Ebitengineは60FPS（16.7ms/frame）で Update()/Draw() を呼ぶ
- coreのティックはプレイヤーの操作に応じて進行する（ターン制）
- 1フレーム=1ティック **ではない**
- プレイヤーがアクションを確定するか「次のティック」を押すまで、同じSnapshotを描画し続ける
- 早送りモードでは1フレームに複数ティックを処理する

### 2.3 パッケージ構成

```
game/
├── go.mod                    # module github.com/ponpoko/chaosseed/game
├── main.go                   # Ebitengine起動
├── controller/               # ゲーム制御（ティック管理、アクション組み立て）
├── scene/
│   ├── manager.go            # シーン遷移管理
│   ├── title.go              # タイトル画面
│   ├── select.go             # シナリオ選択画面
│   ├── ingame.go             # メインゲーム画面
│   └── result.go             # 結果画面（勝利/敗北）
├── view/
│   ├── mapview.go            # タイルマップ描画
│   ├── entity.go             # 仙獣/侵入者スプライト描画
│   ├── ui.go                 # UIパネル（ステータス、アクションバー）
│   └── tooltip.go            # ツールチップ（ホバー情報）
├── input/
│   ├── mouse.go              # マウス入力処理
│   ├── keyboard.go           # キーボード入力処理
│   └── action.go             # 入力→PlayerAction変換
├── asset/
│   ├── placeholder.go        # 仮アセット生成（色付き矩形）
│   ├── tileset.go            # タイルセット管理（将来のドット絵差し替え対応）
│   └── palette.go            # カラーパレット定義
├── save/
│   ├── checkpoint.go         # セーブ/ロード
│   └── config.go             # ゲーム設定（ウィンドウサイズ等）
└── docs/
    └── PRD.md
```

---

## 3. 画面設計

### 3.1 InGame 画面レイアウト

```
┌─────────────────────────────────────────────────────────────────┐
│  ChiPool: 150/500  │  CoreHP: ████████░░ 80/100  │  Tick: 42  │  ← Top Bar
├─────────────────────┬───────────────────────────────────────────┤
│                     │                                           │
│                     │                                           │
│     Map View        │        Info Panel                         │
│   (タイルマップ)     │  ┌─────────────────────────┐              │
│                     │  │ 選択中: 部屋 #3 (Fire)  │              │
│   32x32 タイル      │  │ Level: 2/5              │              │
│   で洞窟を表示      │  │ Chi: 45/100             │              │
│                     │  │ 仙獣: 麒麟 Lv3 Guard   │              │
│                     │  └─────────────────────────┘              │
│                     │                                           │
│                     │  ┌─────────────────────────┐              │
│                     │  │ 侵入者: 3体接近中       │              │
│                     │  │ 次の波: 8 ticks後       │              │
│                     │  └─────────────────────────┘              │
│                     │                                           │
├─────────────────────┴───────────────────────────────────────────┤
│  [掘る] [通路] [召喚] [強化] [待機]  │  [▶1tick] [▶▶早送り] [⏸] │  ← Action Bar
└─────────────────────────────────────────────────────────────────┘
```

### 3.2 Map View 詳細

**タイルの色分け（仮アセット）:**

| セルタイプ            | 色                                          | 文字ラベル |
| --------------------- | ------------------------------------------- | ---------- |
| Wall（壁）            | 濃い灰色                                    |            |
| RoomFloor（部屋床）   | 属性色（火=赤, 水=青, 木=緑, 金=黄, 土=茶） |            |
| CorridorFloor（通路） | 薄い灰色                                    |            |
| Entrance（入口）      | 白                                          |            |
| DragonHole（龍穴）    | 紫                                          | ★          |
| HardRock              | 非常に濃い灰色                              | ✕          |
| Water                 | 深い青                                      | ～         |

**エンティティの表示（仮アセット）:**

| エンティティ | 表示                    | サイズ            |
| ------------ | ----------------------- | ----------------- |
| 仙獣         | 属性色の円 + 種族頭文字 | 24x24（タイル内） |
| 侵入者       | 赤い三角 + クラス頭文字 | 24x24             |
| 戦闘中       | タイルに赤い枠線        | タイル全体        |

**インタラクション:**

| 操作                             | 効果                                 |
| -------------------------------- | ------------------------------------ |
| セルをホバー                     | ツールチップ（セル情報、部屋情報）   |
| 部屋をクリック                   | Info Panelにその部屋の詳細表示       |
| 壁セルをクリック（掘るモード中） | その位置に部屋掘削のPlayerAction生成 |

### 3.3 アクション操作フロー

**部屋を掘る:**

1. Action Bar の [掘る] をクリック or キー `D` → 掘削モードに入る
2. マップ上の壁セルをクリック → 部屋位置を指定
3. 属性選択パネルが表示（Fire/Water/Wood/Metal/Earth）→ クリックで選択
4. PlayerAction(DigRoom) が pending キューに追加
5. [▶1tick] で実行、または追加アクションを積む

**通路を掘る:**

1. [通路] or キー `C` → 通路モード
2. 始点の部屋をクリック → 終点の部屋をクリック
3. PlayerAction(DigCorridor) がキューに追加

**仙獣を召喚する:**

1. [召喚] or キー `S` → 召喚モード
2. 配置先の部屋をクリック
3. 種族選択パネルが表示 → クリックで選択
4. PlayerAction(SummonBeast) がキューに追加

**部屋をアップグレードする:**

1. [強化] or キー `U` → 強化モード
2. 対象の部屋をクリック
3. PlayerAction(UpgradeRoom) がキューに追加

**待機:**

- [待機] or キー `W` → PlayerAction なしでティック進行

**ティック進行:**

- [▶1tick] or スペースキー → pending キューのアクションで1ティック進行
- [▶▶早送り] or キー `F` → 次の侵入波到達 or N ティックまで自動進行
- [⏸] or Escape → 早送り停止

### 3.4 ウィンドウサイズとCave制約

v1.0.0ではスクロール/ズームを実装しない。Caveが画面に収まるサイズに制限する。

| 要素           | サイズ                    |
| -------------- | ------------------------- |
| タイルサイズ   | 32x32 px                  |
| Map View最大   | 24x20 タイル = 768x640 px |
| Info Panel幅   | 320 px                    |
| Top Bar高さ    | 40 px                     |
| Action Bar高さ | 48 px                     |
| ウィンドウ全体 | 1088x728 px               |

チュートリアル/標準シナリオのCaveサイズがこの制約に収まることを確認する。
収まらない場合はシナリオ側のCaveサイズを調整する。

---

## 4. 仮アセットシステム

### 4.1 設計原則

**全アセットをコードで生成する。画像ファイルは使わない（v1.0.0）。**

```go
// asset/placeholder.go
func GenerateTileset() *ebiten.Image {
    // 32x32 の色付き矩形を属性ごとに生成
    // Wall → 灰色、Fire部屋 → 赤、Water部屋 → 青 ...
}

func GenerateBeastSprite(element types.Element) *ebiten.Image {
    // 24x24 の属性色の円 + 種族頭文字
}

func GenerateInvaderSprite(class string) *ebiten.Image {
    // 24x24 の赤い三角 + クラス頭文字
}
```

### 4.2 将来のドット絵差し替え

```go
// asset/tileset.go
type TilesetProvider interface {
    GetTile(cellType world.CellType, element types.Element) *ebiten.Image
    GetBeastSprite(species string, level int) *ebiten.Image
    GetInvaderSprite(class string) *ebiten.Image
}

// v1.0.0: PlaceholderProvider（コード生成）
// v1.1.0: ImageProvider（ドット絵ファイル読み込み）
```

TilesetProvider インターフェースにより、仮アセットとドット絵を差し替え可能にする。
v1.0.0の全描画コードはTilesetProvider経由でアセットを取得する。

---

## 5. テスト戦略

### 5.1 GUI開発の難しさとRalph Loop適合

GUIの「見た目の正しさ」は go test で自動検証できない。
以下の戦略でテスト可能な部分を最大化する。

| レイヤー           | テスト方法                                 | 自動化     |
| ------------------ | ------------------------------------------ | ---------- |
| controller/        | ティック進行、アクション組み立て、状態遷移 | ✅ go test |
| input/ → action    | 入力座標→PlayerAction変換                  | ✅ go test |
| scene/manager      | シーン遷移ロジック                         | ✅ go test |
| view/mapview       | タイル座標計算、セル→タイルマッピング      | ✅ go test |
| view/entity        | スプライト配置座標計算                     | ✅ go test |
| view/ui            | レイアウト計算、テキスト生成               | ✅ go test |
| asset/             | アセット生成（nil でないこと、サイズ確認） | ✅ go test |
| 描画結果の目視確認 | スクリーンショット比較 or 手動確認         | ❌ 手動    |

### 5.2 Ralph Loop 運用の調整

GUI開発ではtasks.mdの各タスクの完了条件を以下のように工夫する：

- **ロジック部分**: 従来通り go test で検証（controller, input, scene遷移）
- **描画部分**: 「go build が成功する + go vet クリーン」を最低条件とし、目視確認はフェーズ末尾の統合タスクでまとめて行う
- **統合タスク**: フェーズ末尾に「手動プレイ確認」タスクを置く。これはRalph Loopの外（チャットでぽんぽこが実施）

### 5.3 スクリーンショットテスト（将来検討）

Ebitengineは `ebiten.RunGameWithOptions` に `ScreenTransparent` オプションがあり、
描画結果をオフスクリーンバッファに出力可能。将来のリグレッション検出に活用できるが、
v1.0.0ではオーバースペック。

---

## 6. フェーズ分割

### Phase 0: Ebitengine習作 + プロジェクト初期化

**ゴール**: Ebitengineの基本を習得し、タイルマップ描画の最小プロトタイプが動く

タスク:

1. game/ ディレクトリ作成、go.mod初期化、go.work更新
2. Ebitengineの「Hello World」— ウィンドウ表示、矩形描画、マウスクリック座標取得
3. 仮タイルセット生成（PlaceholderProvider）— 全CellType分の32x32矩形
4. coreのCaveデータからタイルマップ描画 — Cave→タイル座標変換、全セル描画
5. マウスホバーでセル座標表示（ツールチップの原型）
6. Makefile、CLAUDE.md更新（game追記）

### Phase 1: Game Controller + ティック進行

**ゴール**: coreのSimulationEngineをGUIから駆動でき、ティック進行/一時停止/早送りが動く

タスク:

1. controller/controller.go: GameController（engine管理、Step呼び出し、Snapshot保持）
2. ティック進行モード3種: 手動（ボタン/スペースキー）、早送り、一時停止
3. シナリオ読み込み（組み込みシナリオ対応）
4. GameSnapshot→描画データ変換: 部屋属性に基づくタイル色変更
5. 仙獣/侵入者のスプライト描画（PlaceholderProvider経由）
6. Top Bar: ChiPool、CoreHP、ティック数の数値表示
7. 統合テスト: GameControllerのティック進行ロジック

### Phase 2: 操作システム

**ゴール**: マウス/キーボードで全PlayerActionが実行できる

タスク:

1. input/action.go: 操作モードステートマシン（Normal, DigRoom, DigCorridor, Summon, Upgrade）
2. Action Bar描画: 5つのアクションボタン + ティック制御ボタン
3. DigRoomフロー: モード切替→壁クリック→属性選択パネル→PlayerAction生成
4. DigCorridorフロー: モード切替→始点部屋クリック→終点部屋クリック→PlayerAction生成
5. SummonBeastフロー: モード切替→部屋クリック→種族選択パネル→PlayerAction生成
6. UpgradeRoomフロー: モード切替→部屋クリック→PlayerAction生成
7. 操作バリデーション: 無効な位置/コスト不足時のフィードバック（赤枠、エラーテキスト）
8. キーボードショートカット: D/C/S/U/W/Space/F/Escape
9. 統合テスト: 入力座標→PlayerAction変換ロジック

### Phase 3: シーン管理 + UI

**ゴール**: タイトル→シナリオ選択→ゲーム→結果の完全なゲームフローが動く

タスク:

1. scene/manager.go: シーン遷移管理（SceneInterface定義、Push/Pop/Switch）
2. scene/title.go: タイトル画面（ゲーム名表示、「開始」「ロード」ボタン）
3. scene/select.go: シナリオ選択画面（組み込みシナリオ一覧、選択→ゲーム開始）
4. scene/result.go: 結果画面（勝利/敗北表示、統計サマリー、「再プレイ」「タイトルへ」ボタン）
5. Info Panel: 選択中の部屋/仙獣/侵入者の詳細表示
6. ツールチップ: マウスホバーでセル/エンティティ情報表示
7. 統合テスト: シーン遷移ロジック

### Phase 4: セーブ/ロード + 仕上げ

**ゴール**: チェックポイント保存/復元が動き、チュートリアルシナリオを完走できる

タスク:

1. save/checkpoint.go: GameControllerの状態をJSON保存/復元（coreのCheckpointを活用）
2. save/config.go: ゲーム設定（ウィンドウサイズ、早送り速度）の保存/復元
3. InGame画面にセーブ/ロードUI統合（キーボードショートカット: Ctrl+S / Ctrl+L）
4. タイトル画面の「ロード」ボタン → ファイル選択 → ゲーム再開
5. 勝敗判定の画面遷移統合（ゲーム終了→Result画面）
6. 戦闘発生時の視覚フィードバック（部屋の枠線を赤く点滅）
7. 侵入波到達時の警告表示

### Phase 5: 統合検証 + リリース

**ゴール**: チュートリアルシナリオをGUIで完走し、D002を体感できる

タスク:

1. チュートリアルシナリオの完走テスト（ぽんぽこが手動プレイ）
2. 標準シナリオの完走テスト（手動プレイ）
3. D002体感フィードバック収集（5つの面白さが機能しているか）
4. パフォーマンス確認: 60FPS維持（M4 Mac mini）
5. HANDOFF.md更新
6. DECISIONS.md更新
7. LESSONS.md更新
8. v1.0.0タグ

---

## 7. 技術方針

### 7.1 依存関係

| ライブラリ    | 用途                  | 備考                         |
| ------------- | --------------------- | ---------------------------- |
| Ebitengine v2 | 2Dゲームエンジン      | ウィンドウ、描画、入力、音声 |
| ebitenutil    | デバッグ表示（FPS等） | Ebitengine付属               |

CLAUDE.mdの方針通り、gameはEbitengine + 必要なライブラリを使用する。

### 7.2 Ebitengine の基本構造

```go
// main.go
type Game struct {
    sceneManager *scene.Manager
}

func (g *Game) Update() error {
    return g.sceneManager.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
    g.sceneManager.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    return 1088, 728  // 固定解像度
}

func main() {
    ebiten.SetWindowSize(1088, 728)
    ebiten.SetWindowTitle("ChaosForge — 風水回廊記")
    ebiten.RunGame(&Game{...})
}
```

### 7.3 描画パフォーマンス

coreのベンチマーク（Phase 7-G）で50部屋のOnCaveChangedが55μs。
60FPSのフレーム予算16.7msに対してcore側の処理はほぼ無視できる。
描画側のボトルネックはタイル数 × Draw呼び出しだが、
24x20 = 480タイル程度ならEbitengineのバッチ描画で余裕。

### 7.4 テスト方針

- ロジック層（controller, input, scene遷移）: go test、テーブル駆動
- 描画層: go build + go vet をCI条件、目視確認はフェーズ末尾
- カバレッジ目標: ロジック層で80%。描画層はカバレッジ対象外

### 7.5 Ralph Loop 適合

- 各タスクの完了条件にはgo testで検証可能な項目を必ず含める
- 描画の「正しさ」はtasks.mdの完了条件に含めない（ループが停止しないため）
- 代わりに「コンパイル成功 + vet クリーン + ロジックテストパス」を条件とする
- フェーズ末尾の統合タスクで「手動プレイ確認」を実施（Ralph Loopの外）

---

## 8. 将来の拡張（v1.0.0 スコープ外）

| 機能                   | 備考                                                  |
| ---------------------- | ----------------------------------------------------- |
| ドット絵アセット       | TilesetProvider差し替えで対応。アセット調達は別途検討 |
| BGM・SE                | Ebitengineのaudioパッケージで対応可能                 |
| 龍脈可視化             | DragonVeinの経路をタイル上にオーバーレイ描画          |
| 気の流れアニメーション | RoomChi変化をパーティクルで表現                       |
| 経済グラフ             | ChiPoolの推移をグラフ表示                             |
| スクロール/ズーム      | 大きなCaveへの対応                                    |
| リプレイ再生UI         | simのReplay機能をGUI化                                |
| ミニマップ             | 大きなCave用の全体俯瞰                                |

---

## 9. D002 との対応関係

simのメトリクス（B01〜B11）は「壊れるサインの自動検出」。
gameは「壊れていない状態で、面白さが本当に機能しているかの人間による判定」。

```
sim Batch Mode  → B01〜B11 アラート0件を保証（自動）
sim Dashboard   → アラートがあれば修正（自動）
game            → アラート0件の状態で、面白いかを判定（人間）
```

**gameの究極的な成果物は「ぽんぽこがD002の5つの面白さを体感できること」。**

Phase 5の手動プレイテストでの判定項目:

| 面白さ       | 判定方法                                             |
| ------------ | ---------------------------------------------------- |
| 計画する     | マップを見て「ここに部屋を掘りたい」と思えるか       |
| 妥協する     | 地形制約で理想が阻まれ、代替案を考える瞬間があるか   |
| 中断される   | 構築中に侵入波が来て「まだ準備できてない！」と焦るか |
| 噛み合う     | 不完全な防衛で敵を撃退し「よし！」と感じるか         |
| 完璧が来ない | クリア時に「次はもっとうまくやれる」と思えるか       |
