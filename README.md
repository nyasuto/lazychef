# LazyChef 🍳

個人向けの自炊継続支援サービス。AIを活用して週単位の買い物リストとレシピを提案し、ずぼらな人でも継続できる自炊をサポートします。

## 🚀 クイックスタート

**最短3ステップで開始:**
```bash
# 1. リポジトリクローン
git clone https://github.com/nyasuto/lazychef.git
cd lazychef

# 2. 環境設定
cp .env.example .env
# .envファイルでOPENAI_API_KEYを設定

# 3. 起動
make quickstart
```

詳細は [QUICKSTART.md](./QUICKSTART.md) をご覧ください。

## 📋 主要機能

### 🤖 AI レシピ生成システム
- **GPT-5 Enhanced Generation** - 3段階生成プロセス（着想→執筆→批評）
- **AI自動生成 Phase 1** - ディメンション分析に基づく自動レシピ生成
- **多様性重視** - 5次元（料理分類・たんぱく質・調理法・調味・手軽さ）での分析
- **カバレッジ分析** - レシピの網羅率を可視化（現在2.8%→目標500組み合わせ）

### 🛡️ 食品安全・品質管理
- **食品安全チェック** - USDA基準による温度・衛生検証
- **品質管理システム** - GPT-5による構造化レシピ検証
- **重複検出** - Embedding による類似レシピ自動検出

### 🏭 高性能・スケーラブル機能
- **Batch API** - OpenAI Batch APIで50%コスト削減
- **高度なレート制御** - トークン・リクエスト・予算の3階層管理
- **コスト効率分析** - リアルタイム使用量監視

## 🎯 プロジェクト概要

### 解決する課題
- 一時的には自炊できるが、継続的に外食に戻ってしまう問題
- 毎日何を作るか考えるのが面倒
- 買い物で何を買えばいいか分からない
- 材料を無駄にしてしまう

### コアコンセプト
**「これを買えば1週間分作れる」まとめ買い提案システム**

## 🚀 技術スタック

### Phase 1: MVP（現在）
```yaml
Backend:
  - Go (Gin/Echo フレームワーク)
  - SQLite + JSON機能 (セットアップ不要)
  - OpenAI API (レシピ生成)

Frontend:
  - React (Claude Codeが開発担当)
  - Tailwind CSS (スタイリング)
  - Axios (API通信)

Deploy:
  - ローカル環境での開発
  - バイナリ配布 (Go)
  - 静的ファイルホスティング (React)
```

### Phase 2: 本番環境（将来）
- PostgreSQL + JSONB（必要に応じて移行）
- Docker化
- クラウドデプロイ（Fly.io / Railway）

## 📊 データベース設計

### SQLite + JSON構造

```sql
-- レシピテーブル
CREATE TABLE recipes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    data JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 週間献立テーブル  
CREATE TABLE meal_plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    week_data JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- ユーザー設定テーブル
CREATE TABLE user_preferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    preferences JSON NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### JSONデータ構造

#### レシピデータ
```json
{
  "title": "10分豚キャベツ炒め",
  "cooking_time": 10,
  "ingredients": [
    {"name": "豚こま肉", "amount": "200g"},
    {"name": "キャベツ", "amount": "1/4個"},
    {"name": "醤油", "amount": "大さじ1"}
  ],
  "steps": [
    "キャベツをざく切り",
    "豚肉炒めて、キャベツ入れる", 
    "醤油かけて完成"
  ],
  "tags": ["簡単", "豚肉", "10分以内"],
  "season": "all",
  "laziness_score": 9.5,
  "nutrition_info": {
    "calories": 250,
    "protein": 20
  }
}
```

#### 週間献立データ
```json
{
  "start_date": "2025-01-27",
  "shopping_list": [
    {"item": "豚こま肉", "amount": "400g"},
    {"item": "キャベツ", "amount": "1個"},
    {"item": "もやし", "amount": "3袋"},
    {"item": "卵", "amount": "1パック"}
  ],
  "daily_recipes": {
    "monday": {"recipe_id": 1, "title": "豚キャベツ炒め"},
    "tuesday": {"recipe_id": 2, "title": "もやし卵とじ"},
    "wednesday": {"recipe_id": 3, "title": "豚こまチャーハン"},
    "thursday": {"recipe_id": 4, "title": "キャベツの卵とじ"},
    "friday": {"recipe_id": 5, "title": "もやしと豚の味噌炒め"}
  },
  "total_cost_estimate": 1500
}
```

## 🔧 セットアップ手順

### 1. 環境準備
```bash
# リポジトリクローン
git clone https://github.com/yourname/lazychef.git
cd lazychef

# Go依存関係インストール
cd backend
go mod init lazychef
go get -u github.com/gin-gonic/gin
go get -u github.com/mattn/go-sqlite3
go get -u github.com/joho/godotenv

# React環境セットアップ (Claude Codeに依頼)
cd ../frontend
npx create-react-app . --template minimal
npm install axios tailwindcss
```

### 2. 環境変数設定
```bash
# .env ファイル作成
cp .env.example .env

# OpenAI APIキー設定
# .env
OPENAI_API_KEY=your_api_key_here
PORT=8080
FRONTEND_URL=http://localhost:3000
```

### 3. データベース初期化
```bash
cd backend
go run scripts/init_db.go
```

### 4. アプリケーション起動
```bash
# ターミナル1: バックエンド起動
cd backend
go run cmd/api/main.go

# ターミナル2: フロントエンド起動
cd frontend
npm start
```

### 5. Claude Codeでフロントエンド開発
```bash
# Claude Codeに以下を依頼
"LazyChefのフロントエンドを実装して。
- APIエンドポイント: http://localhost:8080
- 週間レシピ表示、買い物リスト、レシピ生成機能
- Tailwind CSSでシンプルなデザイン"
```

## 📁 プロジェクト構造

```
lazychef/
├── README.md            # このファイル
├── go.mod              # Go依存関係
├── go.sum
├── .env.example        # 環境変数テンプレート
├── Makefile           # ビルド・実行コマンド
│
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go                    # エントリーポイント
│   ├── internal/
│   │   ├── config/
│   │   │   └── openai.go                 # OpenAI API設定
│   │   ├── database/
│   │   │   └── sqlite.go                 # DB接続・初期化
│   │   ├── handlers/
│   │   │   ├── recipe_handler.go         # レシピ関連API
│   │   │   ├── meal_plan_handler.go      # 献立関連API
│   │   │   └── admin_handler.go          # 管理者用API
│   │   ├── models/
│   │   │   ├── recipe.go                 # レシピデータモデル
│   │   │   ├── meal_plan.go              # 献立データモデル
│   │   │   └── diversity.go              # 多様性分析モデル
│   │   ├── services/
│   │   │   ├── generator.go              # 基本レシピ生成
│   │   │   ├── enhanced_generator.go     # GPT-5 Enhanced生成
│   │   │   ├── auto_generation_service.go # AI自動生成システム
│   │   │   ├── batch_generator.go        # Batch API生成
│   │   │   ├── diversity_service.go      # 多様性分析サービス
│   │   │   ├── embedding_deduplicator.go # 重複検出システム
│   │   │   ├── food_safety_validator.go  # 食品安全検証
│   │   │   ├── quality_check_service.go  # 品質チェック
│   │   │   └── token_rate_limiter.go     # 高度レート制御
│   │   └── middleware/
│   │       └── cors.go                   # CORS設定
│   └── data/
│       └── recipes.db                    # SQLiteデータベース
│
├── frontend/              # Claude Codeが主担当
│   ├── package.json
│   ├── package-lock.json
│   ├── public/
│   │   └── index.html
│   ├── src/
│   │   ├── App.js
│   │   ├── index.js
│   │   ├── api/
│   │   │   └── client.js     # API通信
│   │   ├── components/
│   │   │   ├── RecipeCard.jsx
│   │   │   ├── WeeklyPlan.jsx
│   │   │   ├── ShoppingList.jsx
│   │   │   └── RecipeGenerator.jsx
│   │   └── styles/
│   │       └── index.css     # Tailwind CSS
│   └── .gitignore
│
└── scripts/
    ├── init_db.sql          # DB初期化スクリプト
    ├── generate_recipes.go  # レシピ一括生成
    └── migrate_to_postgres.sql  # PostgreSQL移行用
```

## 🎯 主要機能

### ✅ 完了機能（2025年8月現在）

#### 🤖 AI生成システム
- [x] **基本レシピ生成** - GPT-3.5による単発生成
- [x] **GPT-5 Enhanced生成** - 3段階プロセス（着想→執筆→批評）
- [x] **AI自動生成 Phase 1** - ディメンション分析による自動生成
- [x] **多様性分析** - 5次元でのカバレッジ分析
- [x] **食品安全検証** - USDA基準準拠の温度・衛生チェック
- [x] **品質管理** - 構造化レシピ検証

#### 🏭 スケーラブル機能
- [x] **Batch API** - 大規模生成（50%コスト削減）
- [x] **重複検出** - Embedding-based similarity検出
- [x] **高度レート制御** - 3階層（リクエスト・トークン・予算）管理
- [x] **コスト監視** - リアルタイム使用量・効率性分析

#### 📱 基本機能
- [x] **レシピ検索** - 多様な条件での絞り込み
- [x] **週間献立作成** - ユーザー設定に基づく自動計画
- [x] **買い物リスト生成** - 食材の最適化・統合

### 🚧 開発中機能

#### Phase 2: 完全自動化（予定）
- [ ] **AI完全自動生成** - 人間の介入なしでの大規模レシピ生成
- [ ] **実時間品質フィルタ** - 生成と同時の品質・安全性チェック
- [ ] **学習型多様性制御** - ユーザー好みを反映した多様性バランス

#### Phase 3: UX強化（予定）
- [ ] **冷蔵庫スキャン** - 画像認識による食材管理
- [ ] **調理履歴記録** - パーソナライズ学習
- [ ] **栄養バランス分析** - 健康管理連携

## 🔨 開発履歴・今後のタスク

### ✅ 完了済み開発 (2025年8月)

#### Phase 0: 基盤構築
- [x] **Go + Gin基盤** - RESTful API基盤実装
- [x] **SQLiteデータベース** - JSON機能活用したスキーマレス設計
- [x] **OpenAI API連携** - GPT-3.5/4による基本生成
- [x] **CORS・ミドルウェア** - フロントエンド連携設定

#### Phase 1: GPT-5基盤構築
- [x] **Enhanced Generation** - 3段階生成プロセス実装
- [x] **構造化出力** - JSON Schema準拠レシピ生成
- [x] **食品安全システム** - USDA基準準拠検証エンジン
- [x] **品質管理システム** - GPT-5による自動品質チェック

#### Phase 2: スケール・効率化
- [x] **Batch API統合** - OpenAI Batch API活用（50%コスト削減）
- [x] **Embedding重複検出** - 類似レシピ自動検出システム
- [x] **高度レート制御** - トークン・リクエスト・予算の3階層管理
- [x] **リアルタイム監視** - コスト効率性分析ダッシュボード

#### Phase 3: 多様性・自動化
- [x] **多様性分析エンジン** - 5次元での網羅率分析
- [x] **AI自動生成 Phase 1** - ディメンション分析による戦略的生成

### 🚧 進行中・次期開発

#### Phase 4: 完全自動化 (開発中)
- [ ] **AI自動生成 Phase 2** - 実際のGPT生成統合
- [ ] **バックグラウンド処理** - 大規模生成のキュー管理
- [ ] **品質フィルタリング** - 生成時リアルタイム品質チェック

#### Phase 5: フロントエンド強化 (Claude Code担当)
- [ ] **管理ダッシュボード** - 分析結果の可視化UI
- [ ] **レシピ管理UI** - 生成・編集・削除インターフェース
- [ ] **モニタリングUI** - コスト・使用量リアルタイム表示

## 🧪 テスト実行

```bash
# バックエンドテスト
cd backend
go test ./...
go test -v -cover ./...

# フロントエンドテスト（Claude Codeが作成）
cd frontend
npm test
```

## 📝 API仕様（Go Backend）

### 🤖 AI レシピ生成
```bash
# 基本レシピ生成
POST /api/recipes/generate
{
  "ingredients": ["豚肉", "キャベツ", "もやし"],
  "season": "winter",
  "max_cooking_time": 15
}

# GPT-5 Enhanced 生成
POST /api/recipes/generate-enhanced
{
  "stage": "authoring",
  "reasoning_effort": "high",
  "structured_outputs": true,
  "ingredients": ["鶏肉", "玉ねぎ"],
  "max_cooking_time": 10
}

# AI自動生成（Phase 1）
POST /api/admin/auto-generation/generate
{
  "count": 5,
  "strategy": "diversity_gap_fill",
  "max_cooking_time": 15
}
```

### 📊 分析・管理機能
```bash
# カバレッジ分析
GET /api/admin/auto-generation/coverage
# → レシピの多様性分析、不足領域の特定

# 重複スキャン
POST /api/admin/duplicate-detection/scan
{
  "similarity_threshold": 0.85,
  "batch_size": 100
}

# トークン使用量監視
GET /api/admin/metrics/token-usage
GET /api/admin/metrics/cost-efficiency
```

### 🛡️ 品質・安全チェック
```bash
# 食品安全検証
POST /api/recipes/validate-safety
{
  "title": "鶏肉のソテー",
  "ingredients": [...],
  "steps": [...]
}

# 品質チェック
POST /api/recipes/validate-quality
{
  "title": "簡単パスタ",
  "cooking_time": 15,
  "steps": [...]
}
```

### 🗂️ データ操作
```bash
# レシピ検索
GET /api/recipes/search?tag=簡単&ingredient=豚肉&limit=20

# 週間献立作成
POST /api/meal-plans/create
{
  "start_date": "2025-01-27",
  "preferences": {
    "max_cooking_time": 15
  }
}
```

### CORS設定
バックエンドは `http://localhost:3000` からのリクエストを許可

## 🚀 デプロイ

### Streamlitでの超高速PoC

もし最速でPoCを試したい場合：

```go
// backend/cmd/simple/main.go
package main

import (
    "encoding/json"
    "net/http"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()
    
    // CORS設定
    r.Use(func(c *gin.Context) {
        c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
        c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        c.Header("Access-Control-Allow-Headers", "Content-Type")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    })
    
    r.GET("/api/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    r.POST("/api/recipes/generate", generateRecipes)
    r.POST("/api/meal-plans/create", createMealPlan)
    
    r.Run(":8080")
}
```

### フロントエンド開発の役割分担

| タスク | 担当 | 詳細 |
|--------|------|------|
| API設計・実装 | あなた | Go + Gin/Echo |
| UI実装 | Claude Code | React + Tailwind |
| データベース | あなた | SQLite + JSON |
| スタイリング | Claude Code | レスポンシブデザイン |
| エラーハンドリング | 両方 | API側とUI側で協調 |

## 📊 パフォーマンス目標・実績

### ⚡ 現在の実績
- **基本レシピ生成**: 2-4秒（GPT-3.5/4）
- **Enhanced生成**: 8-15秒（GPT-5 3段階）
- **カバレッジ分析**: ~50ms（14レシピ、5次元分析）
- **重複検出**: ~200ms（100レシピ比較）
- **データベース検索**: < 50ms（SQLite + JSON）

### 🎯 目標値
- **AI自動生成**: < 30秒（10レシピ一括）
- **Batch処理**: ~24時間（1000レシピ）
- **品質チェック**: < 5秒（安全性・品質統合）
- **管理画面ロード**: < 2秒

### 💰 コスト効率
- **Batch API**: 通常APIの50%コスト
- **キャッシュ命中率**: >30% (24時間持続)
- **重複検出による節約**: ~20%（類似レシピ生成回避）

## 🔄 PostgreSQL移行計画

SQLiteからPostgreSQLへの移行は以下の条件で検討:
- ユーザー数 > 100人
- データサイズ > 1GB
- 同時接続要求 > 10

移行スクリプト: `scripts/migrate_to_postgres.py`

## 🐛 既知の課題・改善点

### 🔍 現在の課題
- [ ] **調理時間の精度**: AIレシピの調理時間が実際より短い場合がある
- [ ] **食材単位統一**: 「適量」「少々」などの曖昧な表記が混在
- [ ] **季節判定**: より精密な旬の食材推奨システムが必要
- [ ] **Phase 1制約**: 自動生成がまだプレースホルダー（Phase 2で実装予定）

### 🚀 技術的改善項目
- [ ] **並列処理**: 大量生成時のパフォーマンス最適化
- [ ] **エラー回復**: 生成失敗時の自動リトライ機能
- [ ] **メモリ最適化**: 長時間稼働時のメモリリーク防止
- [ ] **ログ強化**: より詳細な分析・デバッグ情報

### 🎯 UX改善目標
- [ ] **生成進捗表示**: リアルタイム進捗バーの実装
- [ ] **エラー通知**: ユーザーフレンドリーなエラーメッセージ
- [ ] **レシピ編集機能**: 生成後の微調整インターフェース

## 📚 参考資料

- [SQLite JSON1 Documentation](https://www.sqlite.org/json1.html)
- [FastAPI Documentation](https://fastapi.tiangolo.com/)
- [Streamlit Documentation](https://docs.streamlit.io/)
- [OpenAI API Reference](https://platform.openai.com/docs/)

## 📄 ライセンス

MIT License

## 🤝 貢献

個人プロジェクトですが、フィードバック歓迎です！

---

**開発メモ**: 
- まずは自分が1ヶ月使い続けられるものを作る
- 完璧を求めず、動くものを優先
- ユーザーフィードバックベースで改善