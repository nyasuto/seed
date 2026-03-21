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

## Phase 0-B: エコシステム整備

- golangci-lint v2 では設定ファイルに `version: "2"` が必須。v1 形式の設定はエラーになる。
- golangci-lint v2 では `typecheck` は独立 linter ではなくなった（常に有効）。`gosimple` は `staticcheck` に統合された。enable リストに含めるとエラーになる。
