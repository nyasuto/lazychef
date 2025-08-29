import React from 'react';
import Modal from '../common/Modal';
import type { Recipe } from '../../types';

interface RecipeDetailModalProps {
  isOpen: boolean;
  onClose: () => void;
  recipe: Recipe | null;
}

const RecipeDetailModal: React.FC<RecipeDetailModalProps> = ({
  isOpen,
  onClose,
  recipe
}) => {
  if (!recipe) return null;

  // ラジネススコアの表示用
  const getLazinessDisplay = (score: number) => {
    const stars = '⭐'.repeat(Math.floor(score));
    return `${score.toFixed(1)} ${stars}`;
  };

  // 難易度の表示
  const getDifficultyColor = (score: number) => {
    if (score >= 9) return 'text-green-600 bg-green-50';
    if (score >= 7) return 'text-yellow-600 bg-yellow-50';
    return 'text-red-600 bg-red-50';
  };

  const getDifficultyText = (score: number) => {
    if (score >= 9) return '超簡単';
    if (score >= 7) return '簡単';
    if (score >= 5) return '普通';
    return '少し手間';
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={recipe.title}
      size="lg"
      className="max-h-[90vh] overflow-y-auto"
    >
      <div className="space-y-6">
        {/* レシピ情報サマリー */}
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4 p-4 bg-gray-50 rounded-lg">
          <div className="text-center">
            <div className="text-sm text-gray-600">調理時間</div>
            <div className="text-lg font-semibold text-blue-600">
              {recipe.cooking_time}分
            </div>
          </div>
          
          <div className="text-center">
            <div className="text-sm text-gray-600">ラジネス度</div>
            <div className="text-sm font-medium">
              {getLazinessDisplay(recipe.laziness_score)}
            </div>
          </div>
          
          <div className="text-center col-span-2 md:col-span-1">
            <div className="text-sm text-gray-600">難易度</div>
            <div className={`inline-block px-2 py-1 rounded-full text-xs font-medium ${getDifficultyColor(recipe.laziness_score)}`}>
              {getDifficultyText(recipe.laziness_score)}
            </div>
          </div>
        </div>

        {/* タグ */}
        {recipe.tags && recipe.tags.length > 0 && (
          <div>
            <h3 className="font-medium text-gray-900 mb-2">タグ</h3>
            <div className="flex flex-wrap gap-2">
              {recipe.tags.map((tag, index) => (
                <span
                  key={index}
                  className="inline-block px-3 py-1 text-xs font-medium bg-blue-100 text-blue-800 rounded-full"
                >
                  {tag}
                </span>
              ))}
            </div>
          </div>
        )}

        {/* 材料リスト */}
        <div>
          <h3 className="font-medium text-gray-900 mb-3 flex items-center">
            <span className="mr-2">🥘</span>
            材料
            {recipe.servings && (
              <span className="ml-2 text-sm text-gray-600">
                ({recipe.servings}人分)
              </span>
            )}
          </h3>
          <div className="bg-white border border-gray-200 rounded-lg overflow-hidden">
            <div className="divide-y divide-gray-200">
              {recipe.ingredients.map((ingredient, index) => (
                <div key={index} className="px-4 py-3 flex justify-between items-center hover:bg-gray-50">
                  <span className="font-medium text-gray-900">{ingredient.name}</span>
                  <span className="text-gray-600">{ingredient.amount}</span>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* 調理手順 */}
        <div>
          <h3 className="font-medium text-gray-900 mb-3 flex items-center">
            <span className="mr-2">👨‍🍳</span>
            調理手順
          </h3>
          <div className="space-y-3">
            {recipe.steps.map((step, index) => (
              <div key={index} className="flex items-start space-x-3">
                <div className="flex-shrink-0 w-6 h-6 bg-blue-600 text-white text-sm font-medium rounded-full flex items-center justify-center">
                  {index + 1}
                </div>
                <div className="flex-1 text-gray-700 leading-relaxed">
                  {step}
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* その他の情報 */}
        {(recipe.cuisine || recipe.difficulty) && (
          <div className="pt-4 border-t border-gray-200">
            <div className="grid grid-cols-2 gap-4 text-sm">
              {recipe.cuisine && (
                <div>
                  <span className="text-gray-600">料理ジャンル:</span>
                  <span className="ml-2 font-medium">{recipe.cuisine}</span>
                </div>
              )}
              {recipe.difficulty && (
                <div>
                  <span className="text-gray-600">難易度:</span>
                  <span className="ml-2 font-medium">{recipe.difficulty}</span>
                </div>
              )}
            </div>
          </div>
        )}

        {/* アクションボタン */}
        <div className="flex justify-end space-x-3 pt-4 border-t border-gray-200">
          <button
            onClick={onClose}
            className="px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-md transition-colors"
          >
            閉じる
          </button>
          <button
            onClick={() => {
              // TODO: レシピをお気に入りに追加する機能
              console.log('お気に入りに追加:', recipe.title);
            }}
            className="px-4 py-2 bg-blue-600 text-white hover:bg-blue-700 rounded-md transition-colors"
          >
            お気に入りに追加
          </button>
        </div>
      </div>
    </Modal>
  );
};

export default RecipeDetailModal;