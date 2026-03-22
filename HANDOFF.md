# HANDOFF.md — chaosseed プロジェクト引き継ぎ文書

## プロジェクト概要

カオスシード（風水回廊記）インスパイアのダンジョン経営シミュレーション。
Go モノレポ構成で、core（ゲームロジックライブラリ）、sim（CLIシミュレーター）、game（Ebitengine GUIクライアント）を実装済み。

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

### game（v1.0.0）

Phase 0〜4 で Ebitengine GUI クライアントを完成:

| Phase | 内容 |
|-------|------|
| 0 | Ebitengine 習作 + プロジェクト初期化（カラーパレット、仮タイルセット、マップ描画、マウスホバー） |
| 1 | GameController + ティック進行（core 接続、手動/早送り/一時停止、Snapshot→描画変換、TopBar） |
| 2 | 操作システム（ActionMode ステートマシン、ActionBar、DigRoom/DigCorridor/SummonBeast/UpgradeRoom フロー） |
| 3 | シーン管理 + UI（SceneManager、Title/Select/InGame/Result シーン、InfoPanel） |
| 4 | セーブ/ロード + 仕上げ（チェックポイント保存/復元、ゲーム設定永続化、戦闘/侵入の視覚フィードバック） |

## アーキテクチャ

### sim

```
chaosseed-sim CLI
    ├── --human  → adapter/human  → GameServer → core.SimulationEngine
    ├── --ai     → adapter/ai     → GameServer → core.SimulationEngine
    ├── --batch  → adapter/batch  → GameServer → core.SimulationEngine
    ├── --balance→ balance        → adapter/batch → GameServer
    └── --replay → server.Replay  → GameServer → core.SimulationEngine
```

### game

```
chaosseed-game (Ebitengine)
    ├── SceneManager ── Title → Select → InGame → Result
    │                          └── Load → InGame
    ├── InGame
    │   ├── GameController → core.SimulationEngine
    │   ├── InputStateMachine → ActionMode (Normal/DigRoom/DigCorridor/Summon/Upgrade)
    │   ├── MapView + EntityRenderer → 描画
    │   ├── TopBar + ActionBar + InfoPanel → UI
    │   ├── FeedbackOverlay + BattleFeedbackOverlay → 視覚フィードバック
    │   └── Save/Load → ~/.chaosforge/saves/
    └── asset → PlaceholderProvider (TilesetProvider)
```

**依存方向**: `game → core`（sim とは独立）

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

### GameController（`game/controller/`）

core の `SimulationEngine` を GUI 向けにラップ:
- `Snapshot()` で読み取り専用の GameSnapshot を取得（描画層に最適）
- `AddAction()` で PlayerAction をキューに積み、`AdvanceTick()` で実行
- ティック進行モード: Manual / FastForward / Paused
- `SaveCheckpoint()` / `LoadCheckpoint()` でセーブ/ロード

### TilesetProvider（`game/asset/`）

仮アセットシステム。`TilesetProvider` インターフェースで描画とアセットを分離:
- `PlaceholderProvider`: 全 CellType 分の 32x32 色付き矩形を動的生成
- 将来、ドット絵アセットに差し替える場合は `TilesetProvider` の別実装を提供するだけで対応可能

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
- **D019**: SummonBeast は種族選択ではなく Element 選択方式。core が species を自動決定
- **D020**: Scene インターフェースの Draw は `image.Image` 引数。headless テスト可能

詳細は [DECISIONS.md](./DECISIONS.md) を参照。

## 未解決の将来課題

- ドット絵アセットの作成（現在は PlaceholderProvider の色付き矩形）
- BGM/SE の追加
- 龍脈の可視化（気の流れをパーティクルやアニメーションで表現）
- 罠の盗賊回避率（盗賊の SPD による罠回避メカニクス）
- 侵入者AIの高度化（複数ステップ先読み、仙獣回避）
- standard シナリオ向けのスイープ値チューニング
- `game/testdata/tutorial.json` の embed 重複解消（core の `LoadBuiltinScenario` への統一）
- `BuildRoomRenderMap` の毎フレーム再構築の最適化（パフォーマンス問題が顕在化した場合）
- ebiten 依存パッケージのテストが headless 環境で実行不可（GLFW init panic）
