# chaosseed-core

カオスシード（風水回廊記）にインスパイアされたダンジョン経営シミュレーションのコアメカニクスエンジン。
純Go実装、ゲームエンジン非依存、決定論的設計。

## 特徴

- 五行（木火土金水）の相生・相克に基づくゲームメカニクス
- 決定論的設計: 同じ乱数シードで同じ結果を保証
- 標準ライブラリのみで構築（外部依存なし）
- 後続の CLI シミュレーター / GUI クライアントから利用可能なライブラリ

## アーキテクチャ

```
chaosseed-core/
├── types/        # 共有型定義（Pos, Direction, Element, RNG, Tick）
├── world/        # 洞窟マップシステム（Grid, Room, Corridor, Cave）
├── fengshui/     # 風水システム（後続フェーズ）
├── senju/        # 仙獣システム（後続フェーズ）
├── invasion/     # 侵入システム（後続フェーズ）
├── economy/      # 経済システム（後続フェーズ）
├── scenario/     # シナリオシステム（後続フェーズ）
├── simulation/   # 統合シミュレーション（後続フェーズ）
└── testutil/     # テストヘルパー
```

パッケージ依存方向: `types` ← `world` ← `fengshui` ← `senju` ← `invasion` ← `economy` ← `scenario` ← `simulation`

## ビルドとテスト

```bash
# テスト実行
go test ./...

# race detector 付き
go test -race ./...

# vet
go vet ./...

# カバレッジ
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 必要要件

- Go 1.22 以上

## ライセンス

MIT License. 詳細は [LICENSE](LICENSE) を参照。
