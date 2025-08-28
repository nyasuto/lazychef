import React, { useState } from 'react';
import RecipeSearch from '../components/recipe/RecipeSearch';
import RecipeCard from '../components/recipe/RecipeCard';
import LoadingSpinner from '../components/common/LoadingSpinner';
import { useRecipes } from '../hooks/useRecipes';
import type { Recipe } from '../types';
import type { SearchFilters } from '../components/recipe/RecipeSearch';

const RecipeListPage: React.FC = () => {
  const { recipes, total, loading, error, searchRecipes, clearError } = useRecipes();
  const [currentPage, setCurrentPage] = useState(1);
  const [currentFilters, setCurrentFilters] = useState<SearchFilters>({
    query: '',
    tags: [],
    ingredients: [],
    maxCookingTime: null,
    minLazinessScore: null,
  });

  const recipesPerPage = 12;
  const totalPages = Math.ceil(total / recipesPerPage);

  const handleSearch = async (filters: SearchFilters) => {
    setCurrentFilters(filters);
    setCurrentPage(1);
    await searchRecipes(filters, 1);
  };

  const handlePageChange = async (page: number) => {
    setCurrentPage(page);
    await searchRecipes(currentFilters, page);
  };

  const handleRecipeClick = (recipe: Recipe) => {
    console.log('Recipe clicked:', recipe);
    // Phase 3でレシピ詳細ページに遷移する機能を実装予定
  };

  const hasActiveFilters = 
    currentFilters.query ||
    currentFilters.tags.length > 0 ||
    currentFilters.ingredients.length > 0 ||
    currentFilters.maxCookingTime ||
    currentFilters.minLazinessScore;

  return (
    <div className="space-y-6">
      {/* ヘッダー */}
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold text-gray-900">レシピ一覧</h1>
        <div className="text-sm text-gray-600">
          {total > 0 && (
            <span>
              {hasActiveFilters ? '検索結果: ' : '全レシピ: '}
              <span className="font-medium text-gray-900">{total.toLocaleString()}件</span>
            </span>
          )}
        </div>
      </div>

      {/* 検索フォーム */}
      <RecipeSearch onSearch={handleSearch} loading={loading} />

      {/* エラー表示 */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <span className="text-red-600 mr-2">⚠️</span>
              <span className="text-red-800">{error}</span>
            </div>
            <button
              onClick={clearError}
              className="text-red-600 hover:text-red-800 text-sm font-medium"
            >
              閉じる
            </button>
          </div>
        </div>
      )}

      {/* ローディング状態 */}
      {loading && (
        <div className="flex justify-center py-12">
          <LoadingSpinner size="lg" />
        </div>
      )}

      {/* レシピ一覧 */}
      {!loading && (
        <>
          {recipes.length > 0 ? (
            <>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                {recipes.map((recipe) => (
                  <RecipeCard 
                    key={recipe.id} 
                    recipe={recipe}
                    onClick={handleRecipeClick}
                  />
                ))}
              </div>

              {/* ページネーション */}
              {totalPages > 1 && (
                <div className="flex justify-center items-center space-x-2 py-8">
                  <button
                    onClick={() => handlePageChange(currentPage - 1)}
                    disabled={currentPage === 1}
                    className="px-3 py-2 rounded-lg text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed bg-gray-100 text-gray-700 hover:bg-gray-200"
                  >
                    ← 前へ
                  </button>

                  <div className="flex space-x-1">
                    {Array.from({ length: Math.min(totalPages, 7) }, (_, i) => {
                      let pageNum;
                      if (totalPages <= 7) {
                        pageNum = i + 1;
                      } else if (currentPage <= 4) {
                        pageNum = i + 1;
                      } else if (currentPage >= totalPages - 3) {
                        pageNum = totalPages - 6 + i;
                      } else {
                        pageNum = currentPage - 3 + i;
                      }

                      return (
                        <button
                          key={pageNum}
                          onClick={() => handlePageChange(pageNum)}
                          className={`px-3 py-2 rounded-lg text-sm font-medium ${
                            currentPage === pageNum
                              ? 'bg-blue-600 text-white'
                              : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                          }`}
                        >
                          {pageNum}
                        </button>
                      );
                    })}
                  </div>

                  <button
                    onClick={() => handlePageChange(currentPage + 1)}
                    disabled={currentPage === totalPages}
                    className="px-3 py-2 rounded-lg text-sm font-medium disabled:opacity-50 disabled:cursor-not-allowed bg-gray-100 text-gray-700 hover:bg-gray-200"
                  >
                    次へ →
                  </button>
                </div>
              )}
            </>
          ) : !loading && (
            <div className="text-center py-12">
              <div className="text-6xl mb-4">🔍</div>
              <h3 className="text-xl font-medium text-gray-900 mb-2">
                {hasActiveFilters ? 'レシピが見つかりません' : 'レシピがありません'}
              </h3>
              <p className="text-gray-600 mb-6">
                {hasActiveFilters 
                  ? '検索条件を変更してみてください'
                  : 'まだレシピが登録されていません'
                }
              </p>
              {hasActiveFilters && (
                <button
                  onClick={() => handleSearch({
                    query: '',
                    tags: [],
                    ingredients: [],
                    maxCookingTime: null,
                    minLazinessScore: null,
                  })}
                  className="btn-primary"
                >
                  すべてのレシピを見る
                </button>
              )}
            </div>
          )}
        </>
      )}
    </div>
  );
};

export default RecipeListPage;