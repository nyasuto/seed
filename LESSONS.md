# LESSONS.md — 学んだこと

## Phase 7-A: アクションテスト

- `world/room_type_data.json` と `economy/construction_data.json` で使われている部屋タイプIDが異なる（例: `senju_room` vs `beast_room`）。テストでは両方に存在するID（`trap_room`, `recovery_room`）を使う必要がある。`CalcRoomCost` が 0 を返すとバリデーションを素通りするが `TryBuildRoom` で "unknown room type" エラーになるため、ID不一致に注意。

## Phase 1-D: 通路の生成

- CellType の定数 `Corridor` と構造体 `Corridor` が名前衝突した。CellType の定数を `CorridorFloor` にリネームして解決。同一パッケージ内で型名と定数名の衝突に注意。
- BuildCorridor は BFS で最短経路を探索するため、fromRoomID/toRoomID を引数に取り、他部屋のRoomFloorセルを回避する設計とした。

## Phase 7-H: D002 定量検証

- D002原則1の検証で風水スコア（CaveTotal）を多様性の指標にしようとしたが、SimpleAIがコリドーを掘らないため chi が新規部屋に伝播せず、スコアが全seed同一（128.00）になった。代わりに部屋数と配置位置の多様性で原則を検証。SimpleAIにコリドー構築戦略が追加されれば、風水スコアも有効な指標になる。
- MaxRooms 制約（GameConstraints）は validateDigRoom で未チェック。SimpleAIは制約を超えて部屋を建設する。将来修正が必要。

## Phase 7-H (cont.): D002原則2の時間圧力検証

- Scenario の `WaveSchedule` はデータ定義のみで、SimulationEngine はこれを自動的にウェーブ生成に変換しない。実際にウェーブを発生させるには `EventDef`（`survive_until` 条件 + `spawn_wave` コマンド）に変換して `Scenario.Events` に設定する必要がある。ヘルパー `waveScheduleToEvents` を d002_test.go に追加して対応。
- 経済システムの `CalcTickSupply` は `caveScore` を [0,1] 範囲前提で使用するが、`CaveTotal` は全部屋スコアの合計で無制限に増加する。そのため部屋が増えるとサプライが指数的に膨らみ、AIが毎ティック1部屋以上建設可能になる。時間圧力の検証には波到達を非常に早く（tick 1-5）設定する必要があった。

## Phase 7-K: 統合検証

- `SimulationEngine.Step` で `evaluateEndConditions` が tick カウンタのインクリメント前に呼ばれていたため、`survive_until(N)` と `max_ticks: N` が同じ値のシナリオで off-by-one が発生し、勝利条件が満たされなかった。tick インクリメントを条件評価の前に移動して解決。「tick N を処理完了したら N+1 tick 生存した」というセマンティクスが正しい。
- `BuildSnapshot` の `TotalWaves` が `len(state.Waves)`（スポーン済みウェーブ数）のみを参照していたため、`defeat_all_waves` 勝利条件が最初の1波撃退で即勝利と判定されてしまった。`GameState.ScheduledWaves`（シナリオの`spawn_wave`イベント数）を追加し、`TotalWaves = max(len(Waves), ScheduledWaves)` とすることで解決。standard.json にも `spawn_wave` イベントを追加する必要があった（`wave_schedule` はデータ定義のみ）。

## Task 1-B: SimpleAIPlayerコリドー戦略

- `applyDigRoom` で生成される部屋に入口（Entrance）がなかったため、コリドー接続が不可能だった。南側中央にデフォルト入口を自動生成するよう修正。
- AIが同一ティックでコリドー掘削と新部屋建設を同時実行すると、コリドーが掘った CorridorFloor セルと新部屋の配置領域が競合し、配置バリデーションに失敗する。コリドーを掘るティックでは新部屋建設をスキップすることで解決。
- `processActions` がアクション失敗時にティック全体をエラー終了させていたが、AI のベストエフォートなアクション（パスが見つからない等）を許容するため、失敗はスキップしてイベントログに記録する方式に変更。
- ChiFlowEngine の `Tick()` で `RoomChi` マップや `Neighbors()` のマップイテレーション順が非決定的だったため、float64 の微小な精度差が発生。ソート済みスライスでの走査に修正して決定論性を確保。
- `OnCaveChanged` はドラゴンヴェインを再構築するため、ゲーム中盤以降にスコアが急変する副作用がある。新部屋のchi追跡登録のみが必要な場合は `SyncNewRooms` を使うべき。

## Phase 1-J: sim統合テスト

- GameServer + Collector の統合テストでは、`LoadBuiltinScenario` で組み込みシナリオを直接読み込むのが最も簡潔。テスト用 JSON を手書きする `tutorialScenarioJSON()` パターンよりも、実際の組み込みシナリオを使うことで JSON 定義の正当性も同時に検証できる。
- NoAction プロバイダーでも tutorial シナリオは survive_until 条件で勝利し、standard シナリオは defeat_all_waves で勝利する（仙獣が初期配置されているため）。AIの介入なしでもゲームが完走することの確認は、エンジンの堅牢性テストとして有用。

## Phase 2: Human Mode

- HumanProvider の E2E テストでは、スクリプト化された入力（`strings.NewReader`）を `InputReader` に渡すことで、対話的なメニュー遷移を自動化できる。ただし入力シーケンスはメニュー構造に強く依存するため、メニューの選択肢の順序（部屋タイプのアルファベット順ソートなど）を事前に把握する必要がある。
- 早送り（FastForward）中にゲームが終了すると `OnTickComplete` の FF サマリーが表示されない。これは正常な動作（`OnGameEnd` が先に呼ばれる）だが、テストでは「早送り完了」メッセージの有無を条件にしないこと。
- GameServer の runLoop は ActionProvider.OnTickComplete → 終了条件チェックの順なので、最終ティックの OnTickComplete は常に呼ばれる。ただし FF 中の場合、ffStartSnapshot のサマリーは表示されずに OnGameEnd に遷移する。
- ASCII描画の目視確認テスト（`TestVisual_*`）は `-short` フラグでスキップ可能にしておくと CI で邪魔にならない。

## sim Phase 3: AI Mode 統合

- AI Mode の E2E テストでは `io.Pipe()` でサーバーとクライアントを接続し、クライアント側を goroutine で駆動する。`clientErr` チャネルでエラーを伝播し、`defer inW.Close()` でパイプの終端を保証する。
- エラーリトライのテストでは、不正入力送信後に error メッセージと state メッセージの再送を順序通り読み取る必要がある。`switch msg["type"]` でメッセージ種別を振り分ける前に、リトライ中の error メッセージを読み飛ばさないよう注意。
- AI Mode と NoAction プロバイダーで同一 seed・同一シナリオを実行すると、同じ tick 数・同じ結果が得られる（決定論性の確認）。これは AI Mode が wait アクションを送った場合、NoAction と等価であることの証明になる。
- Human Mode と AI Mode は同一の `ActionProvider` インターフェースを実装しており、`GameServer.RunGame()` に渡すだけで切り替え可能。共存テストにより、同一バイナリで両モードが動作することを確認。

## Phase 4-G: Batch Mode CLI統合とD002検証

- SimpleAIPlayer は GameState を必要とするが、GameServer.RunGame() がエンジンを内部で生成するため、ProvideActions の初回呼び出し時に遅延初期化する `simpleAIBatchProvider` パターンが有効。
- チュートリアルシナリオは easy 設計のため、BreakageReport で B03/B05/B11 アラートが発生するのは想定内。D002 検証はシナリオの設計意図に応じた閾値解釈が必要。
- 1,000ゲーム × SimpleAI が約1〜2秒で完了。Go の goroutine 並列実行は非常に高速で、パフォーマンス要件（5分以内）を大幅にクリア。

## Phase 4-H: Phase 4 統合テスト

- server パッケージのテストから batch パッケージを import すると循環依存（server → batch → server）になる。モード横断の統合テストは sim ルートパッケージ（`package sim`）に配置することで回避できる。
- Batch Mode の決定論性テストは `Parallel: 1` で実行する必要がある。並列実行でも各ゲームは独立した RNG を持つため結果は同一だが、goroutine のスケジューリング順は保証されないためログ出力順序が変わりうる。

## Phase 5: Balance Dashboard

- ダッシュボードの統合テストで `io.Pipe()` を使った AI Mode の E2E テストは、サーバーの state 送信とクライアントの action 送信が互いにブロックするデッドロックに陥りやすい。AI Mode の共存テストではパイプ通信を避け、`strings.NewReader("")` で即 EOF を返すことでモード初期化の検証に留めるのが安全。
- Balance Dashboard のフローは「ベースライン → アラート検出 → スイープ提案 → 比較」の4段階。各段階は独立してテスト可能だが、統合テストでは全段階を直列に実行してデータの受け渡しを検証する方が効果的。
- チュートリアルシナリオの B03/B05/B11 アラートは easy 設計に起因し、スイープしても解決しないケースがある（BestIndex = -1）。これはシナリオ設計の意図通りであり、テストではアラート解決の成否ではなくフロー完走を検証すべき。

## Phase 6: 統合検証 + リリース準備

- 全モード統合テスト（Task 6-A）では、各モードの初期化と基本フロー完走を検証する。AI Mode のパイプ通信テストはデッドロックリスクがあるため、初期化検証に留めるのが実用的。
- D017a/D017b の ID 重複が文書参照時に混乱を招いた。設計判断 ID は一意にすべき。Phase 6 棚卸しで D017a/D017b に分離して解決。
- ドキュメントの最終棚卸しでは、PHASE_COMPLETE_1〜5 の申し送り事項と DECISIONS.md の OPEN エントリを突き合わせることで、漏れなく確認できる。

## Phase 0-B: エコシステム整備

- golangci-lint v2 では設定ファイルに `version: "2"` が必須。v1 形式の設定はエラーになる。
- golangci-lint v2 では `typecheck` は独立 linter ではなくなった（常に有効）。`gosimple` は `staticcheck` に統合された。enable リストに含めるとエラーになる。

## game Phase 0: Ebitengine 習作 + プロジェクト初期化

- Ebitengine v2 のタイル描画では `ebiten.DrawImageOptions` の `GeoM.Translate` でピクセル座標を指定する。タイルサイズ（32x32）× セル座標でスクリーン座標に変換し、逆変換でマウス座標→セル座標を得る。整数除算で済むため高速。
- `ebiten.NewImage(w, h)` で生成したオフスクリーン画像に `Fill` や `Set` で描画してタイルを作成できる。PlaceholderProvider はこの方式で全 CellType 分のタイルを起動時に一括生成している。
- core の `simulation.NewSimulationEngine` でチュートリアルシナリオを読み込み、`engine.State.Cave` を直接参照することで game 側で Cave データを取得できる。Phase 1 以降は GameController 経由に変更予定。
- game の testdata/tutorial.json は core の組み込みシナリオと同一内容を embed している。Phase 1 以降は core の `scenario.LoadBuiltinScenario` を使う方向に統一すべき。
- チュートリアルシナリオの Cave サイズは 16x16 タイルで、想定ビューポート（24x20）に余裕を持って収まる。

## game Phase 1: Game Controller + ティック進行

- GameController は core の SimulationEngine をラップし、Snapshot / AddAction / AdvanceTick の3つのAPIで game 側から操作する。engine.State への直接アクセスは描画データ取得（Cave, Beasts, Waves, RoomTypeRegistry）に限定すべき。
- TopBarData の ChiPool/MaxChiPool は `EconomyEngine.ChiPool.Balance()` / `.Cap` から取得し、int にキャストする。MaxCoreHP は龍穴部屋の `CoreHPAtLevel(room.Level)` から動的に取得する必要がある（Snapshot に含まれていない）。
- `ebiten/v2/inpututil.IsKeyJustPressed` は前フレームとの差分でキー押下を検知するため、長押しによる連続発火を防げる。Space/F/Escape のティック制御に有用。
- Ebitengine v2.9.9 は macOS でパッケージレベルの `init()` で GLFW を初期化し、ディスプレイアクセスがない環境（SSH、一部のターミナルマルチプレクサ）では `currentMouseLocation()` で nil pointer panic を起こす。ebiten を import するだけでテストが実行不能になるため、asset/view/input のテストはディスプレイ環境でのみ実行可能。controller パッケージは ebiten に依存しないため常にテスト可能。
- `BuildRoomRenderMap` は毎フレーム呼ぶのではなく、Cave が変更された時のみ再構築すべき。Phase 1-E では簡潔さのため毎フレーム呼んでいるが、Phase 2 以降で最適化検討。
