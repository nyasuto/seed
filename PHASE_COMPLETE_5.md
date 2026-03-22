# PHASE_COMPLETE_5 — Phase 5: 統合検証 + リリース準備

## 実装した内容

### Task 5-C: ドキュメント更新とリリース準備

- **HANDOFF.md**: game の完了フェーズ（Phase 0〜4）、game アーキテクチャ（GameController + SceneManager + View/Input 分離）、TilesetProvider の仮アセットシステム、未解決の将来課題を追記
- **DECISIONS.md**: game 全フェーズの最終棚卸し。ACTIVE 12件、RESOLVED 9件を確認。v1.0.0 スコープ外の将来課題3件と技術的負債3件を明記
- **LESSONS.md**: Phase 5 の知見追記（ebiten 依存/非依存パッケージの分離、3レイヤーモノレポの独立性、仮アセットシステムの設計判断）
- **game/README.md**: ビルド方法、テスト方法、操作方法（キーボードショートカット一覧）、アーキテクチャ、シナリオ一覧、セーブデータパスを記載

### Task 5-A / 5-B: 手動プレイテスト（Ralph Loop 外）

手動プレイテスト（Task 5-A: チュートリアル、Task 5-B: 標準シナリオ + パフォーマンス確認）は Ralph Loop 外で実施予定。

## テスト結果

- `go test ./core/... -count=1 -race`: 全パス
- `go test ./sim/... -count=1 -race`: 全パス
- `cd game && go test ./controller/... ./save/... -count=1 -race`: 全パス（headless-safe）
- `cd game && go vet ./...`: クリーン

## 未解決の課題

### 技術的負債（v1.0.0 で許容）
- `game/testdata/tutorial.json` の embed 重複（core の `LoadBuiltinScenario` への統一は将来対応）
- `BuildRoomRenderMap` の毎フレーム再構築（パフォーマンス問題が顕在化した場合に対応）
- ebiten 依存パッケージのテストが headless 環境で実行不可（GLFW init panic）

### 将来の拡張課題
- ドット絵アセットの作成（現在は PlaceholderProvider の色付き矩形）
- BGM/SE の追加
- 龍脈の可視化（気の流れをパーティクルやアニメーションで表現）
- 罠の盗賊回避率（盗賊の SPD による罠回避メカニクス）
- 侵入者AIの高度化（複数ステップ先読み、仙獣回避）
- standard シナリオ向けのスイープ値チューニング

## DECISIONS.md 棚卸し

全 D001〜D020 を確認済み。Phase 5 で新たな設計判断は発生せず。
- ACTIVE 12件: D002, D003, D005, D008, D009, D010, D011, D012, D013, D017a, D019, D020
- RESOLVED 9件: D001, D004, D006, D007, D014, D015, D016, D017b, D018
- v1.0.0 スコープ外3件: 罠の盗賊回避率、侵入者AI高度化、standardスイープ

## LESSONS.md から特に重要な知見

- ebiten 依存パッケージと非依存パッケージの分離が headless テスト（CI）の鍵
- core→sim→game の3レイヤーモノレポで game は core のみに依存し sim には非依存の設計を維持
- PlaceholderProvider による仮アセットシステムは TilesetProvider インターフェースの抽象化により将来の差し替えが容易
