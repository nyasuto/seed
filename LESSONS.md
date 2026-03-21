# LESSONS.md — 学んだこと

## Phase 1-D: 通路の生成

- CellType の定数 `Corridor` と構造体 `Corridor` が名前衝突した。CellType の定数を `CorridorFloor` にリネームして解決。同一パッケージ内で型名と定数名の衝突に注意。
- BuildCorridor は BFS で最短経路を探索するため、fromRoomID/toRoomID を引数に取り、他部屋のRoomFloorセルを回避する設計とした。

## Phase 0-B: エコシステム整備

- golangci-lint v2 では設定ファイルに `version: "2"` が必須。v1 形式の設定はエラーになる。
- golangci-lint v2 では `typecheck` は独立 linter ではなくなった（常に有効）。`gosimple` は `staticcheck` に統合された。enable リストに含めるとエラーになる。
