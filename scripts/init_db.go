package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

const (
	dbPath = "../backend/data/recipes.db"
	schemaFile = "init_db.sql"
)

func main() {
	fmt.Println("🗄️ Initializing LazyChef Database...")
	
	// Create data directory if it doesn't exist
	dataDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}
	
	// Remove existing database file
	if _, err := os.Stat(dbPath); err == nil {
		fmt.Println("📤 Removing existing database...")
		if err := os.Remove(dbPath); err != nil {
			log.Fatalf("Failed to remove existing database: %v", err)
		}
	}
	
	// Create new database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()
	
	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	fmt.Println("✅ Database connection established")
	
	// Read and execute schema
	schemaSQL, err := ioutil.ReadFile(schemaFile)
	if err != nil {
		log.Fatalf("Failed to read schema file: %v", err)
	}
	
	fmt.Println("📋 Creating database schema...")
	if _, err := db.Exec(string(schemaSQL)); err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}
	
	fmt.Println("✅ Schema created successfully")
	
	// Insert sample data
	fmt.Println("📝 Inserting sample recipes...")
	if err := insertSampleRecipes(db); err != nil {
		log.Fatalf("Failed to insert sample recipes: %v", err)
	}
	
	fmt.Println("📅 Inserting sample meal plan...")
	if err := insertSampleMealPlan(db); err != nil {
		log.Fatalf("Failed to insert sample meal plan: %v", err)
	}
	
	// Verify data
	fmt.Println("🔍 Verifying database...")
	if err := verifyDatabase(db); err != nil {
		log.Fatalf("Database verification failed: %v", err)
	}
	
	fmt.Println("🎉 Database initialization completed successfully!")
	fmt.Printf("📊 Database created at: %s\n", dbPath)
}

func insertSampleRecipes(db *sql.DB) error {
	recipes := []string{
		`{
			"title": "10分豚キャベツ炒め",
			"cooking_time": 10,
			"ingredients": [
				{"name": "豚こま肉", "amount": "200g"},
				{"name": "キャベツ", "amount": "1/4個"},
				{"name": "醤油", "amount": "大さじ1"},
				{"name": "塩胡椒", "amount": "少々"}
			],
			"steps": [
				"キャベツをざく切りする",
				"豚肉を炒めて、キャベツを入れる",
				"醤油と塩胡椒で味付けして完成"
			],
			"tags": ["簡単", "豚肉", "10分以内", "ずぼら"],
			"season": "all",
			"laziness_score": 9.5,
			"serving_size": 1,
			"difficulty": "easy",
			"total_cost": 250,
			"nutrition_info": {
				"calories": 280,
				"protein": 22
			}
		}`,
		`{
			"title": "もやし卵とじ",
			"cooking_time": 5,
			"ingredients": [
				{"name": "もやし", "amount": "1袋"},
				{"name": "卵", "amount": "2個"},
				{"name": "醤油", "amount": "小さじ1"},
				{"name": "ごま油", "amount": "小さじ1"}
			],
			"steps": [
				"もやしを洗う",
				"フライパンでもやしを炒める",
				"溶き卵を入れて炒め合わせる",
				"醤油とごま油で味付け"
			],
			"tags": ["簡単", "卵", "5分以内", "ずぼら", "安い"],
			"season": "all",
			"laziness_score": 9.8,
			"serving_size": 1,
			"difficulty": "easy",
			"total_cost": 120,
			"nutrition_info": {
				"calories": 180,
				"protein": 14
			}
		}`,
		`{
			"title": "キャベツの味噌汁",
			"cooking_time": 8,
			"ingredients": [
				{"name": "キャベツ", "amount": "2枚"},
				{"name": "味噌", "amount": "大さじ1"},
				{"name": "だしの素", "amount": "小さじ1/2"},
				{"name": "水", "amount": "400ml"}
			],
			"steps": [
				"キャベツを適当に切る",
				"水を沸騰させてだしの素を入れる",
				"キャベツを入れて2分煮る",
				"味噌を溶いて完成"
			],
			"tags": ["汁物", "キャベツ", "10分以内", "和食"],
			"season": "all",
			"laziness_score": 8.5,
			"serving_size": 1,
			"difficulty": "easy",
			"total_cost": 80,
			"nutrition_info": {
				"calories": 45,
				"protein": 3
			}
		}`,
		`{
			"title": "豚こまチャーハン",
			"cooking_time": 12,
			"ingredients": [
				{"name": "豚こま肉", "amount": "150g"},
				{"name": "ご飯", "amount": "1膳分"},
				{"name": "卵", "amount": "1個"},
				{"name": "長ねぎ", "amount": "1/2本"},
				{"name": "醤油", "amount": "大さじ1"},
				{"name": "ごま油", "amount": "大さじ1"}
			],
			"steps": [
				"豚こま肉を炒める",
				"溶き卵を入れて炒める",
				"ご飯を入れて炒める",
				"ねぎと調味料を入れて完成"
			],
			"tags": ["チャーハン", "豚肉", "15分以内", "ワンパン"],
			"season": "all",
			"laziness_score": 8.0,
			"serving_size": 1,
			"difficulty": "easy",
			"total_cost": 300,
			"nutrition_info": {
				"calories": 520,
				"protein": 25
			}
		}`,
		`{
			"title": "レンジでナスの蒸し物",
			"cooking_time": 6,
			"ingredients": [
				{"name": "なす", "amount": "2本"},
				{"name": "ポン酢", "amount": "大さじ1"},
				{"name": "ごま油", "amount": "小さじ1"},
				{"name": "刻みねぎ", "amount": "適量"}
			],
			"steps": [
				"なすを適当に切る",
				"耐熱皿に入れラップをして5分レンチン",
				"ポン酢とごま油をかける",
				"刻みねぎをのせて完成"
			],
			"tags": ["レンジ", "なす", "火を使わない", "さっぱり", "夏"],
			"season": "summer",
			"laziness_score": 9.7,
			"serving_size": 1,
			"difficulty": "easy",
			"total_cost": 150,
			"nutrition_info": {
				"calories": 90,
				"protein": 2
			}
		}`,
	}
	
	for i, recipeJSON := range recipes {
		_, err := db.Exec("INSERT INTO recipes (data) VALUES (?)", recipeJSON)
		if err != nil {
			return fmt.Errorf("failed to insert recipe %d: %v", i+1, err)
		}
	}
	
	fmt.Printf("✅ Inserted %d sample recipes\n", len(recipes))
	return nil
}

func insertSampleMealPlan(db *sql.DB) error {
	mealPlanJSON := `{
		"start_date": "2025-01-27",
		"shopping_list": [
			{"item": "豚こま肉", "amount": "350g", "cost": 450, "category": "meat"},
			{"item": "キャベツ", "amount": "1個", "cost": 150, "category": "vegetable"},
			{"item": "もやし", "amount": "3袋", "cost": 90, "category": "vegetable"},
			{"item": "卵", "amount": "1パック", "cost": 200, "category": "dairy"},
			{"item": "なす", "amount": "2本", "cost": 120, "category": "vegetable"},
			{"item": "長ねぎ", "amount": "1本", "cost": 80, "category": "vegetable"},
			{"item": "醤油", "amount": "1本", "cost": 180, "category": "seasoning"},
			{"item": "味噌", "amount": "1パック", "cost": 200, "category": "seasoning"},
			{"item": "ごま油", "amount": "1本", "cost": 300, "category": "seasoning"}
		],
		"daily_recipes": {
			"monday": {"recipe_id": 1, "title": "10分豚キャベツ炒め"},
			"tuesday": {"recipe_id": 2, "title": "もやし卵とじ"},
			"wednesday": {"recipe_id": 4, "title": "豚こまチャーハン"},
			"thursday": {"recipe_id": 3, "title": "キャベツの味噌汁"},
			"friday": {"recipe_id": 5, "title": "レンジでナスの蒸し物"}
		},
		"total_cost_estimate": 1770,
		"week_theme": "基本のずぼら飯",
		"ingredient_reuse": {
			"豚こま肉": ["monday", "wednesday"],
			"キャベツ": ["monday", "thursday"],
			"卵": ["tuesday", "wednesday"],
			"醤油": ["monday", "wednesday", "thursday"]
		},
		"nutrition_summary": {
			"total_calories": 1115,
			"avg_calories_per_day": 223,
			"total_protein": 66,
			"balance_score": 7.5
		}
	}`
	
	_, err := db.Exec("INSERT INTO meal_plans (week_data) VALUES (?)", mealPlanJSON)
	if err != nil {
		return fmt.Errorf("failed to insert meal plan: %v", err)
	}
	
	fmt.Println("✅ Inserted sample meal plan")
	return nil
}

func verifyDatabase(db *sql.DB) error {
	// Check recipes table
	var recipeCount int
	err := db.QueryRow("SELECT COUNT(*) FROM recipes").Scan(&recipeCount)
	if err != nil {
		return fmt.Errorf("failed to count recipes: %v", err)
	}
	fmt.Printf("📊 Recipes table: %d records\n", recipeCount)
	
	// Check meal_plans table
	var mealPlanCount int
	err = db.QueryRow("SELECT COUNT(*) FROM meal_plans").Scan(&mealPlanCount)
	if err != nil {
		return fmt.Errorf("failed to count meal plans: %v", err)
	}
	fmt.Printf("📊 Meal plans table: %d records\n", mealPlanCount)
	
	// Check user_preferences table
	var userPrefCount int
	err = db.QueryRow("SELECT COUNT(*) FROM user_preferences").Scan(&userPrefCount)
	if err != nil {
		return fmt.Errorf("failed to count user preferences: %v", err)
	}
	fmt.Printf("📊 User preferences table: %d records\n", userPrefCount)
	
	// Test JSON queries
	var title string
	err = db.QueryRow("SELECT title FROM recipes WHERE laziness_score > 9.0 LIMIT 1").Scan(&title)
	if err != nil {
		return fmt.Errorf("failed to test JSON query: %v", err)
	}
	fmt.Printf("🔍 JSON query test: Found recipe '%s'\n", title)
	
	return nil
}