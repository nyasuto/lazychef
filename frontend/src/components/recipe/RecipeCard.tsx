import React from 'react';
import type { Recipe } from '../../types';

interface RecipeCardProps {
  recipe: Recipe;
  onClick?: (recipe: Recipe) => void;
}

const RecipeCard: React.FC<RecipeCardProps> = ({ recipe, onClick }) => {
  const handleClick = () => {
    if (onClick) {
      onClick(recipe);
    }
  };

  const getLazinessScoreColor = (score: number) => {
    if (score >= 8) return 'bg-green-100 text-green-800';
    if (score >= 6) return 'bg-yellow-100 text-yellow-800';
    return 'bg-red-100 text-red-800';
  };

  const getLazinessScoreText = (score: number) => {
    if (score >= 8) return 'とっても簡単';
    if (score >= 6) return 'まあまあ簡単';
    return 'ちょっと難しい';
  };

  return (
    <div 
      className={`card hover:shadow-lg transition-shadow duration-200 ${onClick ? 'cursor-pointer' : ''}`}
      onClick={handleClick}
    >
      {/* ヘッダー */}
      <div className="flex justify-between items-start mb-3">
        <h3 className="text-lg font-semibold text-gray-900 line-clamp-2">
          {recipe.title}
        </h3>
        <div className={`px-2 py-1 rounded-full text-xs font-medium whitespace-nowrap ml-2 ${getLazinessScoreColor(recipe.laziness_score)}`}>
          {getLazinessScoreText(recipe.laziness_score)}
        </div>
      </div>

      {/* 調理時間・材料数 */}
      <div className="flex items-center gap-4 mb-3 text-sm text-gray-600">
        <div className="flex items-center gap-1">
          <span>⏱️</span>
          <span>{recipe.cooking_time}分</span>
        </div>
        <div className="flex items-center gap-1">
          <span>🥘</span>
          <span>{recipe.ingredients.length}種類</span>
        </div>
      </div>

      {/* 材料一覧 */}
      <div className="mb-3">
        <p className="text-sm text-gray-600 mb-1">材料:</p>
        <div className="flex flex-wrap gap-1">
          {recipe.ingredients.slice(0, 4).map((ingredient, index) => {
            const ingredientName = typeof ingredient === 'string' ? ingredient : ingredient.name;
            return (
              <span 
                key={`${recipe.id}-ingredient-${ingredientName}-${index}`}
                className="inline-block bg-gray-100 text-gray-700 px-2 py-1 rounded text-xs"
              >
                {ingredientName}
              </span>
            );
          })}
          {recipe.ingredients.length > 4 && (
            <span className="inline-block bg-gray-100 text-gray-700 px-2 py-1 rounded text-xs">
              +{recipe.ingredients.length - 4}種類
            </span>
          )}
        </div>
      </div>

      {/* 手順数 */}
      <div className="mb-3">
        <p className="text-sm text-gray-600">
          <span>📝</span>
          <span className="ml-1">{recipe.steps.length}ステップで完成</span>
        </p>
      </div>

      {/* タグ */}
      {recipe.tags && recipe.tags.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {(() => {
            // 基本タグ（全レシピ共通で不要）を除外
            const excludeBasicTags = ['簡単', '時短', 'ずぼら'];
            const mealTypeCategories = ['主食', '副菜', '汁物', 'デザート', '丼・ワンプレート', '常備菜・作り置き', 'おやつ・甘味'];
            
            // 有用なタグのみをフィルタリング
            const usefulTags = recipe.tags.filter(tag => !excludeBasicTags.includes(tag));
            const mealTypeTag = usefulTags.find(tag => mealTypeCategories.includes(tag));
            const otherUsefulTags = usefulTags.filter(tag => !mealTypeCategories.includes(tag));
            
            // meal_typeタグを最初に、その後その他の有用なタグを表示
            const displayTags = [];
            if (mealTypeTag) {
              displayTags.push(mealTypeTag);
            }
            displayTags.push(...otherUsefulTags.slice(0, 3)); // meal_typeタグ + 最大3つの追加タグ
            
            return (
              <>
                {displayTags.map((tag, index) => (
                  <span 
                    key={`${recipe.id}-tag-${tag}-${index}`}
                    className={`inline-block px-2 py-1 rounded-full text-xs font-medium ${
                      mealTypeCategories.includes(tag) 
                        ? 'bg-purple-100 text-purple-800 border border-purple-200' 
                        : 'bg-blue-100 text-blue-700'
                    }`}
                  >
                    #{tag}
                  </span>
                ))}
                {usefulTags.length > displayTags.length && (
                  <span className="inline-block bg-gray-100 text-gray-700 px-2 py-1 rounded-full text-xs">
                    +{usefulTags.length - displayTags.length}
                  </span>
                )}
              </>
            );
          })()}
        </div>
      )}
    </div>
  );
};

export default RecipeCard;