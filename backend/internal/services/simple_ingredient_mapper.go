package services

import (
	"strings"
)

// SimpleIngredientMapper provides basic ingredient synonym mapping
// This is Phase 1 - simple dictionary-based mapping
// Phase 2 will add AI-powered embeddings and classification
type SimpleIngredientMapper struct {
	synonymMap map[string][]string
}

// NewSimpleIngredientMapper creates a new ingredient mapper with predefined synonyms
func NewSimpleIngredientMapper() *SimpleIngredientMapper {
	mapper := &SimpleIngredientMapper{
		synonymMap: make(map[string][]string),
	}

	// Initialize with common Japanese ingredient synonyms
	mapper.initializeSynonyms()

	return mapper
}

// initializeSynonyms sets up the synonym dictionary
// This addresses the immediate problem: UI searches "鶏肉" but DB has "鶏胸肉"
func (m *SimpleIngredientMapper) initializeSynonyms() {
	// Meat categories - the main issue from Issue #87
	m.synonymMap["肉類"] = []string{
		"牛切り落とし", "豚こま肉", "鶏胸肉", "牛肉", "豚肉", "鶏肉",
	}
	m.synonymMap["牛肉"] = []string{
		"牛切り落とし", "牛こま肉", "牛肉切り落とし",
	}
	m.synonymMap["豚肉"] = []string{
		"豚こま肉", "豚切り落とし", "豚肉切り落とし",
	}
	m.synonymMap["鶏肉"] = []string{
		"鶏胸肉", "鶏むね肉", "チキンブレスト", "鶏もも肉", "鶏ささみ",
	}

	// Vegetables
	m.synonymMap["野菜"] = []string{
		"玉ねぎ", "人参", "じゃがいも", "もやし", "キャベツ", "白菜",
		"レタス", "きゅうり", "トマト", "ピーマン", "ねぎ", "にんにく",
		"ニンジン", "ジャガイモ", "タマネギ",
	}
	m.synonymMap["玉ねぎ"] = []string{
		"玉ねぎ", "タマネギ", "オニオン", "たまねぎ",
	}
	m.synonymMap["人参"] = []string{
		"人参", "にんじん", "ニンジン", "キャロット",
	}
	m.synonymMap["じゃがいも"] = []string{
		"じゃがいも", "ジャガイモ", "ポテト", "馬鈴薯",
	}

	// Seafood
	m.synonymMap["魚介類"] = []string{
		"鮭", "ツナ缶", "わかめ", "サケ", "サーモン", "まぐろ缶", "ツナ",
	}
	m.synonymMap["魚"] = []string{
		"鮭", "サケ", "サーモン", "まぐろ", "マグロ",
	}

	// Grains and noodles
	m.synonymMap["穀物"] = []string{
		"ご飯", "パスタ", "うどん", "米", "白米", "ライス", "スパゲッティ",
	}
	m.synonymMap["麺類"] = []string{
		"パスタ", "うどん", "スパゲッティ", "スパゲティ", "ラーメン", "そば", "そうめん",
	}

	// Dairy and eggs
	m.synonymMap["卵"] = []string{
		"卵", "たまご", "玉子", "エッグ",
	}
	m.synonymMap["豆腐"] = []string{
		"豆腐", "とうふ", "トウフ", "木綿豆腐", "絹豆腐",
	}

	// Seasonings
	m.synonymMap["調味料"] = []string{
		"塩", "胡椒", "しょうゆ", "味噌", "砂糖", "みりん", "酒",
		"サラダ油", "ごま油", "醤油", "こしょう", "コショウ",
	}
	m.synonymMap["塩"] = []string{
		"塩", "しお", "食塩", "海塩",
	}
	m.synonymMap["胡椒"] = []string{
		"胡椒", "こしょう", "コショウ", "ペッパー", "黒胡椒", "白胡椒",
	}
	m.synonymMap["しょうゆ"] = []string{
		"しょうゆ", "醤油", "ショウユ", "しょう油", "濃口醤油", "薄口醤油",
	}
	m.synonymMap["味噌"] = []string{
		"味噌", "みそ", "ミソ", "赤味噌", "白味噌", "合わせ味噌",
	}
	m.synonymMap["油"] = []string{
		"サラダ油", "ごま油", "オリーブオイル", "植物油", "ゴマ油", "胡麻油",
	}

	// Common aliases and variations
	m.addBidirectionalMappings()
}

// addBidirectionalMappings ensures that if A maps to B, then B also maps to A
func (m *SimpleIngredientMapper) addBidirectionalMappings() {
	// Create reverse mappings for better matching
	// For example, if someone searches for "鶏胸肉", it should also match "鶏肉" searches
	additionalMappings := make(map[string][]string)

	for category, ingredients := range m.synonymMap {
		for _, ingredient := range ingredients {
			// Add the category to each ingredient's synonym list
			if additionalMappings[ingredient] == nil {
				additionalMappings[ingredient] = []string{}
			}
			additionalMappings[ingredient] = append(additionalMappings[ingredient], category)

			// Add other ingredients in the same category as synonyms
			for _, sibling := range ingredients {
				if sibling != ingredient {
					additionalMappings[ingredient] = append(additionalMappings[ingredient], sibling)
				}
			}
		}
	}

	// Merge additional mappings
	for key, values := range additionalMappings {
		if existingValues, exists := m.synonymMap[key]; exists {
			m.synonymMap[key] = append(existingValues, values...)
		} else {
			m.synonymMap[key] = values
		}
	}

	// Remove duplicates
	for key, values := range m.synonymMap {
		m.synonymMap[key] = removeDuplicateStrings(values)
	}
}

// ExpandIngredientTerms takes a user search term and returns all possible ingredient names
// This is the core function that solves Issue #87
func (m *SimpleIngredientMapper) ExpandIngredientTerms(searchTerms []string) []string {
	var allIngredients []string

	for _, term := range searchTerms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}

		// Add the original term
		allIngredients = append(allIngredients, term)

		// Add synonyms if they exist
		if synonyms, exists := m.synonymMap[term]; exists {
			allIngredients = append(allIngredients, synonyms...)
		}

		// Also check for partial matches (case-insensitive)
		for key, synonyms := range m.synonymMap {
			if strings.Contains(strings.ToLower(key), strings.ToLower(term)) ||
				strings.Contains(strings.ToLower(term), strings.ToLower(key)) {
				allIngredients = append(allIngredients, synonyms...)
			}
		}
	}

	// Remove duplicates and return
	return removeDuplicateStrings(allIngredients)
}

// GetSupportedCategories returns all top-level categories for UI
func (m *SimpleIngredientMapper) GetSupportedCategories() []string {
	categories := []string{
		"肉類", "野菜", "魚介類", "穀物", "麺類", "調味料",
	}
	return categories
}

// GetCategoryIngredients returns all ingredients in a specific category
func (m *SimpleIngredientMapper) GetCategoryIngredients(category string) []string {
	if ingredients, exists := m.synonymMap[category]; exists {
		return ingredients
	}
	return []string{}
}

// AddSynonym allows dynamic addition of synonyms (for future AI integration)
func (m *SimpleIngredientMapper) AddSynonym(term string, synonyms []string) {
	if existingSynonyms, exists := m.synonymMap[term]; exists {
		m.synonymMap[term] = append(existingSynonyms, synonyms...)
		m.synonymMap[term] = removeDuplicateStrings(m.synonymMap[term])
	} else {
		m.synonymMap[term] = synonyms
	}
}

// Helper function to remove duplicate strings
func removeDuplicateStrings(slice []string) []string {
	keys := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if item != "" && !keys[item] {
			keys[item] = true
			result = append(result, item)
		}
	}

	return result
}
