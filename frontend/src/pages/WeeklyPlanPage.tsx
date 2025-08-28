import React, { useState } from 'react';
import WeeklyPlan from '../components/meal-plan/WeeklyPlan';
import ShoppingList from '../components/meal-plan/ShoppingList';
import Button from '../components/common/Button';
import LoadingSpinner from '../components/common/LoadingSpinner';
import { createMealPlan } from '../services/api';
import type { MealPlan, CreateMealPlanRequest, Recipe, ShoppingItem } from '../types';

const WeeklyPlanPage: React.FC = () => {
  const [mealPlan, setMealPlan] = useState<MealPlan | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentWeekStart, setCurrentWeekStart] = useState<Date>(getStartOfWeek(new Date()));

  // 週の開始日（月曜日）を取得
  function getStartOfWeek(date: Date): Date {
    const d = new Date(date);
    const day = d.getDay();
    const diff = d.getDate() - day + (day === 0 ? -6 : 1); // 月曜日を週の開始にする
    return new Date(d.setDate(diff));
  }

  // 週間献立を生成
  const handleGenerateMealPlan = async () => {
    setIsLoading(true);
    setError(null);
    
    try {
      const request: CreateMealPlanRequest = {
        days: 7,
        servings: 2,
        preferences: {
          dietary_restrictions: [],
          preferred_tags: ['簡単', '時短'],
          budget_limit: 5000,
          laziness_preference: 8
        }
      };

      const newMealPlan = await createMealPlan(request);
      
      // 現在の週開始日を設定
      const updatedMealPlan = {
        ...newMealPlan,
        week_start_date: currentWeekStart.toISOString().split('T')[0]
      };
      
      setMealPlan(updatedMealPlan);
    } catch (err) {
      setError(err instanceof Error ? err.message : '献立の生成に失敗しました');
    } finally {
      setIsLoading(false);
    }
  };

  // サンプル献立データ（デモ用）
  const createSampleMealPlan = (): MealPlan => {
    const startDate = currentWeekStart.toISOString().split('T')[0];
    
    return {
      id: 1,
      week_start_date: startDate,
      daily_recipes: {
        [startDate]: {
          breakfast: {
            id: 1,
            title: '簡単目玉焼きトースト',
            ingredients: [
              { name: '食パン', amount: '2枚' },
              { name: '卵', amount: '2個' },
              { name: 'バター', amount: '大さじ1' }
            ],
            steps: ['パンをトースト', '卵を焼く', 'パンに卵をのせる'],
            cooking_time: 5,
            laziness_score: 9,
            tags: ['簡単', '朝食', '時短']
          },
          lunch: {
            id: 2,
            title: 'ツナマヨおにぎり',
            ingredients: [
              { name: 'ご飯', amount: '茶碗2杯' },
              { name: 'ツナ缶', amount: '1缶' },
              { name: 'マヨネーズ', amount: '大さじ2' },
              { name: '海苔', amount: '4枚' }
            ],
            steps: ['ツナとマヨを混ぜる', 'ご飯に具を混ぜる', '海苔で包む'],
            cooking_time: 8,
            laziness_score: 8,
            tags: ['簡単', '昼食', 'おにぎり']
          }
        }
      },
      shopping_list: [
        { item: '食パン', amount: '1斤', category: 'パン', cost: 150 },
        { item: '卵', amount: '1パック', category: '乳製品', cost: 250 },
        { item: 'バター', amount: '1個', category: '乳製品', cost: 300 },
        { item: 'ご飯', amount: '2kg', category: '穀物', cost: 800 },
        { item: 'ツナ缶', amount: '3缶', category: '魚介類', cost: 450 },
        { item: 'マヨネーズ', amount: '1本', category: '調味料', cost: 200 },
        { item: '海苔', amount: '1袋', category: 'その他', cost: 300 }
      ],
      total_cost: 2450
    };
  };

  // 週移動
  const handlePreviousWeek = () => {
    const newDate = new Date(currentWeekStart);
    newDate.setDate(newDate.getDate() - 7);
    setCurrentWeekStart(newDate);
  };

  const handleNextWeek = () => {
    const newDate = new Date(currentWeekStart);
    newDate.setDate(newDate.getDate() + 7);
    setCurrentWeekStart(newDate);
  };

  // レシピクリック
  const handleRecipeClick = (recipe: Recipe) => {
    console.log('Recipe clicked:', recipe.title);
    // TODO: レシピ詳細モーダル表示
  };

  // 買い物リスト項目トグル
  const handleItemToggle = (index: number, checked: boolean) => {
    if (!mealPlan) return;

    const updatedItems = [...mealPlan.shopping_list];
    updatedItems[index] = { ...updatedItems[index], checked };
    
    setMealPlan({
      ...mealPlan,
      shopping_list: updatedItems
    });
  };

  // 買い物リスト項目追加
  const handleItemAdd = (item: ShoppingItem) => {
    if (!mealPlan) return;

    setMealPlan({
      ...mealPlan,
      shopping_list: [...mealPlan.shopping_list, item]
    });
  };

  // 買い物リスト項目削除
  const handleItemRemove = (index: number) => {
    if (!mealPlan) return;

    const updatedItems = mealPlan.shopping_list.filter((_, i) => i !== index);
    setMealPlan({
      ...mealPlan,
      shopping_list: updatedItems
    });
  };

  return (
    <div className="space-y-6">
      {/* ヘッダー */}
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold text-gray-900">週間献立</h1>
        <div className="flex gap-3">
          <Button
            onClick={() => setMealPlan(createSampleMealPlan())}
            variant="outline"
          >
            デモ表示
          </Button>
          <Button
            onClick={handleGenerateMealPlan}
            disabled={isLoading}
          >
            {isLoading ? (
              <>
                <LoadingSpinner size="sm" />
                <span className="ml-2">生成中...</span>
              </>
            ) : (
              '🤖 AI献立生成'
            )}
          </Button>
        </div>
      </div>

      {/* 週ナビゲーション */}
      <div className="card">
        <div className="flex items-center justify-between">
          <Button onClick={handlePreviousWeek} variant="outline">
            ← 前の週
          </Button>
          <div className="text-center">
            <h2 className="text-lg font-semibold">
              {currentWeekStart.toLocaleDateString('ja-JP', {
                year: 'numeric',
                month: 'long',
                day: 'numeric'
              })}の週
            </h2>
          </div>
          <Button onClick={handleNextWeek} variant="outline">
            次の週 →
          </Button>
        </div>
      </div>

      {/* エラー表示 */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-center">
            <div className="text-red-600 mr-3">⚠️</div>
            <div>
              <h3 className="font-medium text-red-800">エラーが発生しました</h3>
              <p className="text-red-700 text-sm mt-1">{error}</p>
            </div>
          </div>
        </div>
      )}

      {/* 献立表示 */}
      {mealPlan ? (
        <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">
          {/* 週間献立 */}
          <div className="xl:col-span-2">
            <WeeklyPlan 
              mealPlan={mealPlan}
              onRecipeClick={handleRecipeClick}
              onEditClick={() => console.log('Edit clicked')}
            />
          </div>

          {/* 買い物リスト */}
          <div>
            <ShoppingList 
              items={mealPlan.shopping_list}
              onItemToggle={handleItemToggle}
              onItemAdd={handleItemAdd}
              onItemRemove={handleItemRemove}
              showAddForm={true}
            />
          </div>
        </div>
      ) : (
        /* 空状態 */
        <div className="card">
          <div className="text-center py-12">
            <div className="text-6xl mb-4">📅</div>
            <h3 className="text-xl font-medium text-gray-900 mb-2">週間献立を作成しましょう</h3>
            <p className="text-gray-600 mb-6">
              AIが材料を使い回した最適な1週間の献立を作成し、<br />
              買い物リストも自動生成します
            </p>
            <div className="text-sm text-gray-500">
              機能：
              <ul className="mt-2 space-y-1 max-w-md mx-auto">
                <li>• 7日×3食の自動献立生成</li>
                <li>• 材料使い回し最適化</li>
                <li>• 買い物リスト自動作成</li>
                <li>• 食費概算表示</li>
                <li>• ラジネス重視の簡単レシピ</li>
              </ul>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default WeeklyPlanPage;