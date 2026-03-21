# chaosseed-sim PRD v1.0.0 — CLIシミュレーター

> chaosseed-core v1.0.0 のコアメカニクスエンジンを活用し、
> ゲームをプレイ・評価・調整するためのCLIツール。

---

## 1. プロジェクトの目的

### 1.1 なぜ sim が必要か

chaosseed-core は純粋なロジックライブラリであり、「ゲーム体験」を提供しない。
sim は **3つの異なる利用者** にゲームへの窓を提供する：

| 利用者 | モード | 目的 |
|--------|--------|------|
| ぽんぽこ（人間） | Human Mode | D002の5つの面白さが本当に機能しているかを体感する |
| LLM（Claude Code等） | AI Mode | ゲーム状態を構造化データで受け取り、戦略を評価・改善する |
| 統計エンジン | Batch Mode | 数千ゲームを自動実行し、バランスパラメータを定量評価する |

### 1.2 成功基準

1. **Human Mode**: ぽんぽこがチュートリアルシナリオを最初から最後までプレイでき、D002の「妥協」「中断」を体感できる
2. **AI Mode**: Claude Code が JSON I/O でゲームを1ゲーム完走でき、戦略の良し悪しをスコアで評価できる
3. **Batch Mode**: 1,000ゲームのバッチ実行が5分以内に完了し、BreakageReport（壊れるサイン検出）が出力される
4. **バランスダッシュボード**: パラメータ変更→再シミュレーション→比較が1コマンドで実行できる

---

## 2. モノレポ構成

### 2.1 ディレクトリ構造

```
chaosseed/
├── go.work                  # Go Workspace定義
├── core/
│   ├── go.mod               # module github.com/ponpoko/chaosseed/core
│   ├── types/
│   ├── world/
│   ├── fengshui/
│   ├── senju/
│   ├── invasion/
│   ├── economy/
│   ├── scenario/
│   ├── simulation/
│   ├── cmd/caveviz/
│   ├── testutil/
│   └── docs/PRD.md
├── sim/
│   ├── go.mod               # module github.com/ponpoko/chaosseed/sim
│   ├── server/              # Game Server（コアループ管理）
│   ├── adapter/
│   │   ├── human/           # Human Mode（対話メニュー）
│   │   ├── ai/              # AI Mode（JSON I/O）
│   │   └── batch/           # Batch Mode（ヘッドレス統計）
│   ├── render/              # ASCII描画（coreのcavevizを拡張）
│   ├── metrics/             # D002メトリクス収集・レポート
│   ├── balance/             # バランス調整ダッシュボード
│   ├── cmd/chaosseed-sim/   # メインエントリポイント
│   └── docs/PRD.md          # この文書
├── game/                    # （将来：Ebitengine GUI）
│   └── go.mod
├── DECISIONS.md             # プロジェクト横断の設計判断
├── HANDOFF.md               # プロジェクト横断の引き継ぎ文書
└── LESSONS.md               # プロジェクト横断の知見
```

### 2.2 go.work

```go
go 1.22

use (
    ./core
    ./sim
)
```

### 2.3 sim の go.mod

```go
module github.com/ponpoko/chaosseed/sim

go 1.22

require github.com/ponpoko/chaosseed/core v0.0.0
```

go.work により、ローカル開発時は core の変更が即座に sim に反映される。
リリース時は core にタグを打ち、sim の go.mod でバージョン指定する。

---

## 3. アーキテクチャ

### 3.1 レイヤー構成

```
┌─────────────────────────────────────────────────┐
│  CLI Entry Point (cmd/chaosseed-sim)            │
│  --human / --ai / --batch / --balance           │
├─────────────────────────────────────────────────┤
│  Adapter Layer                                  │
│  ┌──────────┐ ┌──────────┐ ┌──────────────────┐ │
│  │  Human   │ │    AI    │ │     Batch        │ │
│  │  Menu UI │ │  JSON IO │ │  Headless Stats  │ │
│  └────┬─────┘ └────┬─────┘ └────┬─────────────┘ │
│       │            │            │               │
├───────┴────────────┴────────────┴───────────────┤
│  Game Server (server/)                          │
│  - セッション管理                                │
│  - ActionProvider インターフェース実装            │
│  - GameSnapshot → 各アダプターへの状態配信        │
│  - チェックポイント / リプレイ管理                │
├─────────────────────────────────────────────────┤
│  Metrics Collector (metrics/)                   │
│  - 毎ティックのメトリクス収集                     │
│  - D002検証メトリクス                            │
│  - ゲーム終了時のサマリー生成                     │
├─────────────────────────────────────────────────┤
│  chaosseed-core (SimulationRunner)              │
│  - Step(actions) / Run(actionProvider)           │
│  - GameSnapshot / Checkpoint / Replay           │
└─────────────────────────────────────────────────┘
```

### 3.2 Game Server の責務

Game Server は **core の SimulationRunner を薄くラップ** するレイヤー。
core の D013（actionProvider パターン）がすでに抽象化を提供しているため、
Game Server は以下に集中する：

1. **セッションライフサイクル**: シナリオ読み込み → ゲーム開始 → ティック進行 → ゲーム終了
2. **ActionProvider ブリッジ**: アダプターからのアクションを core の ActionProvider 型に変換
3. **Snapshot 配信**: 毎ティックの GameSnapshot をアダプターと Metrics Collector に配信
4. **チェックポイント/リプレイ**: 保存・復元・再生の操作を core に委譲

```go
// server/server.go
type GameServer struct {
    runner    *simulation.SimulationRunner
    metrics   *metrics.Collector
    snapshot  *simulation.GameSnapshot  // 最新の状態
}

// ActionProvider はアダプターが実装するインターフェース
type ActionProvider interface {
    // ProvideActions はゲーム状態を受け取り、プレイヤーアクションを返す
    ProvideActions(snapshot *simulation.GameSnapshot) ([]simulation.PlayerAction, error)
    // OnTickComplete はティック完了後の通知（描画・ログ等）
    OnTickComplete(snapshot *simulation.GameSnapshot, tickMetrics *metrics.TickMetrics)
    // OnGameEnd はゲーム終了時の通知
    OnGameEnd(result *simulation.RunResult, summary *metrics.GameSummary)
}
```

### 3.3 アダプター詳細

#### Human Mode（adapter/human/）

```
┌────────────────────────────────────────┐
│  === Tick 42 ===                       │
│                                        │
│  [ASCII Map + 仙獣 + 侵入者表示]       │
│                                        │
│  気: 150/500  CoreHP: 80/100           │
│  風水スコア: 72  部屋: 5/10            │
│                                        │
│  ⚔ 侵入者: 3体（部屋2に2体、通路に1体）│
│  🐉 仙獣: 麒麟(Lv3/Guard) 朱雀(Lv2)   │
│                                        │
│  === アクション ===                     │
│  1. 部屋を掘る                          │
│  2. 通路を掘る                          │
│  3. 仙獣を召喚する                      │
│  4. 部屋をアップグレードする             │
│  5. 何もしない（1ティック進める）        │
│  6. 早送り（Nティック）                  │
│  ---                                   │
│  s. セーブ  l. ロード  r. リプレイ保存  │
│  q. 終了                                │
│                                        │
│  > _                                   │
└────────────────────────────────────────┘
```

- 数字選択 → サブメニュー（例：「1. 部屋を掘る」→ 座標選択 → 属性選択）
- 無効なアクションはバリデーション後にエラー表示、再入力
- 早送り中は描画をスキップ（完了後に結果表示）

#### AI Mode（adapter/ai/）

stdin/stdout の JSON Lines プロトコル。1行1メッセージ。

```jsonl
← {"type":"state","tick":42,"snapshot":{...},"valid_actions":[...]}
→ {"type":"action","actions":[{"kind":"dig_room","pos":[5,3],"element":"fire"}]}
← {"type":"state","tick":43,"snapshot":{...},"valid_actions":[...]}
→ {"type":"action","actions":[{"kind":"wait"}]}
...
← {"type":"game_end","result":"victory","summary":{...},"metrics":{...}}
```

設計ポイント：
- **valid_actions フィールド**: 現在実行可能なアクション一覧をJSON配列で提供。LLMが無効なアクションを試行するコストを削減
- **snapshot は core の GameSnapshot をそのままJSON化**: LLMが全ゲーム状態にアクセス可能
- **metrics フィールド**: D002メトリクスをゲーム中・ゲーム後に提供
- **エラーハンドリング**: 不正なJSON → エラーメッセージ返却 → 再入力待ち

#### Batch Mode（adapter/batch/）

```bash
# 基本実行
chaosseed-sim --batch --scenario tutorial.json --games 1000 --output results.json

# AI付きバッチ（SimpleAI / RandomAI / カスタム）
chaosseed-sim --batch --scenario tutorial.json --games 1000 --ai simple --output results.json

# パラメータスイープ
chaosseed-sim --batch --scenario tutorial.json --games 1000 \
  --sweep "economy.supply_multiplier=0.5,1.0,1.5,2.0" \
  --output sweep_results.json
```

出力フォーマット（JSON）：
```json
{
  "config": { "scenario": "tutorial.json", "games": 1000, "ai": "simple" },
  "summary": {
    "win_rate": 0.62,
    "avg_ticks": 187,
    "avg_rooms_built": 6.3,
    "avg_core_hp_remaining": 45.2
  },
  "breakage_report": {
    "alerts": [
      {
        "metric_id": "B06",
        "broken_sign": "常に圧勝",
        "value": 0.35,
        "threshold": 0.30,
        "direction": "侵入者の強さを上げる / リソース供給を下げる"
      }
    ],
    "clean": ["B01","B02","B03","B04","B05","B07","B08","B09","B10","B11"]
  },
  "raw_metrics": {
    "B01_ticks_before_first_wave": { "mean": 32.1, "p10": 20, "p90": 45 },
    "B02_actions_before_first_wave": { "mean": 8.2, "p10": 4, "p90": 12 },
    "B03_terrain_block_rate": { "mean": 0.18 },
    "B04_zero_buildable_rate": 0.02,
    "B05_wave_overlap_rate": 0.58,
    "B06_stomp_rate": 0.35,
    "B07_early_wipe_rate": 0.05,
    "B08_perfection_rate": 0.00,
    "B09_avg_room_level_ratio": 0.42,
    "B10_layout_entropy": 0.81,
    "B11_resource_surplus_rate": 0.15
  }
}
```

### 3.4 Metrics Collector（metrics/）— 壊れるサイン検出器

全モード共通で動作する。Game Server が毎ティック呼び出す。

**設計原則**: メトリクスは「面白さの証明」ではなく「壊れるサインの検出」に使う。
面白さはぽんぽこが Human Mode で体感して判断するもの。メトリクスの仕事は
「この面白さが壊れていないか？」を自動監視し、壊れていたら調整の方向を示すこと。

D002 のパラメータ調整判断基準テーブルがそのまま測定仕様になる。

#### 壊れるサイン検出テーブル

| 面白さ | 壊れるサイン | 検出メトリクス | 検出条件（アラーム） | 調整の方向 |
|---|---|---|---|---|
| 計画する | 計画前に敵が来る | B01: TicksBeforeFirstWave — 初波到達ティック | < シナリオ定義の最低猶予ティック | 初波までの猶予を伸ばす |
| 計画する | 計画前に敵が来る | B02: ActionsBeforeFirstWave — 初波到達前のPlayerAction数 | < 3（最低限の建設すらできない） | 初波までの猶予を伸ばす |
| 妥協する | 常に理想配置が可能 | B03: TerrainBlockRate — 建設試行のうち地形制約(HardRock/Water)で阻止された割合 | < 0.05（制約がほぼ機能していない） | 地形制約密度を上げる |
| 妥協する | 常に詰む | B04: ZeroBuildableRate — 開始時に建設可能セルが極端に少ないゲームの割合 | > 0.10（10%以上のseedで詰み寸前） | 地形制約密度を下げる / ValidateTerrain閾値調整 |
| 中断される | 構築と防衛が完全に交互 | B05: WaveOverlapRate — 直前Nティック内に建設アクションがあった状態で侵入波が到達した割合 | < 0.30（波が常に建設の合間に来る＝中断にならない） | 侵入波を構築中に重ねる |
| 噛み合う | 常に圧勝 | B06: StompRate — CoreHP 80%以上残して勝利の割合 | > 0.30（防衛が緩すぎる） | 侵入者の強さを上げる / リソース供給を下げる |
| 噛み合う | 常に惨敗 | B07: EarlyWipeRate — 全ティックの50%以内にCoreHP=0の割合 | > 0.20（準備が間に合わず壊滅） | 侵入者の強さを下げる / 初波猶予を伸ばす |
| 完璧が来ない | クリア前に全MAX到達 | B08: PerfectionRate — クリア時に全部屋がMaxLvに到達しているゲームの割合 | > 0.05（完璧なダンジョンが完成してしまう） | リソース総量を絞る / クリア条件を早める |
| 完璧が来ない | クリア前に全MAX到達 | B09: AvgRoomLevelRatio — クリア時の全部屋Lv平均 / MaxLv | > 0.80（ほぼ完成している） | リソース総量を絞る / クリア条件を早める |

加えて、D002の4つのアンチパターンを直接検出する：

| アンチパターン | 検出メトリクス | 検出条件 |
|---|---|---|
| 完璧なダンジョンが完成する → ゲームの死 | B08 + B09 | B08 > 0.05 または B09 > 0.80 |
| 最適解が1つに収束 → リプレイ性の喪失 | B10: LayoutEntropy — 異なるseed間での最終部屋配置のエントロピー | < 閾値（配置パターンが収束） |
| 侵入者が弱すぎ → 判断の必要性が消滅 | B06 | B06 > 0.30 |
| リソースが潤沢 → トレードオフの消滅 | B11: ResourceSurplusRate — ゲーム後半でChiPoolが最大値の80%以上を維持したティック割合 | > 0.50（常に余裕がある） |

#### 命名規則

メトリクスIDは `B` プレフィックス（**B**roken sign）。面白さの「証明」ではなく「壊」れるサインの検出であることを名前で明示する。

#### BreakageReport

```go
// metrics/breakage.go
type BreakageAlert struct {
    MetricID   string  // "B01" 〜 "B11"
    BrokenSign string  // 「計画前に敵が来る」等、D002テーブルの壊れるサイン
    Value      float64
    Threshold  float64
    Direction  string  // 「初波までの猶予を伸ばす」等、D002テーブルの調整の方向
}

type BreakageReport struct {
    Alerts  []BreakageAlert  // 検出条件に引っかかったもののみ
    Clean   []string         // 問題なしのメトリクスID一覧
}

func (c *Collector) DetectBreakage() *BreakageReport
```

**アラートが0件 = 壊れていない。面白いかどうかは人間が判断する。**

---

## 4. バランス調整ダッシュボード（balance/）

```bash
chaosseed-sim --balance --scenario tutorial.json --games 500
```

実行フロー：
1. ベースラインのバッチ実行（500ゲーム）
2. 壊れるサイン検出（BreakageReport）
3. アラートのあるメトリクスに対して、D002テーブルの「調整の方向」に基づくパラメータスイープを自動提案
4. スイープ実行 → アラート解消の確認

```
=== Balance Dashboard: tutorial.json ===

Baseline (500 games):
  Win Rate: 62.0%   Avg Ticks: 187

Breakage Report:
  ✅ B01 TicksBeforeFirstWave       32.1   (threshold: > min_grace)
  ✅ B02 ActionsBeforeFirstWave      8.2   (threshold: >= 3)
  ✅ B03 TerrainBlockRate            0.18  (threshold: >= 0.05)
  ✅ B04 ZeroBuildableRate           0.02  (threshold: <= 0.10)
  ✅ B05 WaveOverlapRate             0.58  (threshold: >= 0.30)
  🔴 B06 StompRate                   0.35  (threshold: <= 0.30) ← 常に圧勝
  ✅ B07 EarlyWipeRate               0.05  (threshold: <= 0.20)
  ✅ B08 PerfectionRate              0.00  (threshold: <= 0.05)
  ✅ B09 AvgRoomLevelRatio           0.42  (threshold: <= 0.80)
  ✅ B10 LayoutEntropy               0.81  (threshold: >= min)
  ✅ B11 ResourceSurplusRate         0.15  (threshold: <= 0.50)

1 alert detected.

B06: 常に圧勝
  調整の方向: 侵入者の強さを上げる / リソース供給を下げる
  Suggested sweep: invasion.base_attack_power = [8, 10, 12, 15, 18]

Run sweep? [y/N]: y

Sweep Results:
  base_attack_power=8  → B06=0.45 🔴  WinRate=0.78
  base_attack_power=10 → B06=0.35 🔴  WinRate=0.62  ← baseline
  base_attack_power=12 → B06=0.22 ✅  WinRate=0.55
  base_attack_power=15 → B06=0.11 ✅  WinRate=0.41
  base_attack_power=18 → B06=0.04 ✅  WinRate=0.28  ← but B07=0.25 🔴 新たな壊れ

Best: base_attack_power=12 (B06 resolved, no new alerts)
Apply to scenario? [y/N]:
```

**ダッシュボードの判断基準**: アラートが0件になるパラメータを探す。
新たなアラートが発生するパラメータは除外する。
複数パラメータの同時スイープが必要な場合は、1つずつ順番に調整する。

---

## 5. core 残存課題の修正

sim の開発初期フェーズ（Phase 1）でモノレポ移行と同時に修正する。

| 課題 | 修正方針 | 影響 |
|------|---------|------|
| MaxRooms制約未チェック | `validateDigRoom` に `GameConstraints.MaxRooms` チェック追加 | core/simulation/action.go |
| SimpleAIのコリドー戦略なし | SimpleAI に「新部屋建設後、隣接部屋への通路を掘る」ロジック追加 | core/simulation/ai_simple.go |
| CalcTickSupplyのcaveScore正規化 | CaveTotal を MaxPossibleChi で正規化（0.0〜1.0にクランプ） | core/economy/supply.go |

修正後、core に v1.1.0 タグを打つ。

---

## 6. フェーズ分割

### Phase 0: モノレポ移行とプロジェクト初期化

**ゴール**: go.work ベースのモノレポが動作し、sim のスケルトンが `go build` できる

タスク:
1. chaosseed/ ルートディレクトリ作成、go.work 配置
2. 既存 core リポジトリを core/ に移動（git history 保持）
3. sim/ ディレクトリ作成、go.mod 初期化
4. DECISIONS.md, HANDOFF.md, LESSONS.md をルートに移動
5. sim/cmd/chaosseed-sim/main.go のスケルトン作成
6. Makefile（build, test, lint, vet）
7. `go work sync` → `go build ./...` → `go test ./...` 全パス

### Phase 1: Game Server + core修正

**ゴール**: Game Server が core の SimulationRunner を駆動でき、ActionProvider インターフェースが定義されている

タスク:
1. core 残存課題3件の修正（MaxRooms, コリドー戦略, caveScore正規化）
2. core v1.1.0 タグ
3. server/server.go: GameServer 構造体、NewGameServer, Start, Step
4. server/provider.go: ActionProvider インターフェース定義
5. server/session.go: セッションライフサイクル（Load → Play → End）
6. Metrics Collector のスケルトン（metrics/collector.go）
7. 統合テスト: GameServer + core の SimpleAI で1ゲーム完走

### Phase 2: Human Mode（対話メニュー）

**ゴール**: ぽんぽこがターミナルでチュートリアルシナリオをプレイできる

タスク:
1. render/ascii.go: core の RenderFullStatus を拡張（色付き、見やすいレイアウト）
2. adapter/human/menu.go: メインメニュー（6アクション + ユーティリティ）
3. adapter/human/submenu.go: サブメニュー（座標選択、属性選択、ティック数入力）
4. adapter/human/provider.go: ActionProvider 実装
5. adapter/human/display.go: ティック結果の表示（戦闘ログ、経済変化、イベント通知）
6. cmd/chaosseed-sim/main.go: `--human` フラグ → Human Mode 起動
7. チェックポイント save/load の Human Mode 統合
8. E2Eテスト: チュートリアルシナリオのスクリプト実行

### Phase 3: AI Mode（JSON I/O）

**ゴール**: stdin/stdout の JSON Lines で1ゲーム完走でき、Claude Code から操作できる

タスク:
1. adapter/ai/protocol.go: メッセージ型定義（StateMessage, ActionMessage, ErrorMessage, GameEndMessage）
2. adapter/ai/provider.go: ActionProvider 実装（stdin読み取り → パース → バリデーション）
3. adapter/ai/serializer.go: GameSnapshot → JSON 変換（valid_actions 生成含む）
4. cmd/chaosseed-sim/main.go: `--ai` フラグ → AI Mode 起動
5. エラーハンドリング: 不正JSON、無効アクション、タイムアウト
6. E2Eテスト: パイプ経由でのゲーム完走
7. ドキュメント: AI Mode プロトコル仕様書（LLMのプロンプトに含められる形式）

### Phase 4: Batch Mode + D002メトリクス

**ゴール**: 1,000ゲームのバッチ実行が完了し、BreakageReport（壊れるサイン検出）が出力される

タスク:
1. metrics/collector.go: 壊れるサイン検出メトリクス全11件（B01〜B11）の実装
2. metrics/breakage.go: BreakageReport、アラート判定
3. metrics/report.go: JSON/CSV レポート生成
4. adapter/batch/runner.go: 並列バッチ実行（goroutine プール）
5. adapter/batch/sweep.go: パラメータスイープ実行
6. cmd/chaosseed-sim/main.go: `--batch` フラグ + オプション
7. パフォーマンステスト: 1,000ゲーム × SimpleAI が5分以内
8. D002検証: Phase 7-H の検証をBreakageReportで再現（全アラート0件を確認）

### Phase 5: バランス調整ダッシュボード

**ゴール**: 1コマンドでベースライン→健全性チェック→スイープ提案→比較が完了する

タスク:
1. balance/dashboard.go: ダッシュボード本体（BreakageReport表示）
2. balance/suggest.go: アラートからD002テーブルの「調整の方向」に基づくパラメータスイープの自動提案
3. balance/compare.go: ベースライン vs スイープ結果の比較表生成（新たなアラート発生の検出含む）
4. balance/apply.go: パラメータ変更のシナリオJSONへの反映
5. cmd/chaosseed-sim/main.go: `--balance` フラグ
6. E2Eテスト: チュートリアルシナリオでのダッシュボード実行

### Phase 6: 統合検証 + リリース

**ゴール**: 全モードの統合テスト完了、v1.0.0 リリース

タスク:
1. 全モードの統合テスト
2. Human Mode でぽんぽこが実際にプレイ → D002フィードバック
3. AI Mode で Claude Code が1ゲーム完走 → 戦略評価テスト
4. Batch Mode で 1,000ゲーム → BreakageReport アラート0件
5. ダッシュボードでアラートがある場合はパラメータ調整 → 再検証 → アラート0件
6. HANDOFF.md 更新（sim 完了分を追記）
7. DECISIONS.md 更新（sim で生まれた設計判断を追記）
8. v1.0.0 タグ

---

## 7. CLI インターフェース仕様

```bash
# Human Mode（デフォルト）
chaosseed-sim --human --scenario path/to/scenario.json
chaosseed-sim --human --scenario tutorial  # 組み込みシナリオ

# AI Mode
chaosseed-sim --ai --scenario path/to/scenario.json
chaosseed-sim --ai --scenario tutorial --timeout 30s

# Batch Mode
chaosseed-sim --batch --scenario tutorial --games 1000 --ai simple --output results.json
chaosseed-sim --batch --scenario tutorial --games 1000 --ai random --format csv --output results.csv
chaosseed-sim --batch --scenario tutorial --games 1000 \
  --sweep "economy.supply_multiplier=0.5,1.0,2.0" --output sweep.json

# Balance Dashboard
chaosseed-sim --balance --scenario tutorial --games 500

# ユーティリティ
chaosseed-sim --replay path/to/replay.json          # リプレイ再生
chaosseed-sim --checkpoint path/to/checkpoint.json   # チェックポイントから再開
chaosseed-sim --scenarios                            # 利用可能シナリオ一覧
chaosseed-sim --version
```

---

## 8. 技術方針

### 8.1 依存関係

- **外部依存ゼロ** を維持（core と同様）
- ターミナル制御: 標準ライブラリの `os`, `bufio`, `fmt` のみ
- ANSIエスケープコードによる色付き出力（ライブラリ不使用）
- JSON処理: `encoding/json`（標準ライブラリ）

### 8.2 並列処理（Batch Mode）

- `runtime.NumCPU()` ベースの goroutine プール
- 各ゲームは独立（共有状態なし）→ 完全並列化可能
- `sync.WaitGroup` + チャネルでの結果集約

### 8.3 テスト方針

- テーブル駆動テスト、カバレッジ80%目標（core と同水準）
- Human Mode: スクリプト化された入力でE2Eテスト
- AI Mode: パイプ経由のE2Eテスト
- Batch Mode: 少数ゲーム（10ゲーム）での統合テスト + メトリクス検証

### 8.4 Ralph Loop 適合

- 各フェーズのタスクは「1イテレーション1タスク」の粒度
- Phase末尾で DECISIONS.md 棚卸し → PHASE_COMPLETE → tasks.md 空で停止
- sim 固有の PROMPT.md を作成（core の PROMPT.md をベースに sim の文脈を追加）

---

## 9. 将来の拡張（v1.0.0 スコープ外）

| 機能 | 備考 |
|------|------|
| MCP サーバーモード | AI Mode の上位互換。Claude Desktop から直接接続 |
| Web UI ダッシュボード | バランス調整結果のブラウザ表示 |
| シナリオエディタ | CLI でのシナリオ作成・編集 |
| マルチプレイヤーAI対戦 | 異なるAI戦略の対戦・ランキング |
| core v2 対応 | 罠の盗賊回避率、侵入者AI高度化 等の新機能対応 |

---

## 10. D002 との対応関係

このPRDの全機能は D002「常に不完全な状態での判断」の維持と調整に収束する。

**面白さは人間が体感して判断する。メトリクスは壊れるサインを検出する。**

```
Human Mode   → D002を体感する    → 「面白い」はここでしか判断できない
Batch Mode   → 壊れるサインを検出  → B01〜B11アラート
Dashboard    → 壊れを直す        → パラメータスイープ → アラート0件にする
AI Mode      → 戦略の多様性を検証  → LLMが攻略 → B10(LayoutEntropy)に寄与
```

ダッシュボードが「アラート0件」を達成した状態で、ぽんぽこがHuman Modeでプレイし、
「面白い」と感じればバランス調整は完了。「壊れてないのに面白くない」場合は、
D002テーブル自体の見直し（壊れるサインの定義追加/変更）に戻る。

**simの究極的な成果物は「D002が壊れていないことの保証」と「壊れたときに直すためのツールチェーン」である。**