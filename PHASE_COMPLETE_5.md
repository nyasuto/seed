# PHASE_COMPLETE_5.md — Phase 5: Balance Dashboard 完了報告

## 実装した内容の要約

### Phase 5 全タスク

| タスク | 内容 | 状態 |
|--------|------|------|
| 5-A | ダッシュボード ベースライン実行と壊れるサイン表示 | ✅ |
| 5-B | ダッシュボード スイープ提案と比較 | ✅ |
| 5-C | ダッシュボード パラメータ適用とCLI統合 | ✅ |
| 5-D | Phase 5 統合テストと完了 | ✅ |

### 機能概要

1. **Balance Dashboard** (`sim/balance/`): ベースラインバッチ実行 → 壊れるサイン（B01〜B11）レポート → D002ルールに基づくスイープ提案 → パラメータスイープ比較の4段階フロー
2. **CLI統合** (`--balance` フラグ): `chaosseed-sim --balance --scenario tutorial --games 100` でダッシュボード実行
3. **パラメータ適用** (`balance.ApplyParameter`): シナリオJSONのパラメータを安全に変更（.bak バックアップ付き）

### テスト

- `sim/balance/` — 25+テスト関数（ユニット + 統合）
- `sim/cmd/chaosseed-sim/` — 全モード共存テスト（Human/AI/Batch/Balance）
- 全テスト `-race` 付きでパス

## 未解決の課題や技術的負債

- チュートリアルシナリオのスイープでは B03/B05/B11 が解決しない（easy設計による想定内のアラート）。standard シナリオ向けのスイープ値チューニングは Phase 6 以降
- `runBalanceMode` は os.Stdout/os.Stdin をハードコードしており、テストでの出力キャプチャに制約がある

## 次フェーズへの申し送り事項

### Phase 6: 統合検証 + リリースに向けて

1. **全モード統合テスト**: Human/AI/Batch/Balance の各モードでチュートリアル・standard シナリオを完走させる E2E テスト
2. **リプレイ決定論性**: Batch で記録 → `--replay` で再生 → 同一結果の検証
3. **チェックポイント**: Human Mode でセーブ → `--checkpoint` で復元の検証
4. **ドキュメント**: HANDOFF.md 新規作成、sim/README.md の全モード対応

### DECISIONS.md 棚卸し結果

- D001〜D017: 全エントリ確認済み、Phase 5 で新たに RESOLVED/OPEN になるものなし
- D017 (AI Mode プロトコル): ACTIVE のまま維持
- D017 (D002検証): RESOLVED のまま維持（重複IDだが内容は異なる）
- 未解決課題（罠の盗賊回避率、侵入者AI高度化）: v1.0.0 スコープ外で変更なし

## LESSONS.md から特に重要な知見

- `io.Pipe()` による双方向通信テストはデッドロックに注意。共存テストでは初期化検証に留める
- ダッシュボードの統合テストでは「アラート解決の成否」より「フロー完走」を検証すべき
