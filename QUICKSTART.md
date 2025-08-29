# 🚀 LazyChef クイックスタートガイド

LazyChef AI料理アシスタントをすぐに試すための最短手順です。

## ⚡ 3ステップで開始

### 🖥️ API版 (バックエンドのみ)

#### 1. 環境設定 (初回のみ)
```bash
# OpenAI APIキーを設定
cp .env.example .env
# .envファイルを編集してOPENAI_API_KEYを設定してください
```

#### 2. セットアップ & 起動
```bash
# バックエンドAPI自動セットアップ
make quickstart
```

#### 3. 確認
ブラウザで http://localhost:8080/api/health にアクセス

### 🌐 GUI版 (フロントエンド + バックエンド)

#### 1. 環境設定 (初回のみ)
```bash
# OpenAI APIキーを設定
cp .env.example .env
# .envファイルを編集してOPENAI_API_KEYを設定してください
```

#### 2. セットアップ & 起動
```bash
# GUI版完全自動セットアップ
make quickstart-gui
```

#### 3. 確認
- **GUI**: http://localhost:3000 (メインインターフェース)
- **API**: http://localhost:8080/api/health (APIヘルスチェック)

## 🎯 PoC デモ

### 基本的なレシピ生成
```bash
# シンプルなレシピ生成
curl -X POST http://localhost:8080/api/recipes/generate \
  -H "Content-Type: application/json" \
  -d '{"preferences": {"cooking_time": 10, "ingredients": ["卵", "パン"]}}'
```

### GPT-5拡張機能デモ
```bash
# 食品安全チェック付きレシピ生成
curl -X POST http://localhost:8080/api/recipes/generate-enhanced \
  -H "Content-Type: application/json" \
  -d '{"preferences": {"cooking_time": 15, "ingredients": ["鶏肉", "野菜"]}}'
```

### Batch API機能 (Phase 1)
```bash
# 管理画面の確認
curl http://localhost:8080/api/admin/health

# 重複検出デモ
curl -X POST http://localhost:8080/api/admin/duplicate-detection/scan \
  -H "Content-Type: application/json" \
  -d '{}'
```

## 📋 主要エンドポイント

| 機能 | URL | 説明 |
|------|-----|------|
| ヘルスチェック | `GET /api/health` | サーバー状態確認 |
| レシピ生成 | `POST /api/recipes/generate` | 基本レシピ生成 |
| 拡張生成 | `POST /api/recipes/generate-enhanced` | GPT-5拡張機能 |
| 食品安全検証 | `POST /api/recipes/validate-safety` | 安全性チェック |
| 献立作成 | `POST /api/meal-plans/create` | 週間献立生成 |
| 管理機能 | `GET /api/admin/health` | システム管理 |

## 🔧 開発コマンド

### バックエンドのみ
```bash
# 開発サーバー起動（ホットリロード）
make dev

# テスト実行
make test

# 品質チェック
make quality
```

### フロントエンド開発
```bash
# フロントエンド依存関係インストール
make frontend-install

# フロントエンド開発サーバー起動
make frontend-dev

# フロントエンドビルド
make frontend-build
```

### フルスタック開発
```bash
# バックエンド + フロントエンド同時起動
make fullstack-dev

# データベース再初期化
make reset-db

# デモデータ投入
make demo-data
```

## 🎮 PoC シナリオ

### シナリオ1: 基本的な料理生成
1. 冷蔵庫にある材料でレシピ提案
2. 調理時間・難易度による絞り込み
3. LazyChef制約チェック（≤3手順、≤15分）

### シナリオ2: 高度な機能
1. GPT-5による多段階生成（発想→作成→評価）
2. 食品安全チェック（USDA温度基準）
3. 重複レシピの自動検出

### シナリオ3: 大規模運用
1. Batch APIによるコスト効率的生成
2. Embeddingによる類似度検出
3. トークン使用量・コスト監視

## 🐛 トラブルシューティング

### よくある問題

**1. `OPENAI_API_KEY not found`**
```bash
# .envファイルを確認
cat .env
# APIキーを設定
echo "OPENAI_API_KEY=sk-..." >> .env
```

**2. データベースエラー**
```bash
# データベースを再初期化
make reset-db
```

**3. ポート競合**
```bash
# 別ポートで起動
PORT=8081 make run
```

## 📊 ログ確認

```bash
# リアルタイムログ
make logs

# エラーログのみ
make logs-errors

# API アクセスログ
make logs-api
```

## 🚀 次のステップ

1. **フロントエンド**: `make frontend-setup` でReact UI構築
2. **Docker**: `make docker-run` でコンテナ運用
3. **本番環境**: `make deploy` でデプロイ準備

## 📝 API仕様詳細

詳細なAPI仕様は以下をご参照ください：
- [API Documentation](./docs/API.md)
- [Phase 1 Features](./docs/PHASE1.md)
- [Architecture](./docs/ARCHITECTURE.md)