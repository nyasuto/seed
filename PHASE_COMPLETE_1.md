# PHASE_COMPLETE_1 — Phase 1: core修正 + Game Server

## 実装した内容の要約

### core修正（Task 1-A〜1-D）
- **MaxRooms制約チェック**: `validateDigRoom` に `GameConstraints.MaxRooms` チェックを追加。SimpleAIPlayer も制約を尊重するよう修正
- **SimpleAIPlayerコリドー戦略**: 部屋建設後に隣接部屋への通路を自動的に掘るロジックを追加。気の伝播経路を確保
- **caveScore正規化**: `CaveTotal` を `MaxPossibleScore` で割って [0,1] に正規化してから `CalcTickSupply` に渡すよう修正
- DECISIONS.md に D014〜D016 を記載

### Game Server（Task 1-E〜1-I）
- **ActionProvider インターフェース**: `ProvideActions`, `OnTickComplete`, `OnGameEnd` の3メソッド
- **GameServer**: core の `SimulationEngine` をラップし、tick ループを駆動
- **セッション管理**: 組み込みシナリオ（tutorial, standard）の embed 読み込み + ファイルパス読み込み
- **チェックポイント**: ゲーム状態の JSON 保存/復元。復元後の続行で元実行と同一結果を保証
- **リプレイ**: アクション履歴の保存/再生。決定論的再現を保証
- **Metrics Collector**: ティックごとの統計収集と GameSummary 生成

### 統合テスト（Task 1-J）
- tutorial シナリオ: NoAction プロバイダーで完走、survive_until 条件で勝利（299 ticks）
- standard シナリオ: NoAction プロバイダーで完走、defeat_all_waves 条件で勝利（810 ticks, 5/5 waves）

## 未解決の課題や技術的負債

- `lazyAIProvider` は NoAction を返すだけで SimpleAIPlayer をラップしていない。GameServer の ActionProvider 経由で SimpleAIPlayer を完全に活用するには、エンジン状態へのアクセス方法を整理する必要がある
- 統合テストは NoAction プロバイダーで実行しているため、AI の行動（部屋建設、コリドー掘削）を伴うシナリオの統合テストは未実施
- Metrics Collector の `RoomsBuilt` は DigRoom アクション数のカウントであり、実際に成功した建設数とは異なる可能性がある

## 次フェーズへの申し送り事項

### DECISIONS.md 未解決課題
- 罠の盗賊回避率（将来拡張、v1.0.0 スコープ外）
- 侵入者AI高度化（将来拡張、v1.0.0 スコープ外）

### 技術的な申し送り
- GameServer は毎回新しい `SimulationEngine` を作成する設計。エンジンの再利用は想定していない
- ActionProvider インターフェースは同期的。非同期（チャネルベース等）が必要になった場合は adapter 層で吸収する想定
- 組み込みシナリオ JSON は `sim/server/scenarios/` に配置し `//go:embed` で埋め込み

## LESSONS.md から特に重要な知見

- NoAction プロバイダーでも両シナリオが完走する。仙獣の初期配置だけで防衛が成立するため、AI の介入なしでもゲームエンジンの堅牢性を検証可能
- `LoadBuiltinScenario` でシナリオ読み込みとゲーム完走を一括テストする方がテスト用 JSON 手書きよりも実用的
