# Phase 4 Complete — Batch Mode & メトリクス基盤

## 実装した内容

### Phase 0: プロジェクト初期化
- sim スケルトン作成（go.work, go.mod, CLIエントリポイント）
- sim 用 Makefile と golangci-lint 設定

### Phase 1: core 修正 + Game Server
- core 修正: MaxRooms 制約チェック、SimpleAI コリドー戦略、caveScore 正規化
- GameServer: ActionProvider インターフェース、セッション管理、チェックポイント/リプレイ
- Metrics Collector: per-tick 統計収集、GameSummary 生成

### Phase 2: Human Mode
- HumanProvider: メニュー駆動 TUI（6 メインアクション + サブメニュー）
- 早送り（FastForward）機能
- E2E テスト: スクリプト化された入力でフルプレイスルー

### Phase 3: AI Mode
- AIProvider: JSON Lines プロトコル（StateMessage, ActionMessage, GameEndMessage, ErrorMessage）
- valid_actions ホワイトリスト強制
- エラーリトライ（最大3回）、タイムアウト対応
- モード共存テスト（Human + AI、同一 seed で同一結果）

### Phase 4: Batch Mode & メトリクス
- BatchRunner: 並列バッチ実行（goroutine ベース）
- BreakageDetector: B01〜B11 壊れるサインメトリクス（41 テストケース）
- レポート生成: JSON（PRD セクション 3.3 準拠）、CSV
- パラメータスイープ: RunSweep API（dotted path による JSON パラメータ変更）
- CLI 統合: `--batch` モード（`--games`, `--seed`, `--ai`, `--parallel`, `--sweep`, `--report-format` オプション）
- D002 検証: 1,000 ゲーム × SimpleAI のBreakageReport（critical アラート 0 件確認）

## テスト状況

- `cd sim && go test ./... -count=1 -race`: 全パス（約 30 秒）
- `cd sim && go vet ./...`: クリーン
- テストカバレッジ: 87.9%
- Phase 4 統合テスト（sim ルートパッケージ）:
  - 3 モード共存テスト（Human / AI / Batch）
  - Batch フルパイプライン（run → metrics → breakage → report）
  - Batch 決定論性テスト（同一 seed = 同一結果）
  - パラメータスイープ統合テスト

## パフォーマンス

- 1,000 ゲーム × SimpleAI: 約 1〜2 秒（5 分制限を大幅にクリア）
- Go goroutine による並列実行が非常に効果的

## 未解決の課題・技術的負債

- B10（LayoutEntropy）は room 位置データが GameSnapshot に含まれないため、バッチ後の外部計算が必要
- チュートリアルシナリオの B03/B05/B11 アラートは easy 設計に起因（バランス問題ではない）

## DECISIONS.md ステータス

- OPEN エントリ: なし
- RESOLVED: D001, D004, D006, D007, D014, D015, D016, D017（2件）
- ACTIVE: D002, D003, D005, D008, D009, D010, D011, D012, D013

## 次フェーズ（Phase 5）への申し送り

1. **バランス調整ダッシュボード**: `sim/balance/` パッケージを新規作成。BreakageReport をベースにインタラクティブなダッシュボードを構築
2. **スイープ提案**: アラート発生時に、どのパラメータをどの方向にスイープすべきか提案する機能
3. **パラメータ適用**: ダッシュボードからパラメータ変更を適用し、再バッチ実行する流れ
4. **B10 LayoutEntropy**: room 位置データの収集と entropy 計算の統合が必要

## LESSONS.md から特に重要な知見

- server → batch → server の循環依存を避けるため、モード横断テストは sim ルートパッケージに配置
- SimpleAIPlayer の遅延初期化パターン（`simpleAIBatchProvider`）は、GameServer が内部でエンジンを生成する設計と相性が良い
- D002 検証はシナリオの設計意図に応じた閾値解釈が必要（tutorial の B03/B05/B11 は想定内）
