package services

import (
	"fmt"
	"strings"
)

// RecipeGenerationRequest represents a request to generate recipes
type RecipeGenerationRequest struct {
	Ingredients    []string `json:"ingredients" binding:"required"`
	Season         string   `json:"season" binding:"required,oneof=spring summer fall winter all"`
	MaxCookingTime int      `json:"max_cooking_time" binding:"required,min=1,max=120"`
	Servings       int      `json:"servings,omitempty"`
	Constraints    []string `json:"constraints,omitempty"`
	Preferences    []string `json:"preferences,omitempty"`
}

// PromptTemplate holds template configurations for recipe generation
type PromptTemplate struct {
	SystemPrompt string
	UserPrompt   string
}

// GetRecipeGenerationPrompt creates a prompt for recipe generation
func GetRecipeGenerationPrompt(req RecipeGenerationRequest) PromptTemplate {
	systemPrompt := `あなたはずぼらな人向けのレシピ生成アシスタントです。以下の要件に従って、簡単で失敗しにくいレシピを生成してください。

## 重要な要件
1. 工程は3ステップ以内にする
2. 調理時間は指定された時間以内
3. 洗い物を最小限にする
4. 失敗しにくい調理法を選ぶ
5. 材料を無駄にしない分量で提案する

## レスポンス形式
以下のJSON形式で厳密に出力してください：

{
  "title": "レシピ名",
  "cooking_time": 調理時間（分）,
  "ingredients": [
    {"name": "材料名", "amount": "分量"}
  ],
  "steps": [
    "調理手順1",
    "調理手順2", 
    "調理手順3"
  ],
  "tags": ["簡単", "時短", "ずぼら", "その他のタグ"],
  "season": "適切な季節",
  "laziness_score": 8.5,
  "serving_size": 人数分,
  "difficulty": "easy",
  "total_cost": 推定コスト（円）,
  "nutrition_info": {
    "calories": カロリー,
    "protein": タンパク質（g）
  }
}

## ずぼらスコアの計算基準
- 調理時間5分以内: +3点
- 調理時間10分以内: +2.5点
- 工程2つ以下: +3点
- 工程3つ: +2.5点
- 材料3つ以下: +2点
- 材料5つ以下: +1.5点
- ワンパン/ワンボウル: +2点
- レンジ調理: +2点
- 包丁不要: +1点
- 火を使わない: +1点

最大10点、最小1点でスコアを算出してください。`

	userPrompt := formatUserPrompt(req)

	return PromptTemplate{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	}
}

// formatUserPrompt formats the user prompt based on the request
func formatUserPrompt(req RecipeGenerationRequest) string {
	var prompt strings.Builder

	prompt.WriteString("以下の条件でずぼらな人向けのレシピを1つ生成してください：\n\n")

	// Ingredients
	prompt.WriteString(fmt.Sprintf("## 使用する材料\n%s\n\n", strings.Join(req.Ingredients, ", ")))

	// Season
	seasonMap := map[string]string{
		"spring": "春",
		"summer": "夏",
		"fall":   "秋",
		"winter": "冬",
		"all":    "オールシーズン",
	}
	prompt.WriteString(fmt.Sprintf("## 季節\n%s\n\n", seasonMap[req.Season]))

	// Cooking time
	prompt.WriteString(fmt.Sprintf("## 最大調理時間\n%d分以内\n\n", req.MaxCookingTime))

	// Servings
	servings := req.Servings
	if servings <= 0 {
		servings = 1
	}
	prompt.WriteString(fmt.Sprintf("## 人数\n%d人分\n\n", servings))

	// Constraints
	if len(req.Constraints) > 0 {
		prompt.WriteString(fmt.Sprintf("## 制約条件\n%s\n\n", strings.Join(req.Constraints, ", ")))
	}

	// Preferences
	if len(req.Preferences) > 0 {
		prompt.WriteString(fmt.Sprintf("## 好み・要望\n%s\n\n", strings.Join(req.Preferences, ", ")))
	}

	// Additional instructions
	prompt.WriteString("## 追加指示\n")
	prompt.WriteString("- 手順は簡潔で分かりやすく書く\n")
	prompt.WriteString("- 「適量」「お好みで」などの曖昧な表現は避ける\n")
	prompt.WriteString("- 初心者でも失敗しない具体的な指示を含める\n")
	prompt.WriteString("- 時短テクニックがあれば含める\n")
	prompt.WriteString("- JSON形式での出力を厳守する\n")

	return prompt.String()
}

// GetBatchRecipeGenerationPrompt creates a prompt for generating multiple recipes
func GetBatchRecipeGenerationPrompt(req RecipeGenerationRequest, count int) PromptTemplate {
	systemPrompt := `あなたはずぼらな人向けのレシピ生成アシスタントです。指定された条件で複数のレシピを生成してください。

## 重要な要件
1. 各レシピは工程3ステップ以内
2. 調理時間は指定時間以内
3. バリエーション豊富なレシピを提案
4. 材料の使い回しを考慮
5. 失敗しにくい調理法を選択

## レスポンス形式
以下のJSON配列形式で出力してください：

{
  "recipes": [
    {
      "title": "レシピ名1",
      "cooking_time": 調理時間,
      "ingredients": [{"name": "材料名", "amount": "分量"}],
      "steps": ["手順1", "手順2", "手順3"],
      "tags": ["タグ1", "タグ2"],
      "season": "季節",
      "laziness_score": 8.5,
      "serving_size": 人数,
      "difficulty": "easy",
      "total_cost": コスト,
      "nutrition_info": {"calories": カロリー, "protein": タンパク質}
    }
  ]
}`

	userPrompt := formatBatchUserPrompt(req, count)

	return PromptTemplate{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	}
}

// formatBatchUserPrompt formats the user prompt for batch generation
func formatBatchUserPrompt(req RecipeGenerationRequest, count int) string {
	var prompt strings.Builder

	prompt.WriteString(fmt.Sprintf("以下の条件で%d個のずぼらレシピを生成してください：\n\n", count))

	// Basic conditions
	seasonMap := map[string]string{
		"spring": "春", "summer": "夏", "fall": "秋", "winter": "冬", "all": "オールシーズン",
	}

	prompt.WriteString(fmt.Sprintf("材料: %s\n", strings.Join(req.Ingredients, ", ")))
	prompt.WriteString(fmt.Sprintf("季節: %s\n", seasonMap[req.Season]))
	prompt.WriteString(fmt.Sprintf("最大調理時間: %d分\n", req.MaxCookingTime))

	servings := req.Servings
	if servings <= 0 {
		servings = 1
	}
	prompt.WriteString(fmt.Sprintf("人数: %d人分\n\n", servings))

	// Additional requirements
	prompt.WriteString("## 要求事項\n")
	prompt.WriteString("- 各レシピは調理法が異なること（炒める、煮る、レンジなど）\n")
	prompt.WriteString("- 同じ材料でもバリエーション豊富に\n")
	prompt.WriteString("- すべてのレシピがずぼらスコア7.0以上\n")
	prompt.WriteString("- JSON配列形式での出力を厳守\n")

	if len(req.Constraints) > 0 {
		prompt.WriteString(fmt.Sprintf("\n制約: %s\n", strings.Join(req.Constraints, ", ")))
	}

	return prompt.String()
}

// GetIngredientOptimizationPrompt creates a prompt for optimizing ingredient usage
func GetIngredientOptimizationPrompt(ingredients []string, days int) PromptTemplate {
	systemPrompt := `あなたは食材の使い回しを最適化するアシスタントです。指定された材料を効率的に使い切るレシピプランを提案してください。

## 最適化の目標
1. 材料の無駄を最小限にする
2. 各材料を複数回使用する
3. バランス良い食事を提供
4. 調理時間は短く（15分以内推奨）
5. ずぼらスコア8.0以上を維持

## レスポンス形式
JSON形式で日別のレシピ提案を出力：

{
  "meal_plan": {
    "day1": {"title": "レシピ名", "main_ingredients": ["材料1", "材料2"]},
    "day2": {"title": "レシピ名", "main_ingredients": ["材料2", "材料3"]}
  },
  "ingredient_usage": {
    "材料名": {"total_amount": "総必要量", "usage_days": ["day1", "day3"]}
  },
  "shopping_list": [
    {"item": "材料名", "amount": "購入量", "estimated_cost": コスト}
  ]
}`

	userPrompt := fmt.Sprintf(`以下の材料で%d日分の献立を最適化してください：

材料: %s

要求事項:
- 各材料を最低2回は使用する
- 同じ材料でも異なる調理法を使う
- 保存期間を考慮した使用順序
- 総コストを3000円以内に抑える
- すべてずぼら向けレシピとする`, days, strings.Join(ingredients, ", "))

	return PromptTemplate{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
	}
}
