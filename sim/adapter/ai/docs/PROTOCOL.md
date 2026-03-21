# AI Mode Protocol Specification

chaosseed-sim AI Mode は JSON Lines プロトコルで外部プログラム（LLM等）と通信します。

## 起動方法

```bash
chaosseed-sim --ai --scenario tutorial [--timeout 30s]
```

- `--scenario`: シナリオ名またはファイルパス（デフォルト: `tutorial`）
- `--timeout`: アクション入力のタイムアウト（デフォルト: なし）。タイムアウト時は wait が実行される

## 通信フロー

```
Server (stdout)              Client (stdin)
     |                            |
     |--- state ----------------->|  (毎ティック送信)
     |                            |
     |<-- action -----------------| (クライアントが応答)
     |                            |
     |--- state ----------------->|  (次ティック)
     |          ...               |
     |--- game_end -------------->|  (ゲーム終了)
```

エラー時のリトライフロー:

```
Server                       Client
     |--- state ----------------->|
     |<-- (invalid input) --------|
     |--- error ----------------->|  (エラー内容)
     |--- state ----------------->|  (同じ state を再送)
     |<-- action -----------------| (クライアントが再応答)
```

最大3回リトライ後、自動的に wait アクションが実行されます。

## メッセージ型

すべてのメッセージは1行の JSON（JSON Lines 形式）です。
メッセージは `type` フィールドで識別します。

### Server → Client

#### `state` — ゲーム状態通知

毎ティック送信されます。現在のゲーム状態と、このティックで実行可能なアクション一覧を含みます。

```json
{
  "type": "state",
  "tick": 1,
  "snapshot": {
    "tick": 1,
    "core_hp": 100,
    "chi_pool_balance": 500.0,
    "rooms": [...],
    "beasts": [...],
    "cave_feng_shui_score": 0.5,
    "active_invaders": 0,
    "defeated_waves": 0
  },
  "valid_actions": [
    {"kind": "wait", "params": {}},
    {"kind": "dig_room", "params": {"room_type_id": "senju_room", "x": 3, "y": 5, "width": 3, "height": 3, "cost": 50.0}},
    {"kind": "summon_beast", "params": {"element": "Wood", "cost": 30.0}}
  ]
}
```

- `snapshot`: ゲームの現在状態（部屋、仙獣、気のバランス等）
- `valid_actions`: このティックで実行可能なアクションのリスト。**ここに含まれないアクションは拒否されます**

#### `game_end` — ゲーム終了通知

ゲームが勝利または敗北で終了した際に送信されます。

```json
{
  "type": "game_end",
  "result": "victory",
  "summary": {
    "peak_chi": 1200.5,
    "waves_defeated": 3,
    "final_feng_shui": 0.85,
    "evolutions": 1,
    "damage_dealt": 450,
    "damage_received": 120,
    "deficit_ticks": 5
  },
  "metrics": null
}
```

- `result`: `"victory"` または `"defeat"`
- `summary`: ゲーム統計情報
- `metrics`: メトリクスデータ（将来拡張用、現在は null）

#### `error` — エラー通知

クライアントの入力が不正だった場合に送信されます。エラー後、同じ `state` メッセージが再送されます。

```json
{
  "type": "error",
  "message": "invalid JSON: unexpected end of JSON input"
}
```

### Client → Server

#### `action` — アクション指定

`state` メッセージへの応答として送信します。1ティックにつき1つの `action` メッセージを送信してください。

```json
{
  "type": "action",
  "actions": [
    {"kind": "wait", "params": {}}
  ]
}
```

- `actions` 配列は1つ以上のアクションを含む必要があります
- 各アクションは `valid_actions` に含まれるものでなければなりません

## アクション種別

### `wait` — 何もしない

```json
{"kind": "wait", "params": {}}
```

常に実行可能です。

### `dig_room` — 部屋を掘る

```json
{"kind": "dig_room", "params": {"room_type_id": "senju_room", "x": 3, "y": 5, "width": 3, "height": 3}}
```

| パラメータ | 型 | 説明 |
|-----------|------|------|
| room_type_id | string | 部屋タイプID |
| x | number | X座標 |
| y | number | Y座標 |
| width | number | 幅 |
| height | number | 高さ |

`valid_actions` に含まれる `cost` フィールドは情報提供用です（送信不要）。

### `dig_corridor` — 通路を掘る

```json
{"kind": "dig_corridor", "params": {"from_room_id": 1, "to_room_id": 2}}
```

| パラメータ | 型 | 説明 |
|-----------|------|------|
| from_room_id | number | 接続元の部屋ID |
| to_room_id | number | 接続先の部屋ID |

### `summon_beast` — 仙獣を召喚する

```json
{"kind": "summon_beast", "params": {"element": "Wood"}}
```

| パラメータ | 型 | 説明 |
|-----------|------|------|
| element | string | 属性（`Wood`, `Fire`, `Earth`, `Metal`, `Water`） |

### `upgrade_room` — 部屋をアップグレードする

```json
{"kind": "upgrade_room", "params": {"room_id": 1}}
```

| パラメータ | 型 | 説明 |
|-----------|------|------|
| room_id | number | 対象の部屋ID |

### `evolve_beast` — 仙獣を進化させる

```json
{"kind": "evolve_beast", "params": {"beast_id": 1, "to_species_id": "fire_dragon"}}
```

| パラメータ | 型 | 説明 |
|-----------|------|------|
| beast_id | number | 対象の仙獣ID |
| to_species_id | string | 進化先の種族ID |

### `place_beast` — 仙獣を部屋に配置する

```json
{"kind": "place_beast", "params": {"species_id": "fire_lizard", "room_id": 1}}
```

| パラメータ | 型 | 説明 |
|-----------|------|------|
| species_id | string | 仙獣の種族ID |
| room_id | number | 配置先の部屋ID |

## エラー処理

1. **不正な JSON**: `error` メッセージが返され、`state` が再送される
2. **不正なメッセージ型**: `type` が `"action"` でない場合はエラー
3. **空の actions 配列**: エラー
4. **valid_actions にないアクション**: エラー
5. **最大リトライ超過**: 3回のリトライ後、自動的に wait が実行される
6. **EOF（入力終了）**: ゲームが終了する
7. **タイムアウト**: `--timeout` 指定時、時間内に応答がなければ wait が実行される

## 使用例

### 最小限のクライアント（シェルスクリプト）

```bash
# 全ティック wait で完走
yes '{"type":"action","actions":[{"kind":"wait","params":{}}]}' | \
  chaosseed-sim --ai --scenario tutorial
```

### Python クライアント例

```python
import json
import sys

for line in sys.stdin:
    msg = json.loads(line)
    if msg["type"] == "game_end":
        break
    if msg["type"] == "state":
        # valid_actions から最初の非wait アクションを選択
        actions = msg["valid_actions"]
        chosen = next((a for a in actions if a["kind"] != "wait"), actions[-1])
        response = {"type": "action", "actions": [chosen]}
        print(json.dumps(response), flush=True)
```
