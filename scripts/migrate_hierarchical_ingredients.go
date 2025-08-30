package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// データベースファイルのパス
	dbPath := filepath.Join("..", "data", "recipes.db")
	if len(os.Args) > 1 {
		dbPath = os.Args[1]
	}

	// データベース接続
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("データベースオープンエラー: %v", err)
	}
	defer db.Close()

	// 外部キー制約を有効化
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatalf("外部キー制約有効化エラー: %v", err)
	}

	log.Println("=== 階層的材料分類システム マイグレーション開始 ===")

	// スキーマファイルを読み込み
	schemaPath := "hierarchical_ingredients_schema.sql"
	schemaContent, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Fatalf("スキーマファイル読み込みエラー: %v", err)
	}

	// トランザクション開始
	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("トランザクション開始エラー: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			log.Println("マイグレーション失敗: ロールバック実行")
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				log.Fatalf("コミットエラー: %v", commitErr)
			}
			log.Println("マイグレーション成功: コミット完了")
		}
	}()

	// 既存テーブルの存在確認
	var tableExists bool
	err = tx.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM sqlite_master 
		WHERE type='table' AND name='ingredient_groups'
	`).Scan(&tableExists)
	if err != nil {
		log.Fatalf("テーブル存在確認エラー: %v", err)
		return
	}

	if tableExists {
		log.Println("階層的材料テーブルが既に存在します。スキップします。")
		log.Println("完全な再作成が必要な場合は、先にテーブルを削除してください：")
		log.Println("  DROP TABLE IF EXISTS ingredient_group_mappings;")
		log.Println("  DROP TABLE IF EXISTS specific_ingredients;")
		log.Println("  DROP TABLE IF EXISTS ingredient_groups;")
		return
	}

	// スキーマ実行
	log.Println("1. 階層的材料テーブル作成中...")
	if _, err = tx.Exec(string(schemaContent)); err != nil {
		log.Fatalf("スキーマ実行エラー: %v", err)
		return
	}
	log.Println("   ✓ テーブル作成完了")

	// データ検証
	log.Println("2. データ整合性確認中...")
	
	// 材料グループ数確認
	var groupCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM ingredient_groups").Scan(&groupCount)
	if err != nil {
		log.Fatalf("グループ数確認エラー: %v", err)
		return
	}
	log.Printf("   ✓ 材料グループ: %d件", groupCount)

	// 具体的材料数確認
	var ingredientCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM specific_ingredients").Scan(&ingredientCount)
	if err != nil {
		log.Fatalf("材料数確認エラー: %v", err)
		return
	}
	log.Printf("   ✓ 具体的材料: %d件", ingredientCount)

	// マッピング数確認
	var mappingCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM ingredient_group_mappings").Scan(&mappingCount)
	if err != nil {
		log.Fatalf("マッピング数確認エラー: %v", err)
		return
	}
	log.Printf("   ✓ 材料マッピング: %d件", mappingCount)

	// 階層構造確認（大分類）
	var level1Count int
	err = tx.QueryRow("SELECT COUNT(*) FROM ingredient_groups WHERE level = 1").Scan(&level1Count)
	if err != nil {
		log.Fatalf("大分類確認エラー: %v", err)
		return
	}
	log.Printf("   ✓ 大分類: %d件", level1Count)

	// 階層構造確認（中分類）
	var level2Count int
	err = tx.QueryRow("SELECT COUNT(*) FROM ingredient_groups WHERE level = 2").Scan(&level2Count)
	if err != nil {
		log.Fatalf("中分類確認エラー: %v", err)
		return
	}
	log.Printf("   ✓ 中分類: %d件", level2Count)

	log.Println("3. 検索機能テスト...")
	
	// テスト: 「肉類」で検索して具体的な材料を取得
	rows, err := tx.Query(`
		SELECT DISTINCT si.name, si.display_name
		FROM specific_ingredients si
		JOIN ingredient_group_mappings igm ON si.id = igm.ingredient_id
		JOIN ingredient_groups ig ON igm.group_id = ig.id
		WHERE ig.name = 'meat' OR ig.display_name = '肉類'
		ORDER BY si.name
	`)
	if err != nil {
		log.Fatalf("検索テストエラー: %v", err)
		return
	}
	defer rows.Close()

	log.Println("   「肉類」で検索結果:")
	var meatIngredients []string
	for rows.Next() {
		var name, displayName string
		if err := rows.Scan(&name, &displayName); err != nil {
			log.Fatalf("検索結果取得エラー: %v", err)
			return
		}
		meatIngredients = append(meatIngredients, name)
		log.Printf("     - %s (%s)", name, displayName)
	}
	if len(meatIngredients) == 0 {
		log.Fatalf("肉類の検索結果が0件です。データに問題があります。")
		return
	}

	// テスト: 「鶏肉」で検索して具体的な材料を取得
	rows2, err := tx.Query(`
		SELECT DISTINCT si.name, si.display_name
		FROM specific_ingredients si
		JOIN ingredient_group_mappings igm ON si.id = igm.ingredient_id
		JOIN ingredient_groups ig ON igm.group_id = ig.id
		WHERE ig.name = 'chicken' OR ig.display_name = '鶏肉'
		ORDER BY si.name
	`)
	if err != nil {
		log.Fatalf("鶏肉検索テストエラー: %v", err)
		return
	}
	defer rows2.Close()

	log.Println("   「鶏肉」で検索結果:")
	var chickenIngredients []string
	for rows2.Next() {
		var name, displayName string
		if err := rows2.Scan(&name, &displayName); err != nil {
			log.Fatalf("鶏肉検索結果取得エラー: %v", err)
			return
		}
		chickenIngredients = append(chickenIngredients, name)
		log.Printf("     - %s (%s)", name, displayName)
	}

	log.Println("4. レシピ検索互換性テスト...")
	
	// 既存のレシピデータとの互換性確認
	var recipeCount int
	err = tx.QueryRow("SELECT COUNT(*) FROM recipes").Scan(&recipeCount)
	if err != nil {
		log.Fatalf("レシピ数確認エラー: %v", err)
		return
	}
	log.Printf("   ✓ 既存レシピ数: %d件", recipeCount)

	// 実際のレシピ材料と新システムの材料マッチング確認
	matchQuery := `
		SELECT COUNT(DISTINCT recipe_id) as matched_recipes
		FROM (
			SELECT r.id as recipe_id, json_extract(ingredient.value, '$.name') as ingredient_name
			FROM recipes r, json_each(json_extract(r.data, '$.ingredients')) as ingredient
		) recipe_ingredients
		JOIN specific_ingredients si ON recipe_ingredients.ingredient_name = si.name
	`
	var matchedRecipes int
	err = tx.QueryRow(matchQuery).Scan(&matchedRecipes)
	if err != nil {
		log.Fatalf("レシピマッチング確認エラー: %v", err)
		return
	}
	log.Printf("   ✓ 新システムと互換性のあるレシピ: %d件", matchedRecipes)

	if recipeCount > 0 && matchedRecipes == 0 {
		log.Println("   ⚠️  警告: 既存レシピと新材料システムの間にマッチする材料がありません")
		log.Println("           追加の材料データ登録が必要な可能性があります")
	}

	log.Println("=== マイグレーション完了 ===")
	log.Println("")
	log.Println("次のステップ:")
	log.Println("1. バックエンドAPI (/api/recipes/search) の更新")
	log.Println("2. フロントエンド材料検索UIの改善")
	log.Println("3. 既存レシピで使用されているが新システムに未登録の材料の追加")
	log.Println("")
	log.Println("利用可能なクエリ例:")
	log.Println(`  -- 「肉類」に属するレシピ検索:`)
	log.Println(`  SELECT DISTINCT r.* FROM recipes r, json_each(json_extract(r.data, '$.ingredients')) as ingredient`)
	log.Println(`  JOIN specific_ingredients si ON json_extract(ingredient.value, '$.name') = si.name`)
	log.Println(`  JOIN ingredient_group_mappings igm ON si.id = igm.ingredient_id`)
	log.Println(`  JOIN ingredient_groups ig ON igm.group_id = ig.id`)
	log.Println(`  WHERE ig.display_name = '肉類';`)

	err = nil // マイグレーション成功
}