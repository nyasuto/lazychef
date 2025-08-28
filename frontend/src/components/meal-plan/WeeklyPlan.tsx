import React from 'react';
import type { MealPlan, DailyRecipe, Recipe } from '../../types';
import RecipeCard from '../recipe/RecipeCard';

interface WeeklyPlanProps {
  mealPlan: MealPlan;
  onRecipeClick?: (recipe: Recipe) => void;
  onEditClick?: () => void;
}

const WeeklyPlan: React.FC<WeeklyPlanProps> = ({ 
  mealPlan, 
  onRecipeClick, 
  onEditClick 
}) => {
  const weekDays = ['月', '火', '水', '木', '金', '土', '日'];
  const mealTypes = [
    { key: 'breakfast', label: '朝食', icon: '🌅' },
    { key: 'lunch', label: '昼食', icon: '☀️' },
    { key: 'dinner', label: '夕食', icon: '🌙' }
  ] as const;

  const getDateForDay = (dayIndex: number) => {
    const startDate = new Date(mealPlan.week_start_date);
    const date = new Date(startDate);
    date.setDate(startDate.getDate() + dayIndex);
    return date;
  };

  const formatDate = (date: Date) => {
    return `${date.getMonth() + 1}/${date.getDate()}`;
  };

  const getDayKey = (dayIndex: number) => {
    const date = getDateForDay(dayIndex);
    return date.toISOString().split('T')[0];
  };

  const renderMealSlot = (dayIndex: number, mealType: { key: keyof DailyRecipe; label: string; icon: string }) => {
    const dayKey = getDayKey(dayIndex);
    const dailyRecipe: DailyRecipe = mealPlan.daily_recipes[dayKey] || {};
    const recipe = dailyRecipe[mealType.key as keyof DailyRecipe];

    if (recipe) {
      return (
        <div className="h-full">
          <RecipeCard recipe={recipe} onClick={onRecipeClick} />
        </div>
      );
    }

    return (
      <div className="h-full border-2 border-dashed border-gray-300 rounded-lg flex items-center justify-center bg-gray-50 hover:bg-gray-100 transition-colors duration-200 min-h-[120px]">
        <div className="text-center text-gray-500">
          <div className="text-2xl mb-2">{mealType.icon}</div>
          <div className="text-sm">未設定</div>
        </div>
      </div>
    );
  };

  return (
    <div className="bg-white rounded-lg shadow-md overflow-hidden">
      {/* ヘッダー */}
      <div className="bg-primary-600 text-white p-4">
        <div className="flex justify-between items-center">
          <div>
            <h2 className="text-xl font-semibold">週間献立</h2>
            <p className="text-primary-100 text-sm">
              {new Date(mealPlan.week_start_date).toLocaleDateString('ja-JP', {
                year: 'numeric',
                month: 'long',
                day: 'numeric'
              })}
              の週
            </p>
          </div>
          {onEditClick && (
            <button
              onClick={onEditClick}
              className="bg-white text-primary-600 px-4 py-2 rounded-lg font-medium hover:bg-gray-50 transition-colors duration-200"
            >
              編集
            </button>
          )}
        </div>
      </div>

      {/* 献立グリッド */}
      <div className="p-4">
        <div className="grid grid-cols-1 lg:grid-cols-7 gap-4">
          {/* 食事タイプ列（デスクトップ用） */}
          <div className="hidden lg:block">
            <div className="h-12"></div> {/* ヘッダー空間 */}
            {mealTypes.map((mealType) => (
              <div key={mealType.key} className="h-40 flex items-center justify-center bg-gray-50 rounded-lg mb-4">
                <div className="text-center">
                  <div className="text-xl mb-1">{mealType.icon}</div>
                  <div className="text-sm font-medium text-gray-700">{mealType.label}</div>
                </div>
              </div>
            ))}
          </div>

          {/* 各日の献立 */}
          {weekDays.map((day, dayIndex) => (
            <div key={dayIndex} className="space-y-4">
              {/* 日付ヘッダー */}
              <div className="text-center">
                <div className="font-semibold text-gray-900">{day}曜日</div>
                <div className="text-sm text-gray-600">{formatDate(getDateForDay(dayIndex))}</div>
              </div>

              {/* 食事スロット */}
              <div className="space-y-4">
                {mealTypes.map((mealType) => (
                  <div key={mealType.key}>
                    {/* モバイル用食事ラベル */}
                    <div className="lg:hidden flex items-center gap-2 mb-2">
                      <span>{mealType.icon}</span>
                      <span className="text-sm font-medium text-gray-700">{mealType.label}</span>
                    </div>
                    {renderMealSlot(dayIndex, mealType)}
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* 統計情報 */}
      <div className="bg-gray-50 p-4 border-t">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-center">
          <div>
            <div className="text-2xl font-semibold text-primary-600">
              {Object.values(mealPlan.daily_recipes).reduce((count, daily) => {
                return count + Object.values(daily).filter(Boolean).length;
              }, 0)}
            </div>
            <div className="text-sm text-gray-600">設定済み</div>
          </div>
          <div>
            <div className="text-2xl font-semibold text-primary-600">
              {mealPlan.shopping_list?.length || 0}
            </div>
            <div className="text-sm text-gray-600">買い物項目</div>
          </div>
          <div>
            <div className="text-2xl font-semibold text-primary-600">
              {mealPlan.total_cost ? `¥${mealPlan.total_cost.toLocaleString()}` : '---'}
            </div>
            <div className="text-sm text-gray-600">予算目安</div>
          </div>
          <div>
            <div className="text-2xl font-semibold text-primary-600">
              {Object.values(mealPlan.daily_recipes).reduce((_avg, daily) => {
                const recipes = Object.values(daily).filter(Boolean);
                const totalLaziness = recipes.reduce((sum, recipe) => sum + (recipe?.laziness_score || 0), 0);
                return recipes.length > 0 ? totalLaziness / recipes.length : 0;
              }, 0).toFixed(1)}
            </div>
            <div className="text-sm text-gray-600">平均ラジネス</div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default WeeklyPlan;