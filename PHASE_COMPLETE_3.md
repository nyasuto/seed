# Phase 3 完了報告: AI Mode

## 実装した内容

### Phase 3-A: AI Mode プロトコル定義
- `adapter/ai/protocol.go`: StateMessage, ActionMessage, GameEndMessage, ErrorMessage の型定義
- `adapter/ai/docs/PROTOCOL.md`: プロトコル仕様書

### Phase 3-B: AI Mode valid_actions 生成と Snapshot 変換
- `adapter/ai/serializer.go`: SnapshotToJSON, BuildValidActions, StateBuilder
- 全アクション種別（dig_room, dig_corridor, summon_beast, upgrade_room, evolve_beast, place_beast, wait）の valid_actions 生成

### Phase 3-C: AI Mode ActionProvider 実装とエラーハンドリング
- `adapter/ai/provider.go`: AIProvider（ActionProvider 実装）
- エラーリトライ（最大3回）、タイムアウト対応、valid_actions バリデーション
- parseAndValidate でクライアント入力を検証

### Phase 3-D: AI Mode CLI 統合と E2E テスト
- `cmd/chaosseed-sim/main.go`: `--ai` モードと `--timeout` オプション
- E2E テスト: WaitOnly 完走、BasicStrategy 完走

### Phase 3-E: Phase 3 統合テストと完了
- 統合テスト7件追加:
  - ErrorRetry: 不正 JSON 送信 → エラー・リトライ → 正常復帰
  - InvalidAction: valid_actions にないアクション → エラー・リトライ
  - StandardScenario: standard シナリオでの AI Mode 動作
  - ValidActionsContent: state メッセージの valid_actions 検証
  - GameEndMessage: game_end メッセージのフィールド検証
  - ModeCoexistence: Human Mode と AI Mode の同一シナリオ共存テスト
  - SameScenarioBothModes: 同一 seed での決定論性検証

## 未解決の課題・技術的負債

- `go test -v` でテスト名が表示されないケースがある（テスト自体はパスしている）
- AI Mode のタイムアウトテストは goroutine リークの可能性があるため、本番運用では goroutine の適切なクリーンアップが必要

## 次フェーズへの申し送り事項

### Phase 4: Batch Mode + 壊れるサイン検出
- `metrics/collector.go` は既に GameSummary の基本構造を持つ。B01〜B11 のメトリクスはここに追加する
- `adapter/batch/` パッケージを新規作成し、ヘッドレス並列実行を実装する
- Batch Mode は ActionProvider インターフェースを実装すれば GameServer.RunGame() で動作する

### DECISIONS.md の未解決課題
- D017 (AI Mode プロトコル設計): ACTIVE — プロトコルは安定しており変更予定なし
- 罠の盗賊回避率: 将来拡張（v1.0.0 スコープ外）
- 侵入者AI高度化: 将来拡張（v1.0.0 スコープ外）

## LESSONS.md の重要な知見

- AI Mode の E2E テストは io.Pipe() + goroutine パターンで実現。clientErr チャネルでエラーを伝播する
- Human Mode と AI Mode は ActionProvider インターフェースにより完全に交換可能。同一 GameServer インスタンスで切り替えテスト可能
- 同一 seed + 同一アクション（wait）で NoAction プロバイダーと AI Mode は同一結果を返す（決定論性の保証）
