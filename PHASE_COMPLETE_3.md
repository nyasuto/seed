# PHASE_COMPLETE_3 — Phase 3: シーン管理 + UI

## 実装した内容

Phase 3 では SceneManager によるシーンライフサイクル管理と、全画面シーン（Title, Select, InGame, Result）+ InfoPanel を実装した。

### Task 3-A: シーンマネージャー
- `scene/manager.go`: SceneManager — Switch で OnExit/OnEnter を呼び、Update/Draw を委譲
- `Scene` インターフェースの Draw 引数は `image.Image` で headless テスト対応（D020）

### Task 3-B: タイトル画面とシナリオ選択画面
- `scene/title.go`: TitleScene — ゲーム名表示 + [新しいゲーム]/[ロード] ボタン
- `scene/select.go`: ScenarioSelectScene — 組み込みシナリオ一覧からの選択

### Task 3-C: InGame シーンへの統合
- `scene/ingame.go`: InGameScene — Phase 1-2 の全コンポーネントを統合
- main.go の Game 構造体から責務を移管、SceneManager 経由でシーン切り替え
- ゲーム終了検知 → onGameOver コールバック → Result シーンへ遷移

### Task 3-D: 結果画面
- `scene/result.go`: ResultScene — 勝利/敗北テキスト + 統計サマリー
- [もう一度]/[タイトルへ] ボタンによるナビゲーション

### Task 3-E: Info Panel
- `view/infopanel.go`: InfoPanel — 部屋/仙獣/侵入者の詳細表示
- ModeNormal での部屋クリック → 詳細情報表示
- 何も選択されていない場合はゲーム全体情報を表示

### Task 3-F: Phase 3 統合と確認
- `go build ./...` 成功
- `go vet ./...` クリーン
- `go test ./controller/... -count=1 -race` パス（headless-safe）
- ebiten 依存パッケージ（asset/view/input/scene）のテストはディスプレイ環境でのみ実行可能（LESSONS.md 記載済み）
- DECISIONS.md 棚卸し（D020 追加、全エントリ確認）
- LESSONS.md Phase 3 知見追記

## シーンフロー

```
TitleScene → [新しいゲーム] → ScenarioSelectScene → [シナリオ選択] → InGameScene
                                                                        ↓
                                                               [ゲーム終了]
                                                                        ↓
TitleScene ← [タイトルへ] ← ResultScene ← [もう一度] → ScenarioSelectScene
```

## 未解決の課題

- `game/testdata/tutorial.json` の embed 重複（core の `LoadBuiltinScenario` への統一は Phase 4 以降で検討）
- `BuildRoomRenderMap` の毎フレーム再構築（パフォーマンス問題が顕在化した場合に対応）
- ebiten 依存パッケージのテストが headless 環境で実行不可（GLFW init panic）。scene パッケージの manager_test.go は headless 対応だが、同一パッケージ内の他ファイルが ebiten を import するため実行できない

## 次フェーズへの申し送り

- Phase 4 はセーブ/ロード + 戦闘フィードバック
- InGameScene.Controller() メソッドで GameController にアクセス可能（セーブ/ロード用）
- TitleScene の [ロード] ボタンはスタブ状態。Phase 4-B で実装
- 戦闘/侵入の視覚フィードバック（Phase 4-C）は InGameScene の Draw に追加

## LESSONS.md から特に重要な知見

- Scene インターフェースの Draw を `image.Image` にすることで SceneManager テストが headless 対応（D020）
- scene パッケージ内テストは同一パッケージの ebiten import により headless で panic する。完全な headless テストにはパッケージ分離が必要
- InGameScene は Phase 1-2 の全コンポーネントの統合点。onGameOver コールバックでシーン遷移を main.go の Game 構造体に委譲する設計
