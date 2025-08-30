# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## System Context
- **Project**: LazyChef - AI-powered meal planning for lazy cooks
- **Backend**: Go 1.21 + Gin 1.10 + SQLite3
- **Frontend**: React 18 + Tailwind CSS (not yet implemented)
- **AI**: OpenAI API (gpt-3.5-turbo)

## File Map
```
backend/
├── cmd/api/main.go         # Entry point
├── internal/
│   ├── database/           # SQLite connection wrapper
│   ├── handlers/           # HTTP request handlers
│   ├── services/           # Business logic (generator, planner)
│   ├── models/             # Data structures
│   └── config/             # Configuration management
├── data/                   # SQLite database files (gitignored)
scripts/
├── init_db.sql             # Database schema
└── init_db.go              # Database initialization
```

## Commands
```bash
# Backend
make run                    # Start API server (port 8080)
make test                   # Run all tests with coverage
make build                  # Build binary to bin/lazychef
make init-db                # Initialize database

# Database
cd scripts && go run init_db.go  # Create DB with sample data

# Quick test
curl localhost:8080/api/health   # Health check
```
## MCP
- use playwright MCP to debug and test UI
- use o3 search MCP for technical research or ask your query from 3rd person perspective

## Critical Constraints
- **NEVER commit to main branch** - Always use feature branches
- **NEVER merge PRs automatically** - Human review required
- **Branch naming**: `feat/issue-X-description`, `fix/issue-X-description`
- **All GitHub issues and PRs must be in Japanese**

## Issue Management Process
**CRITICAL: 守るべきプロセス**

1. **Issue解決の正しい順序:**
   - ✅ 実装完了
   - ✅ 品質チェック (`make quality`)
   - ✅ PR作成 (`gh pr create`)
   - ✅ **PR先行 - マージ待ち**
   - ✅ 人間がPRレビュー・マージ
   - ✅ **マージ後にIssueクローズ**

2. **❌ 避けるべき行動:**
   - **PRマージ前のIssueクローズ** - 順序が逆
   - **実装中のIssueクローズ** - 作業未完了
   - **PR作成前のIssueクローズ** - レビューなし

3. **✅ 正しいタイミング:**
   - Issue実装 → PR作成 → 人間レビュー → マージ → **その後** Issue クローズ
   - 緊急修正の場合のみ例外検討

## API Endpoints
```
POST /api/recipes/generate        # Generate single recipe
POST /api/recipes/generate-batch  # Generate multiple recipes
POST /api/meal-plans/create       # Create weekly plan
GET  /api/recipes/search          # Search recipes
GET  /api/health                  # Service health check
```

## Database Schema
- **recipes**: JSON data column with virtual columns for indexing
- **meal_plans**: Weekly plans with shopping lists
- **user_preferences**: User settings (single-user MVP)
- Use `json_extract()` for queries: `WHERE json_extract(data, '$.laziness_score') > 8`

## Environment Variables
```env
OPENAI_API_KEY=required          # OpenAI API key
PORT=8080                        # Server port
FRONTEND_URL=http://localhost:3000
```

## Key Patterns
- **Laziness Score**: 1-10 (10 = easiest), auto-calculated
- **Recipe Steps**: Maximum 3 steps for all recipes
- **Cooking Time**: Target < 15 minutes
- **Error Handling**: Comprehensive with retry logic
- **Caching**: 24-hour in-memory cache for API cost reduction
- **Rate Limiting**: 60 requests/minute to OpenAI


## Do Not Touch
- `backend/data/*.db` - Database files
- Migration files (when created)
- Generated binaries in `bin/`