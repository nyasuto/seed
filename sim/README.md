# chaosseed-sim

カオスシード風ダンジョン経営シミュレーションの CLI シミュレーター。

## ビルド

```bash
cd sim && make build
# または
go build -o chaosseed-sim ./cmd/chaosseed-sim
```

## モード一覧

### Human Mode（対話プレイ）

```bash
chaosseed-sim --human [--scenario tutorial]
```

ターミナル上でメニュー操作によりゲームをプレイする。ASCII描画でダンジョンの状態を表示。早送り機能あり。

### AI Mode（外部プログラム連携）

```bash
chaosseed-sim --ai --scenario tutorial [--timeout 30s]
```

JSON Lines プロトコルで外部プログラム（LLM等）と stdin/stdout で通信する。
毎ティック `state` メッセージ（ゲーム状態 + `valid_actions`）を送信し、クライアントは `action` メッセージで応答する。

プロトコル仕様: [adapter/ai/docs/PROTOCOL.md](./adapter/ai/docs/PROTOCOL.md)

### Batch Mode（統計収集）

```bash
chaosseed-sim --batch --scenario tutorial --games 1000 [--batch-ai simple] [--output results.json] [--format json]
```

指定シナリオを N ゲーム並列実行し、統計レポートを出力する。

オプション:
- `--games N`: 実行ゲーム数（デフォルト: 100）
- `--batch-ai simple|noop`: AI戦略（デフォルト: noop）
- `--output PATH`: 結果出力先（デフォルト: stdout）
- `--format json|csv`: 出力形式（デフォルト: json）
- `--sweep key=v1,v2,v3`: パラメータスイープ

### Balance Dashboard（バランス調整）

```bash
chaosseed-sim --balance --scenario tutorial [--games 100]
```

4段階のインタラクティブフロー:
1. ベースラインバッチ実行
2. 壊れるサイン（B01〜B11）のアラート表示
3. D002ルールに基づくスイープ提案
4. パラメータスイープ比較

### Replay Mode（リプレイ再生）

```bash
chaosseed-sim --replay path/to/replay.json --scenario tutorial
```

保存されたアクション履歴を再生し、決定論的に同一結果を再現する。

## シナリオ

組み込みシナリオ:
- `tutorial`: 初心者向け。リソース潤沢、侵入1波
- `standard`: 標準難易度。複数波、動的スケジューリング

カスタムシナリオはファイルパスで指定可能:
```bash
chaosseed-sim --human --scenario ./my_scenario.json
```

## 壊れるサインメトリクス（B01〜B11）

Batch Mode と Balance Dashboard で検出される11種の「壊れている兆候」メトリクス。
面白さの証明ではなく、バランスが壊れている可能性を示す指標。

詳細は [docs/PRD.md](./docs/PRD.md) セクション3.4 を参照。

## 開発

```bash
make check    # vet + lint + test-race
make test     # テスト実行
make build    # バイナリ生成
make clean    # クリーンアップ
```
