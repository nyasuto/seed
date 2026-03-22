# PHASE_COMPLETE_0 — Ebitengine 習作 + プロジェクト初期化

## 実装した内容

### Task 0-A: プロジェクト初期化と Ebitengine 導入
- game/ ディレクトリに Ebitengine v2 を導入
- 最小アプリ（空ウィンドウ 1088x728、タイトル "ChaosForge"）
- game/Makefile 作成、ルート Makefile に統合

### Task 0-B: カラーパレットと仮タイルセット生成
- asset/palette.go: 五行属性色、地形色、UI色のカラーパレット
- asset/tileset.go: TilesetProvider インターフェース
- asset/placeholder.go: PlaceholderProvider（全CellType の 32x32 色付き矩形タイル生成）

### Task 0-C: Cave データからタイルマップ描画
- view/mapview.go: MapView（CellPos↔ScreenPos 座標変換、Cave タイルマップ描画）
- チュートリアルシナリオの Cave を読み込みデモ描画

### Task 0-D: マウスホバーとセル情報表示
- input/mouse.go: MouseTracker（マウス座標→セル座標変換、毎フレーム追跡）
- view/tooltip.go: Tooltip（ホバー中のセル情報ツールチップ描画）

### Task 0-E: 統合と確認
- チュートリアルシナリオ Cave サイズ 16x16 → 24x20 ビューポートに収まることを確認
- `go test ./... -count=1 -race` 全パス
- `go vet ./...` クリーン
- DECISIONS.md 棚卸し（Phase 0 game 固有の新規判断なし）
- LESSONS.md に Phase 0 知見追記

## 未解決の課題・技術的負債

- game/testdata/tutorial.json は core の組み込みシナリオと同一内容を個別に embed している。Phase 1 で core の `scenario.LoadBuiltinScenario` に統一すべき
- main.go が `simulation.SimulationEngine` を直接使用している。Phase 1 で GameController に置き換え予定

## 次フェーズへの申し送り事項

- **GameController 導入**: Phase 1-A で core の SimulationEngine をラップする GameController を作成し、main.go の直接参照を解消する
- **シナリオ読み込みの統一**: core の組み込みシナリオ読み込み機構を game でも使う方向に統一する
- **DECISIONS.md**: Phase 0 game で新規設計判断は発生しなかった。既存の D001〜D017b に変更なし

## LESSONS.md から特に重要な知見

- Ebitengine の座標変換（タイルサイズ × セル座標）は整数除算で高速に逆変換可能
- `ebiten.NewImage` + `Fill`/`Set` でプレースホルダータイルを起動時に一括生成するパターンが有効
- core の SimulationEngine 経由で Cave データを直接取得できるが、Phase 1 以降は GameController 経由に移行すべき
