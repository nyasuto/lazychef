# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

LazyChef is a personal meal planning service that helps users maintain their cooking habits by providing AI-powered weekly shopping lists and recipes optimized for lazy cooks. The core concept is a "buy this, cook for a week" bulk shopping proposal system.

## Development Commands

### Backend (Go)
```bash
# Initialize Go modules
go mod init lazychef
go mod tidy

# Install dependencies
go get -u github.com/gin-gonic/gin
go get -u github.com/mattn/go-sqlite3
go get -u github.com/joho/godotenv

# Run backend server
go run cmd/api/main.go

# Run tests
go test ./...
go test -v -cover ./...

# Build binary
go build -o bin/lazychef cmd/api/main.go
```

### Frontend (React)
```bash
# Install dependencies
npm install

# Start development server
npm start

# Run tests
npm test

# Build for production
npm run build
```

### Database
```bash
# Initialize SQLite database
go run scripts/init_db.go

# Run database migrations
sqlite3 backend/data/recipes.db < scripts/init_db.sql
```

## Architecture Overview

### Tech Stack
- **Backend**: Go with Gin framework, SQLite with JSON features for data storage
- **Frontend**: React with Tailwind CSS for styling
- **AI Integration**: OpenAI API for recipe generation
- **API Communication**: RESTful JSON APIs with CORS enabled for localhost:3000

### Core Components

1. **Recipe Generation Service** (`backend/internal/services/generator.go`)
   - Integrates with OpenAI API to generate recipes
   - Optimizes for "laziness score" (quick, simple recipes)
   - Considers seasonal ingredients and cooking time constraints

2. **Weekly Planner Service** (`backend/internal/services/planner.go`)
   - Creates weekly meal plans with ingredient reuse optimization
   - Generates consolidated shopping lists
   - Estimates total costs

3. **Database Layer** (`backend/internal/database/sqlite.go`)
   - SQLite with JSON columns for flexible schema
   - Three main tables: recipes, meal_plans, user_preferences
   - All recipe and plan data stored as JSON for easy querying

4. **API Handlers** (`backend/internal/handlers/`)
   - Recipe generation endpoint: POST /api/recipes/generate
   - Meal plan creation: POST /api/meal-plans/create
   - Recipe search: GET /api/recipes/search

## Key API Endpoints

```
POST /api/recipes/generate
- Body: { ingredients: [], season: string, max_cooking_time: number }
- Returns: Generated recipes with laziness scores

POST /api/meal-plans/create
- Body: { start_date: string, preferences: object }
- Returns: Weekly plan with shopping list and daily recipes

GET /api/recipes/search
- Query params: tag, ingredient
- Returns: Filtered recipe list
```

## Development Guidelines

### Working with the Database
- All recipe data is stored as JSON in SQLite
- Use JSON queries for filtering: `SELECT * FROM recipes WHERE json_extract(data, '$.laziness_score') > 8`
- Keep laziness_score between 1-10 (10 = easiest)

### Frontend Components Structure
```
components/
├── RecipeCard.jsx      # Individual recipe display
├── WeeklyPlan.jsx      # 7-day meal plan grid
├── ShoppingList.jsx    # Consolidated ingredients list
└── RecipeGenerator.jsx # AI recipe generation form
```

### Environment Variables
Required in `.env`:
```
OPENAI_API_KEY=your_api_key_here
PORT=8080
FRONTEND_URL=http://localhost:3000
```

### CORS Configuration
Backend allows requests from `http://localhost:3000`. Middleware is in `backend/internal/middleware/cors.go`.

## Data Models

### Recipe JSON Structure
- title: string (recipe name)
- cooking_time: number (minutes)
- ingredients: array of {name, amount}
- steps: array of strings (simplified instructions)
- tags: array of strings
- laziness_score: number (1-10, higher = easier)
- season: string (spring/summer/fall/winter/all)

### Meal Plan JSON Structure
- start_date: ISO date string
- shopping_list: array of {item, amount}
- daily_recipes: object mapping day to recipe
- total_cost_estimate: number (yen)

## Testing Guidelines

- Backend: Use Go's built-in testing with coverage reports
- Frontend: Jest for React component testing
- API: Test CORS headers and JSON responses
- Database: Test JSON queries and data integrity

## Performance Targets
- Recipe generation: < 3 seconds
- Weekly plan creation: < 5 seconds
- Database queries: < 100ms
- Initial page load: < 2 seconds