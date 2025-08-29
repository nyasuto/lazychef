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

- **🤖 AI レシピ生成** - GPT-5による多段階生成（Phase 0）
- **🛡️ 食品安全チェック** - USDA基準による温度検証
- **💰 Batch API** - 50%コスト削減の大規模生成（Phase 1）
- **🔍 重複検出** - Embedding による類似レシピ検出
- **📊 コスト管理** - トークン使用量・予算監視

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
│   │       └── main.go         # エントリーポイント
│   ├── internal/
│   │   ├── database/
│   │   │   └── sqlite.go       # DB接続・初期化
│   │   ├── handlers/
│   │   │   ├── recipe.go      # レシピ関連API
│   │   │   └── planner.go     # 献立関連API
│   │   ├── models/
│   │   │   └── recipe.go      # データモデル
│   │   ├── services/
│   │   │   ├── generator.go   # AI レシピ生成
│   │   │   └── planner.go     # 週間献立作成
│   │   └── middleware/
│   │       └── cors.go        # CORS設定
│   └── data/
│       └── recipes.db         # SQLiteデータベース
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

### MVP機能（Phase 1）
- [x] AIによるずぼらレシピ生成
- [x] 週間献立の自動作成
- [x] 買い物リスト生成
- [x] 材料の使い回し提案
- [x] 季節対応レシピ
- [ ] ユーザー好み学習（基本版）

### 発展機能（Phase 2）
- [ ] 冷蔵庫スキャン機能
- [ ] レシピのお気に入り機能
- [ ] 調理履歴の記録
- [ ] 栄養バランス分析
- [ ] ゲーミフィケーション要素

## 🔨 開発タスク

### 週1: バックエンド基礎実装（あなた担当）
- [ ] Go + Ginでプロジェクトセットアップ
- [ ] SQLiteデータベース接続
- [ ] OpenAI API連携
- [ ] 基本的なAPIエンドポイント作成
- [ ] CORS設定

### 週2: フロントエンド実装（Claude Code担当）
- [ ] React基本セットアップ
- [ ] APIクライアント実装
- [ ] レシピ表示コンポーネント
- [ ] 買い物リスト表示
- [ ] レシピ生成フォーム

### 週3: コア機能統合
- [ ] 週間献立生成アルゴリズム（バックエンド）
- [ ] 材料使い回しロジック（バックエンド）
- [ ] UI/UXブラッシュアップ（Claude Code）

### 週4: 最適化・改善
- [ ] レシピ生成の精度向上
- [ ] パフォーマンス最適化
- [ ] エラーハンドリング強化
- [ ] デプロイ準備

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

### レシピ生成
```
POST /api/recipes/generate
Content-Type: application/json

{
  "ingredients": ["豚肉", "キャベツ", "もやし"],
  "season": "winter",
  "max_cooking_time": 15
}

Response:
{
  "recipes": [
    {
      "id": 1,
      "title": "10分豚キャベツ炒め",
      "cooking_time": 10,
      "ingredients": [...],
      "steps": [...]
    }
  ]
}
```

### 週間献立作成
```
POST /api/meal-plans/create
Content-Type: application/json

{
  "start_date": "2025-01-27",
  "preferences": {
    "exclude_ingredients": ["パクチー"],
    "max_cooking_time": 10
  }
}

Response:
{
  "id": "plan_123",
  "shopping_list": [...],
  "daily_recipes": {...}
}
```

### レシピ検索
```
GET /api/recipes/search?tag=10分以内&ingredient=豚肉

Response:
{
  "recipes": [...],
  "total": 5
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

## 📊 パフォーマンス目標

- レシピ生成: < 3秒
- 週間献立作成: < 5秒
- データベース検索: < 100ms
- 初期ロード: < 2秒

## 🔄 PostgreSQL移行計画

SQLiteからPostgreSQLへの移行は以下の条件で検討:
- ユーザー数 > 100人
- データサイズ > 1GB
- 同時接続要求 > 10

移行スクリプト: `scripts/migrate_to_postgres.py`

## 🐛 既知の課題

- [ ] AIレシピの調理時間が実際より短い場合がある
- [ ] 材料の単位統一が不完全
- [ ] 季節判定の精度向上が必要

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