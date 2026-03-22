# PROMPT.md — chaosseed-game Ralph Loop

あなたは chaosseed プロジェクトの開発者です。
カオスシード（風水回廊記）にインスパイアされたダンジョン経営シミュレーションの Ebitengine GUI クライアントを Go で構築しています。

## あなたの役割

1. `tasks.md` を読み、最初の未完了タスク（`- [ ]`）を **1つだけ** 選ぶ
2. `CLAUDE.md` のコーディング規約に従って実装する
3. テストを書き、`go test ./...` が全パスすることを確認する
4. 完了したタスクのチェックボックスを `- [x]` に更新する
5. 実装中に得た知見があれば `LESSONS.md` に追記する（なければ作成）

## 重要なルール

- **1回の実行で1タスクだけ** を完了させる。欲張らない
- タスクの順序を守る。依存関係が上から下に流れている
- テストが通らない状態でチェックを入れない
- 既存のテストを壊さない。`go test ./...` で全パスを確認
- わからないこと・設計判断が必要な場合は `DECISIONS.md` に記録して次に進む
- PRD（`./game/docs/PRD.md`）を game のリファレンスとして参照する
- PRD（`./sim/docs/PRD.md`）を sim のリファレンスとして参照する
- `HANDOFF.md` で core/sim の完了状態と game への引き継ぎ事項を確認する
- タスクが1つ終わったらgitのコミットを行い、これを区切りとしイテレーションは終える

## プロジェクト構造

```
seed/                     # モノレポルート
├── go.work               # Go Workspace（core, sim, game）
├── Makefile              # ルート Makefile（全モジュール統合）
├── CLAUDE.md             # コーディング規約（必読）
├── PROMPT.md             # このファイル
├── tasks.md              # タスクリスト（チェックボックス管理）
├── LESSONS.md            # 学んだこと（自動追記）
├── DECISIONS.md          # 設計判断の記録
├── HANDOFF.md            # プロジェクト引き継ぎ文書
├── core/                 # コアメカニクスエンジン（v1.1.0 完了）
│   ├── types/            # 共有型定義（Element, Direction, RNG 等）
│   ├── world/            # マップシステム（Cave, Room, Corridor）
│   ├── fengshui/         # 風水システム（ChiFlowEngine, DragonVein）
│   ├── senju/            # 仙獣システム（Beast, Growth, Behavior, Evolution）
│   ├── invasion/         # 侵入システム（InvasionEngine, WaveSchedule）
│   ├── economy/          # 経済システム（ChiPool, SupplyCalculator）
│   ├── scenario/         # シナリオシステム（EventEngine, 勝敗条件）
│   ├── simulation/       # 統合シミュレーション（SimulationEngine, GameSnapshot）
│   └── testutil/         # テストヘルパー
├── sim/                  # CLI シミュレーター（v1.0.0 完了）
│   ├── server/           # Game Server（ActionProvider, Checkpoint, Replay）
│   ├── adapter/          # Human / AI / Batch アダプター
│   ├── metrics/          # 壊れるサイン検出（B01〜B11）
│   ├── balance/          # バランス調整ダッシュボード
│   └── cmd/chaosseed-sim/
└── game/                 # Ebitengine GUI クライアント（現在開発中）
    ├── docs/PRD.md       # game PRD
    ├── controller/       # ゲーム制御（ティック管理、アクション組み立て）
    ├── scene/            # シーン管理（Title, Select, InGame, Result）
    ├── view/             # 描画（MapView, Entity, UI, Tooltip）
    ├── input/            # 入力処理（Mouse, Keyboard, ActionMode）
    ├── asset/            # アセット（Palette, TilesetProvider, Placeholder）
    └── save/             # セーブ/ロード
```

## 現在のフェーズ

tasks.md の先頭コメントで現在のフェーズを確認すること。
フェーズ内の全タスクが完了したら、最後のタスクとして次フェーズの tasks.md を生成する指示がある場合がある。

## フェーズ完了プロトコル

フェーズ最終タスクに到達したら:

1. `DECISIONS.md` の OPEN エントリを確認する:
   - このフェーズで解決されたものは RESOLVED に更新し、解決確認の内容を追記する
   - 次フェーズで対応が必要なものを特定し、PHASE_COMPLETE の申し送りに記載する
2. `PHASE_COMPLETE` ファイルをプロジェクトルートに作成する（既存なら上書き）
3. 以下を記載する:
   - 実装した内容の要約
   - 未解決の課題や技術的負債
   - 次フェーズへの申し送り事項（DECISIONS.md の未解決課題を含む）
   - LESSONS.md から特に重要な知見
4. **tasks.md には新しい未完了タスクを絶対に追加しない**
5. 次フェーズのタスクドラフトを `tasks_phaseN_draft.md` として別ファイルに生成してよい
6. 次フェーズのタスク作成は人間がチャットで行う。CLIが次フェーズのタスクを tasks.md に書いてはならない

## コンテキスト

- Go モノレポ構成: core（ロジック）→ sim（CLI）→ game（GUI）
- core と sim は完成済み。game は core を直接 import して利用する（sim には依存しない）
- 決定論的設計: 乱数は必ず RNG インターフェース経由
- core の `SimulationEngine.Step(actions)` と `GameSnapshot` が game の主要なインターフェース
- 五行（木火土金水）の相生・相克がゲームの根幹
- GUI の「見た目の正しさ」は go test で検証できない。完了条件は「ロジックテスト + コンパイル成功 + vet クリーン」に限定する
