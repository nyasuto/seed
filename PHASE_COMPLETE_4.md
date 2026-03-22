# PHASE_COMPLETE_4 — Phase 4: セーブ/ロード + 仕上げ

## 実装した内容

Phase 4 ではセーブ/ロード機能、タイトル画面のロード UI、戦闘・侵入の視覚フィードバックを実装した。

### Task 4-A: セーブ/ロード
- `save/checkpoint.go`: チェックポイント保存/復元（core の SimulationEngine.Checkpoint/RestoreFromCheckpoint を活用）
- `save/config.go`: ゲーム設定の保存/復元（`~/.chaosforge/config.json`）
- `controller/checkpoint.go`: GameController レベルのセーブ/ロード統合
- InGame シーンに Ctrl+S（セーブ）/ Ctrl+L（ロード）を統合

### Task 4-B: タイトル画面のロード機能
- `scene/title.go`: [ロード] ボタン実装 — セーブファイル一覧表示と選択
- `scene/load.go`: LoadScene — セーブファイル一覧の表示・選択・InGame シーンへの遷移
- セーブファイルが存在しない場合は [ロード] ボタンをグレーアウト

### Task 4-C: 戦闘・侵入の視覚フィードバック
- `view/battle_feedback.go`: BattleFeedbackOverlay — 戦闘中部屋の点滅、侵入波警告、CoreHP バー点滅、仙獣敗北時の点滅
- 前ティックと現ティックの Wave 数比較で侵入波到達を検知
- CoreHP 減少時の赤色点滅表示

### Task 4-D: Phase 4 統合と確認
- `go build ./...` 成功
- `go vet ./...` クリーン
- `go test ./controller/... ./save/... -count=1 -race` パス（headless-safe）
- ebiten 依存パッケージ（asset/view/input/scene）のテストはディスプレイ環境でのみ実行可能（既知の制約）
- core テスト全パス（リグレッションなし）
- DECISIONS.md 棚卸し
- LESSONS.md Phase 4 知見追記

## セーブ/ロードフロー

```
InGame → Ctrl+S → ~/.chaosforge/saves/save_YYYYMMDD_HHMMSS.json 保存
InGame → Ctrl+L → 最新セーブからロード → ゲーム状態復元
Title → [ロード] → LoadScene（一覧表示）→ 選択 → InGame（チェックポイントから復元）
```

## 未解決の課題

- `game/testdata/tutorial.json` の embed 重複（core の `LoadBuiltinScenario` への統一は将来対応）
- `BuildRoomRenderMap` の毎フレーム再構築（パフォーマンス問題が顕在化した場合に対応）
- ebiten 依存パッケージのテストが headless 環境で実行不可（GLFW init panic）

## DECISIONS.md 棚卸し

全 D001〜D020 を確認済み。Phase 4 で新たな設計判断は発生せず。
- D019 (SummonBeast Element選択): ACTIVE — 継続有効
- D020 (Scene Draw image.Image): ACTIVE — 継続有効
- 未解決3件（罠の盗賊回避率、侵入者AI高度化、standardスイープ）は v1.0.0 スコープ外で変更なし

## 次フェーズへの申し送り

- Phase 5 は統合検証 + リリース（手動プレイテスト + ドキュメント更新）
- Task 5-A/5-B は Ralph Loop の外で手動実施
- Task 5-C でドキュメント最終更新とリリース準備
- セーブ/ロード、全操作フロー、戦闘フィードバックが機能する状態で Phase 5 に進める

## LESSONS.md から特に重要な知見

- Ebitengine v2.9.9 は macOS で GLFW init 時に headless 環境では panic する。ebiten を import するだけでテストが実行不能になるため、テスト可能なパッケージ（controller, save）と ebiten 依存パッケージの分離が重要
- Scene インターフェースの Draw を `image.Image` にすることで SceneManager テストが headless 対応可能（D020）
- セーブ/ロードは core の SimulationEngine.Checkpoint/RestoreFromCheckpoint を活用し、GameController レベルで統合する設計が明快
