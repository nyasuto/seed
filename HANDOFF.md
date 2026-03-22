# HANDOFF.md — chaosseed プロジェクト引き継ぎ文書

## プロジェクト概要

カオスシード（風水回廊記）インスパイアのダンジョン経営シミュレーション。
Go モノレポ構成で、core（ゲームロジックライブラリ）と sim（CLIシミュレーター）を実装済み。

## 完了フェーズ

### core（v1.1.0）

7つのサブシステムを段階的に構築:

| パッケージ | 役割 |
|-----------|------|
| `types` | 共有型定義（Element, Direction, RNG 等） |
| `world` | マップシステム（Cave, Room, Corridor, 地形生成） |
| `fengshui` | 風水システム（ChiFlowEngine, DragonVein, 五行相性） |
| `senju` | 仙獣システム（Beast, GrowthEngine, BehaviorEngine, EvolutionEngine） |
| `invasion` | 侵入システム（InvasionEngine, WaveSchedule, 戦闘解決） |
| `economy` | 経済システム（ChiPool, SupplyCalculator, 維持コスト） |
| `scenario` | シナリオシステム（EventEngine, EventCommand, 勝敗条件） |
| `simulation` | 統合シミュレーション（SimulationEngine, 12ステップ Tick 更新順序） |

### sim（v1.0.0）

Phase 0〜6 で CLIシミュレーターを完成:

| Phase | 内容 |
|-------|------|
| 0 | プロジェクト初期化（スケルトン、Makefile） |
| 1 | core修正 + Game Server（ActionProvider, セッション管理, チェックポイント, リプレイ, Metrics Collector） |
| 2 | Human Mode（ターミナルUI、メニュー操作、早送り） |
| 3 | AI Mode（JSON Lines プロトコル、エラーリトライ、タイムアウト） |
| 4 | Batch Mode（並列バッチ実行、B01〜B11壊れるサイン検出、D002検証） |
| 5 | Balance Dashboard（ベースライン → アラート → スイープ提案 → 比較の4段階フロー） |
| 6 | 統合検証 + ドキュメント + リリース準備 |

## アーキテクチャ

```
chaosseed-sim CLI
    ├── --human  → adapter/human  → GameServer → core.SimulationEngine
    ├── --ai     → adapter/ai     → GameServer → core.SimulationEngine
    ├── --batch  → adapter/batch  → GameServer → core.SimulationEngine
    ├── --balance→ balance        → adapter/batch → GameServer
    └── --replay → server.Replay  → GameServer → core.SimulationEngine
```

### Game Server（`sim/server/`）

core の `SimulationEngine` をラップし、以下を提供:
- `ActionProvider` インターフェース経由でアダプター層と接続
- 組み込みシナリオ（tutorial, standard）の `//go:embed` 読み込み
- チェックポイント（ゲーム状態の JSON 保存/復元）
- リプレイ（アクション履歴の保存/再生、決定論的再現保証）
- Metrics Collector（ティックごとの統計収集、GameSummary 生成）

### 3つのアダプター

| アダプター | パッケージ | 役割 |
|-----------|-----------|------|
| Human Mode | `adapter/human` | ターミナルUI。メニュー操作、ASCII描画、早送り機能 |
| AI Mode | `adapter/ai` | JSON Lines プロトコル。LLM等の外部プログラムと stdin/stdout で通信 |
| Batch Mode | `adapter/batch` | ヘッドレス並列バッチ実行。N ゲームの統計収集とレポート生成 |

### Balance Dashboard（`sim/balance/`）

Batch Mode の上位レイヤー。4段階フロー:
1. **ベースライン実行**: 指定シナリオで N ゲームのバッチ実行
2. **壊れるサイン表示**: B01〜B11 のアラート検出と表示
3. **スイープ提案**: D002ルールに基づくパラメータ調整候補の自動生成
4. **比較**: パラメータスイープ実行と結果比較

## 壊れるサインメトリクス（B01〜B11）

「面白さの証明ではなく、壊れている兆候の検出」を目的とする11種のメトリクス:

| ID | 名称 | 検出対象 |
|----|------|---------|
| B01 | TicksBeforeFirstWave | 最初の侵入波までの準備猶予ティック数 |
| B02 | ActionsBeforeFirstWave | 最初の侵入波までの操作可能回数 |
| B03 | TerrainBlockRate | 地形制約による配置不可率 |
| B04 | ZeroBuildableRate | 建設可能位置がゼロのゲーム割合 |
| B05 | WaveOverlapRate | 侵入波が建設中に重なる割合 |
| B06 | StompRate | 常に圧勝してしまう割合 |
| B07 | EarlyWipeRate | 序盤で壊滅する割合 |
| B08 | PerfectionRate | クリア時に全部屋MAXの割合 |
| B09 | AvgRoomLevelRatio | 平均部屋レベル比率 |
| B10 | LayoutEntropy | レイアウト多様性 |
| B11 | ResourceSurplusRate | リソース余剰率 |

## 設計上の重要な判断

- **D002**: コアゲーム体験は「常に不完全な状態での判断」。3原則（不完全性の強制、時間圧力、トレードオフの連続）
- **D009**: 気のシステムは物理層（ChiFlowEngine）と経済層（ChiPool）の二層構造
- **D012**: 1ティック12ステップの更新順序は不変。変更すると決定論的再現が壊れる
- **D013**: PlayerAction は Step() の引数として毎ティック注入。リプレイ・AI・人間が同一インターフェース

詳細は [DECISIONS.md](./DECISIONS.md) を参照。

## 次のステップ: chaosseed-game

`game/` ディレクトリに Ebitengine ベースの GUI クライアントを構築する。

### 構成

```
game → core  (sim とは独立)
```

### 留意事項

- core の `SimulationEngine` は GUI からも同じインターフェースで利用可能（`Step(actions)` を毎フレーム呼ぶ）
- `GameSnapshot` による読み取り専用のゲーム状態取得は描画層に最適
- 組み込みシナリオ JSON は core の `scenario` パッケージから読み込み可能
- sim のアダプター層の設計（ActionProvider パターン）は game でも参考になる

### 未解決の将来課題

- 罠の盗賊回避率（盗賊の SPD による罠回避メカニクス）
- 侵入者AIの高度化（複数ステップ先読み、仙獣回避）
- standard シナリオ向けのスイープ値チューニング
