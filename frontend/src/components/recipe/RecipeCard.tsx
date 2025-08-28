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
          {recipe.ingredients.slice(0, 4).map((ingredient, index) => (
            <span 
              key={index}
              className="inline-block bg-gray-100 text-gray-700 px-2 py-1 rounded text-xs"
            >
              {typeof ingredient === 'string' ? ingredient : ingredient.name}
            </span>
          ))}
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
          {recipe.tags.slice(0, 3).map((tag, index) => (
            <span 
              key={index}
              className="inline-block bg-blue-100 text-blue-700 px-2 py-1 rounded-full text-xs"
            >
              #{tag}
            </span>
          ))}
          {recipe.tags.length > 3 && (
            <span className="inline-block bg-gray-100 text-gray-700 px-2 py-1 rounded-full text-xs">
              +{recipe.tags.length - 3}
            </span>
          )}
        </div>
      )}
    </div>
  );
};

export default RecipeCard;