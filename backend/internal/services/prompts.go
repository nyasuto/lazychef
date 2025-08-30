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
	systemPrompt := `あなたはずぼらな人向けのレシピ生成アシスタントです。以下の新しい「ずぼら制約システム」に従って、バリエーション豊富で失敗しにくいレシピを生成してください。

## Step-Effort Budget制（工程制約の革新）
工程数ではなく「手間の総量」で制限します。各ステップに手間度を割り当て、合計8点以内：

**手間度レベル:**
- とても簡単 (1点): 材料を入れる、混ぜる、放置する
- 簡単 (2点): 切る、炒める、茹でる、レンジ加熱
- やや複雑 (3点): タイミング調整、火加減調整、複数工程の同時進行

**例:** 
- 3ステップ炒め物: 2+3+3=8点 ✅
- 5ステップ煮込み: 1+1+1+1+1=5点 ✅

## 調理時間4段階システム
手間時間と総時間を明確に分離：

1. **Lightning (稲妻)**: 手間≤10分/総時間≤15分 - 炒め物、レンジ料理
2. **Quick (クイック)**: 手間≤15分/総時間≤30分 - 焼き物、軽い煮込み
3. **Hands-off (放置)**: 手間≤15分/総時間≤90分 - 長時間煮込み、オーブン
4. **Set-and-forget (完全放置)**: 手間≤15分/総時間≤8時間 - スロークッカー、発酵

## レスポンス形式
以下のJSON形式で厳密に出力してください：

{
  "title": "レシピ名",
  "active_time": 手間時間（分）,
  "total_time": 総調理時間（分）,
  "time_tier": "Lightning|Quick|Hands-off|Set-and-forget",
  "ingredients": [
    {"name": "材料名", "amount": "分量"}
  ],
  "steps": [
    {"instruction": "調理手順1", "effort_level": 1, "description": "とても簡単"},
    {"instruction": "調理手順2", "effort_level": 2, "description": "簡単"}
  ],
  "step_effort_total": 5,
  "tags": ["簡単", "時短", "ずぼら", "その他のタグ"],
  "season": "適切な季節",
  "laziness_score": 85,
  "laziness_justification": "手間が少なく洗い物も最小限のため",
  "serving_size": 1,
  "difficulty": "easy",
  "total_cost": 推定コスト（円）,
  "nutrition_info": {
    "calories": カロリー,
    "protein": タンパク質（g）
  }
}

## 新ずぼらスコア計算システム（0-100点）
**重み付け評価:**
- 手間時間 (40%): 手間が少ないほど高得点
- 工程複雑度 (30%): Step-Effort合計が少ないほど高得点  
- 材料調達性 (20%): 一般的なスーパーで買える材料ほど高得点
- 洗い物負荷 (10%): 使用する調理器具・皿が少ないほど高得点

**合格基準:**
- 70点以上: ずぼら認定
- 50-69点: 改善余地あり
- 50点未満: 再生成が必要

必ず70点以上のレシピを生成してください。`

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
	prompt.WriteString("- 各ステップに手間度(1-3)を明記し、合計8点以内にする\n")
	prompt.WriteString("- 手間時間と総時間を明確に分けて記載\n")
	prompt.WriteString("- 適切な時間ティア(Lightning/Quick/Hands-off/Set-and-forget)を選択\n")
	prompt.WriteString("- ずぼらスコア70点以上を必ず達成する\n")
	prompt.WriteString("- 「適量」「お好みで」などの曖昧な表現は避ける\n")
	prompt.WriteString("- 初心者でも失敗しない具体的な指示を含める\n")
	prompt.WriteString("- 新JSON形式での出力を厳守する\n")

	return prompt.String()
}

// GetBatchRecipeGenerationPrompt creates a prompt for generating multiple recipes
func GetBatchRecipeGenerationPrompt(req RecipeGenerationRequest, count int) PromptTemplate {
	systemPrompt := `あなたはずぼらな人向けのレシピ生成アシスタントです。新しい「ずぼら制約システム」に従って、バリエーション豊富な複数のレシピを生成してください。

## Step-Effort Budget制 + 時間ティアシステム
- 各レシピの手間度合計を8点以内に制限
- Lightning/Quick/Hands-off/Set-and-forgetの4つの時間ティアを活用
- 異なる時間ティアの組み合わせでバリエーション創出

## 重要な要件
1. 各レシピはStep-Effort Budget 8点以内
2. 複数の時間ティアを組み合わせる（例：Lightning×2, Hands-off×1）
3. 同じ材料でも異なる調理法（炒める/煮込む/レンジ/オーブン）
4. 全レシピがずぼらスコア70点以上
5. 手間時間と総時間を明確に分離

## レスポンス形式
以下のJSON配列形式で出力してください：

{
  "recipes": [
    {
      "title": "レシピ名1",
      "active_time": 手間時間（分）,
      "total_time": 総時間（分）,
      "time_tier": "Lightning|Quick|Hands-off|Set-and-forget",
      "ingredients": [{"name": "材料名", "amount": "分量"}],
      "steps": [
        {"instruction": "手順1", "effort_level": 1, "description": "とても簡単"},
        {"instruction": "手順2", "effort_level": 2, "description": "簡単"}
      ],
      "step_effort_total": 5,
      "tags": ["タグ1", "タグ2"],
      "season": "季節",
      "laziness_score": 85,
      "laziness_justification": "スコア理由",
      "serving_size": 1,
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
	prompt.WriteString("- 各レシピの手間度合計を8点以内に制限\n")
	prompt.WriteString("- 複数の時間ティア（Lightning/Quick/Hands-off/Set-and-forget）を組み合わせる\n")
	prompt.WriteString("- 各レシピは調理法が異なること（炒める、煮る、レンジ、オーブンなど）\n")
	prompt.WriteString("- 同じ材料でもバリエーション豊富に\n")
	prompt.WriteString("- すべてのレシピがずぼらスコア70点以上\n")
	prompt.WriteString("- 新JSON配列形式での出力を厳守\n")

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
