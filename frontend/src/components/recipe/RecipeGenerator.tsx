import React, { useState } from 'react';
import type { GenerateRecipeRequest, LoadingState, Recipe } from '../../types';
import Button from '../common/Button';
import Input from '../common/Input';
import LoadingSpinner from '../common/LoadingSpinner';

interface RecipeGeneratorProps {
  onRecipeGenerated?: (recipe: Recipe) => void;
  onRecipesGenerated?: (recipes: Recipe[]) => void;
  apiEndpoint?: string;
}

const RecipeGenerator: React.FC<RecipeGeneratorProps> = ({
  onRecipeGenerated,
  onRecipesGenerated,
  apiEndpoint = '/api/recipes/generate'
}) => {
  const [formData, setFormData] = useState<GenerateRecipeRequest>({
    ingredients: [],
    cooking_time: 15,
    meal_type: 'dinner',
    dietary_restrictions: [],
    laziness_level: 8
  });

  const [ingredientInput, setIngredientInput] = useState('');
  const [restrictionInput, setRestrictionInput] = useState('');
  const [generateMultiple, setGenerateMultiple] = useState(false);
  const [batchCount, setBatchCount] = useState(3);
  
  const [loadingState, setLoadingState] = useState<LoadingState>({
    isLoading: false,
    error: null
  });

  const mealTypes = [
    { value: 'breakfast', label: '朝食', icon: '🌅' },
    { value: 'lunch', label: '昼食', icon: '☀️' },
    { value: 'dinner', label: '夕食', icon: '🌙' }
  ] as const;

  const commonIngredients = [
    '豚肉', '鶏肉', '牛肉', '卵', '玉ねぎ', 'じゃがいも', '人参', 
    'キャベツ', 'もやし', '豆腐', 'ご飯', 'パン', 'パスタ', 'うどん'
  ];

  const commonRestrictions = [
    'アレルギー対応', 'ベジタリアン', 'ヴィーガン', '低カロリー',
    '低糖質', '高たんぱく', 'グルテンフリー', '減塩'
  ];

  const handleAddIngredient = () => {
    if (ingredientInput.trim() && !formData.ingredients.includes(ingredientInput.trim())) {
      setFormData({
        ...formData,
        ingredients: [...formData.ingredients, ingredientInput.trim()]
      });
      setIngredientInput('');
    }
  };

  const handleRemoveIngredient = (ingredient: string) => {
    setFormData({
      ...formData,
      ingredients: formData.ingredients.filter(i => i !== ingredient)
    });
  };

  const handleAddCommonIngredient = (ingredient: string) => {
    if (!formData.ingredients.includes(ingredient)) {
      setFormData({
        ...formData,
        ingredients: [...formData.ingredients, ingredient]
      });
    }
  };

  const handleAddRestriction = () => {
    if (restrictionInput.trim() && !formData.dietary_restrictions?.includes(restrictionInput.trim())) {
      setFormData({
        ...formData,
        dietary_restrictions: [...(formData.dietary_restrictions || []), restrictionInput.trim()]
      });
      setRestrictionInput('');
    }
  };

  const handleRemoveRestriction = (restriction: string) => {
    setFormData({
      ...formData,
      dietary_restrictions: formData.dietary_restrictions?.filter(r => r !== restriction) || []
    });
  };

  const handleAddCommonRestriction = (restriction: string) => {
    if (!formData.dietary_restrictions?.includes(restriction)) {
      setFormData({
        ...formData,
        dietary_restrictions: [...(formData.dietary_restrictions || []), restriction]
      });
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (formData.ingredients.length === 0) {
      setLoadingState({ isLoading: false, error: '材料を少なくとも1つ追加してください' });
      return;
    }

    setLoadingState({ isLoading: true, error: null });

    try {
      const endpoint = generateMultiple ? '/api/recipes/generate-batch' : apiEndpoint;
      const requestBody = generateMultiple 
        ? { ...formData, count: batchCount }
        : formData;

      const response = await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(requestBody),
      });

      if (!response.ok) {
        throw new Error(`レシピ生成に失敗しました: ${response.status}`);
      }

      const result = await response.json();
      
      if (!result.success) {
        throw new Error(result.error || 'レシピ生成に失敗しました');
      }

      if (generateMultiple && result.data.recipes) {
        onRecipesGenerated?.(result.data.recipes);
      } else if (result.data.recipe) {
        onRecipeGenerated?.(result.data.recipe);
      }

      setLoadingState({ isLoading: false, error: null });
      
    } catch (error) {
      setLoadingState({
        isLoading: false,
        error: error instanceof Error ? error.message : 'レシピ生成中にエラーが発生しました'
      });
    }
  };

  const resetForm = () => {
    setFormData({
      ingredients: [],
      cooking_time: 15,
      meal_type: 'dinner',
      dietary_restrictions: [],
      laziness_level: 8
    });
    setIngredientInput('');
    setRestrictionInput('');
    setLoadingState({ isLoading: false, error: null });
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      {/* ヘッダー */}
      <div className="mb-6">
        <h2 className="text-2xl font-semibold text-gray-900 mb-2">🤖 レシピ生成</h2>
        <p className="text-gray-600">
          材料と条件を設定して、あなた好みの怠けレシピを生成しましょう！
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* 生成モード選択 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">生成モード</label>
          <div className="space-y-2">
            <label className="flex items-center">
              <input
                type="radio"
                checked={!generateMultiple}
                onChange={() => setGenerateMultiple(false)}
                className="mr-2"
              />
              <span>1つのレシピを生成</span>
            </label>
            <label className="flex items-center">
              <input
                type="radio"
                checked={generateMultiple}
                onChange={() => setGenerateMultiple(true)}
                className="mr-2"
              />
              <span>複数のレシピを生成</span>
              {generateMultiple && (
                <select
                  value={batchCount}
                  onChange={(e) => setBatchCount(Number(e.target.value))}
                  className="ml-2 px-2 py-1 border rounded"
                >
                  <option value={3}>3個</option>
                  <option value={5}>5個</option>
                  <option value={10}>10個</option>
                </select>
              )}
            </label>
          </div>
        </div>

        {/* 材料選択 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">
            材料 ({formData.ingredients.length}個選択中)
          </label>
          
          {/* 材料入力 */}
          <div className="flex gap-2 mb-3">
            <Input
              type="text"
              placeholder="材料を入力..."
              value={ingredientInput}
              onChange={(e) => setIngredientInput(e.target.value)}
              onKeyPress={(e: React.KeyboardEvent<HTMLInputElement>) => e.key === 'Enter' && (e.preventDefault(), handleAddIngredient())}
              className="flex-1"
            />
            <Button
              type="button"
              onClick={handleAddIngredient}
              variant="outline"
              disabled={!ingredientInput.trim()}
            >
              追加
            </Button>
          </div>

          {/* よく使う材料 */}
          <div className="mb-3">
            <p className="text-xs text-gray-500 mb-2">よく使う材料:</p>
            <div className="flex flex-wrap gap-2">
              {commonIngredients.map((ingredient) => (
                <button
                  key={ingredient}
                  type="button"
                  onClick={() => handleAddCommonIngredient(ingredient)}
                  disabled={formData.ingredients.includes(ingredient)}
                  className="px-3 py-1 text-xs border rounded-full hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {ingredient}
                </button>
              ))}
            </div>
          </div>

          {/* 選択された材料 */}
          {formData.ingredients.length > 0 && (
            <div className="space-y-2">
              <p className="text-sm text-gray-700">選択された材料:</p>
              <div className="flex flex-wrap gap-2">
                {formData.ingredients.map((ingredient) => (
                  <span
                    key={ingredient}
                    className="inline-flex items-center gap-1 bg-primary-100 text-primary-800 px-3 py-1 rounded-full text-sm"
                  >
                    {ingredient}
                    <button
                      type="button"
                      onClick={() => handleRemoveIngredient(ingredient)}
                      className="text-primary-600 hover:text-primary-800"
                    >
                      ✕
                    </button>
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* 調理時間 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">
            調理時間の上限: {formData.cooking_time}分
          </label>
          <input
            type="range"
            min="5"
            max="60"
            step="5"
            value={formData.cooking_time}
            onChange={(e) => setFormData({ ...formData, cooking_time: Number(e.target.value) })}
            className="w-full"
          />
          <div className="flex justify-between text-xs text-gray-500 mt-1">
            <span>5分</span>
            <span>30分</span>
            <span>60分</span>
          </div>
        </div>

        {/* ラジネスレベル */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">
            怠けレベル: {formData.laziness_level || 8}/10
            <span className="text-xs text-gray-500 ml-2">
              ({(formData.laziness_level || 8) >= 8 ? '超簡単' : (formData.laziness_level || 8) >= 6 ? 'まあまあ' : '少し手間'})
            </span>
          </label>
          <input
            type="range"
            min="1"
            max="10"
            value={formData.laziness_level || 8}
            onChange={(e) => setFormData({ ...formData, laziness_level: Number(e.target.value) })}
            className="w-full"
          />
          <div className="flex justify-between text-xs text-gray-500 mt-1">
            <span>手間をかけたい</span>
            <span>普通</span>
            <span>とことん怠けたい</span>
          </div>
        </div>

        {/* 食事タイプ */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">食事タイプ</label>
          <div className="grid grid-cols-3 gap-3">
            {mealTypes.map((type) => (
              <label
                key={type.value}
                className={`flex items-center justify-center gap-2 p-3 rounded-lg border cursor-pointer transition-colors ${
                  formData.meal_type === type.value
                    ? 'border-primary-500 bg-primary-50 text-primary-700'
                    : 'border-gray-300 hover:bg-gray-50'
                }`}
              >
                <input
                  type="radio"
                  name="meal_type"
                  value={type.value}
                  checked={formData.meal_type === type.value}
                  onChange={(e) => setFormData({ ...formData, meal_type: e.target.value as 'breakfast' | 'lunch' | 'dinner' })}
                  className="sr-only"
                />
                <span className="text-lg">{type.icon}</span>
                <span className="text-sm font-medium">{type.label}</span>
              </label>
            ))}
          </div>
        </div>

        {/* 食事制限 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">食事制限・希望</label>
          
          <div className="flex gap-2 mb-3">
            <Input
              type="text"
              placeholder="制限・希望を入力..."
              value={restrictionInput}
              onChange={(e) => setRestrictionInput(e.target.value)}
              onKeyPress={(e: React.KeyboardEvent<HTMLInputElement>) => e.key === 'Enter' && (e.preventDefault(), handleAddRestriction())}
              className="flex-1"
            />
            <Button
              type="button"
              onClick={handleAddRestriction}
              variant="outline"
              disabled={!restrictionInput.trim()}
            >
              追加
            </Button>
          </div>

          <div className="mb-3">
            <p className="text-xs text-gray-500 mb-2">よくある制限・希望:</p>
            <div className="flex flex-wrap gap-2">
              {commonRestrictions.map((restriction) => (
                <button
                  key={restriction}
                  type="button"
                  onClick={() => handleAddCommonRestriction(restriction)}
                  disabled={formData.dietary_restrictions?.includes(restriction)}
                  className="px-3 py-1 text-xs border rounded-full hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {restriction}
                </button>
              ))}
            </div>
          </div>

          {formData.dietary_restrictions && formData.dietary_restrictions.length > 0 && (
            <div className="space-y-2">
              <p className="text-sm text-gray-700">選択された制限・希望:</p>
              <div className="flex flex-wrap gap-2">
                {formData.dietary_restrictions.map((restriction) => (
                  <span
                    key={restriction}
                    className="inline-flex items-center gap-1 bg-secondary-100 text-secondary-800 px-3 py-1 rounded-full text-sm"
                  >
                    {restriction}
                    <button
                      type="button"
                      onClick={() => handleRemoveRestriction(restriction)}
                      className="text-secondary-600 hover:text-secondary-800"
                    >
                      ✕
                    </button>
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* エラー表示 */}
        {loadingState.error && (
          <div className="bg-red-50 border border-red-200 rounded-lg p-4">
            <p className="text-red-800 text-sm">{loadingState.error}</p>
          </div>
        )}

        {/* アクション */}
        <div className="flex gap-3 pt-4 border-t">
          <Button
            type="submit"
            disabled={loadingState.isLoading || formData.ingredients.length === 0}
            className="flex-1"
          >
            {loadingState.isLoading ? (
              <>
                <LoadingSpinner size="sm" />
                <span className="ml-2">生成中...</span>
              </>
            ) : (
              `🤖 ${generateMultiple ? `${batchCount}個の` : ''}レシピを生成`
            )}
          </Button>
          
          <Button
            type="button"
            variant="outline"
            onClick={resetForm}
            disabled={loadingState.isLoading}
          >
            リセット
          </Button>
        </div>
      </form>
    </div>
  );
};

export default RecipeGenerator;