# PHASE_COMPLETE_2 — Phase 2: Human Mode 完了報告

## 実装した内容

### Game Server (Phase 1-E〜1-J)
- `server.GameServer`: core の SimulationEngine をラップし、ティックループ・チェックポイント・リプレイを管理
- `server.ActionProvider` インターフェース: Human/AI/Batch 全モード共通のアダプタ境界
- `server.GameContextBuilder`: ライブエンジン状態からメニューコンテキストを構築
- `server.Checkpoint/Replay`: JSON ベースのセーブ/ロード/リプレイ機能
- `metrics.Collector`: ティックごとの統計収集（GameSummary 生成）
- `render.RenderFullStatus`: ANSI カラー付き ASCII マップ + ステータスパネル

### Human Mode Adapter (Phase 2-A〜2-F)
- `human.HumanProvider`: 対話型メニュー駆動の ActionProvider 実装
- `human.InputReader`: スクリプト化可能な入力バリデーション
- メインメニュー: 6 アクション + セーブ/ロード/リプレイ/終了
- サブメニュー: 部屋掘削（座標指定）、通路掘削（部屋ID指定）、仙獣召喚（属性選択）、部屋アップグレード
- 早送り（FastForward）: N ティック分の自動進行 + 差分サマリー表示
- チェックポイント操作: セーブ/ロード/リプレイ保存の UI 統合

### Phase 2-G 統合テスト
- E2E テスト: チュートリアルシナリオのスクリプト実行（dig room → summon beast → FF → game end）
- 全サブメニュー遷移テスト: 全メニューパス + back 遷移 + quit キャンセルの統合検証
- ASCII 描画目視確認スクリプト: ゲーム進行各段階の描画出力（手動確認用）

## 未解決の課題・技術的負債

- **早送り中のゲーム終了**: FF 中にゲームが終了すると FF サマリーが表示されない（正常動作だが UX 改善の余地あり）
- **部屋タイプのソート順依存**: E2E テストの入力シーケンスがアルファベット順ソートに依存（部屋タイプ追加時にテスト修正が必要）
- **セーブファイルパスのバリデーション**: 現状ユーザー入力をそのまま使用（パス検証なし）

## 次フェーズへの申し送り事項

### Phase 3: AI Mode（JSON I/O）
- `server.ActionProvider` インターフェースは確立済み。AI Mode は新しい ActionProvider 実装を作るだけ
- `scenario.GameSnapshot` が JSON シリアライズ可能であることを確認済み（チェックポイントで使用）
- メニューコンテキスト（`BuildContext`, `UnitContext`）の情報を JSON Lines で公開する設計を検討

### Phase 4: Batch Mode + Metrics
- `metrics.Collector` は既に全モードで利用可能。壊れるサイン（B01〜B11）の検出ロジックを追加するだけ
- `server.GameServer.RunGame` の戻り値 `RunResult` に統計情報が含まれている

### DECISIONS.md 未解決課題
- 罠の盗賊回避率: 将来拡張（v1.0.0 スコープ外）
- 侵入者AI高度化: 将来拡張（v1.0.0 スコープ外）
- Phase 2 固有の新規 OPEN 課題はなし

## LESSONS.md からの重要な知見

- HumanProvider の E2E テストはスクリプト化入力で自動化可能だが、メニュー構造への依存に注意
- FF 中のゲーム終了時の挙動（OnGameEnd が先に呼ばれ FF サマリーが出ない）はテスト設計時に考慮が必要
- ASCII 目視確認テストは `-short` フラグでスキップ可能にして CI を軽く保つ
