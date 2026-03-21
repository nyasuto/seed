# tasks_phase3_5_draft.md — Phase 3.5: 仙獣行動AI（senju/）

<!-- Phase: 3.5 (senju/) — 仙獣行動AIシステム -->

## Phase 3.5-A: 行動パターン定義（senju/）

- [ ] `senju/behavior.go`: BehaviorType 型（Guard/Patrol/Chase/Flee）。Behavior インターフェース — DecideAction(beast, context BehaviorContext) Action。BehaviorContext 構造体（Beast の現在位置、部屋情報、隣接部屋の仙獣/侵入者情報、RoomChi）
- [ ] `senju/action.go`: Action 構造体（Type ActionType, TargetRoomID int, TargetBeastID int）。ActionType 型（Stay/MoveToRoom/Attack/Retreat）。仙獣が1ティックで取れる行動を表現
- [ ] `senju/behavior_guard.go`: GuardBehavior — 定点防衛。自分の配置部屋に留まる。侵入者が同じ部屋にいればAttack、いなければStay。最もシンプルなAI
- [ ] `senju/behavior_guard_test.go`: 侵入者なし→Stay、侵入者あり→Attack のテスト

## Phase 3.5-B: 巡回と追跡AI（senju/）

- [ ] `senju/behavior_patrol.go`: PatrolBehavior — 巡回。隣接部屋を順番に移動する。巡回経路は配置部屋を起点に隣接グラフから生成。侵入者を発見したらChaseに遷移
- [ ] `senju/behavior_chase.go`: ChaseBehavior — 追跡。発見した侵入者の方向へ隣接部屋を移動。侵入者と同じ部屋に入ったらAttack。一定ティック追跡して見失ったらPatrolに戻る
- [ ] `senju/behavior_flee.go`: FleeBehavior — 逃走。HPが一定割合（25%）以下になったら発動。侵入者から最も遠い隣接部屋へ移動。回復室に到達したら回復状態（Recovering）に遷移
- [ ] `senju/behavior_test.go`: Patrol の巡回経路テスト、Chase の追跡方向テスト、Flee のHP閾値判定テスト、Flee が回復室を目指すテスト

## Phase 3.5-C: 行動エンジン（senju/）

- [ ] `senju/behavior_engine.go`: BehaviorEngine 構造体。NewBehaviorEngine(cave, adjacencyGraph)。AssignBehavior(beast, behaviorType) — 仙獣に行動パターンを割り当て。Tick(beasts, invaderPositions map[int][]int, roomChi) []BeastAction — 全仙獣の行動を1ティック分一括決定:
  1. 各仙獣の現在Behaviorに基づいてDecideAction呼び出し
  2. HP閾値チェック → Flee への自動遷移
  3. 行動の衝突解決（同じ部屋への移動は先着順）
  4. BeastAction リストを返す（実際の移動・攻撃の適用は呼び出し側）
- [ ] `senju/behavior_engine.go`: BeastAction 構造体（BeastID int, Action Action, PreviousRoomID int, ResultingState BeastState）。行動の結果を記録
- [ ] `senju/behavior_engine.go`: ApplyActions(beasts, rooms, actions) error — BeastAction リストを適用して仙獣の位置・状態を更新
- [ ] `senju/behavior_engine_test.go`: 全仙獣Guard→侵入者なしで全員Stay テスト、Patrol仙獣が隣接部屋を巡回するテスト、侵入者発見→Chase遷移テスト、HP低下→Flee遷移テスト、行動の衝突解決テスト

## Phase 3.5-D: 仙獣AI パラメータ外出し

- [ ] `senju/behavior_params.go`: BehaviorParams 構造体（FleeHPThreshold float64, ChaseTimeoutTicks int, PatrolRestTicks int）。DefaultBehaviorParams()。LoadBehaviorParams(data []byte)
- [ ] `senju/behavior_params_data.json`: デフォルト行動パラメータ（逃走HP閾値: 0.25, 追跡タイムアウト: 10ティック, 巡回時の部屋滞在: 3ティック）
- [ ] `senju/behavior_params_test.go`: JSONロードテスト、デフォルト値テスト

## Phase 3.5-E: ASCII可視化への行動レイヤー追加

- [ ] `senju/ascii.go` 更新: 仙獣の行動状態を表示に反映。Guard: `[G]`, Patrol: `[P]`, Chase: `[!]`, Flee: `[←]`, Recovering: `[+]`。侵入者位置のプレースホルダー表示（`??` — Phase 4で本実装）
- [ ] `cmd/caveviz/main.go` 更新: `--ai` フラグで行動状態レイヤー表示

## Phase 3.5-F: 統合検証

- [ ] `senju/ai_integration_test.go`: Cave（部屋5つ、通路接続）+ 仙獣3体（Guard×1, Patrol×1, Chase×1）+ 疑似侵入者位置（map[int][]int で手動設定）→20ティック行動シミュレーション→Guard仙獣は配置部屋に留まることを検証→Patrol仙獣は複数部屋を巡回することを検証→侵入者位置を設定するとChase仙獣が追跡方向に移動することを検証→HP低下でFleeに遷移することを検証
- [ ] `go vet ./...` と `go test -race ./...` がクリーンに通ることを確認
- [ ] Phase 3.5 完了。DECISIONS.md 更新、PHASE_COMPLETE 更新、次フェーズドラフトを `tasks_phase4_draft.md` として生成。**tasks.md には新しい未完了タスクを追加しない**
