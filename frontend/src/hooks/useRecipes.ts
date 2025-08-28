import { useState, useEffect, useCallback } from 'react';
import type { Recipe, SearchRecipesResponse } from '../types';
import { recipeService } from '../services/recipeService';
import type { SearchFilters } from '../components/recipe/RecipeSearch';

interface UseRecipesResult {
  recipes: Recipe[];
  total: number;
  loading: boolean;
  error: string | null;
  searchRecipes: (filters: SearchFilters, page?: number) => Promise<void>;
  clearError: () => void;
}

export const useRecipes = (): UseRecipesResult => {
  const [recipes, setRecipes] = useState<Recipe[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const searchRecipes = useCallback(async (filters: SearchFilters, page = 1) => {
    try {
      setLoading(true);
      setError(null);

      // SearchFiltersから検索パラメータを構築
      const searchParams: Record<string, string> = {
        page: page.toString(),
        limit: '12', // 1ページあたり12件
      };

      if (filters.query.trim()) {
        searchParams.query = filters.query.trim();
      }

      if (filters.tags.length > 0) {
        searchParams.tags = filters.tags.join(',');
      }

      if (filters.ingredients.length > 0) {
        searchParams.ingredients = filters.ingredients.join(',');
      }

      if (filters.maxCookingTime) {
        searchParams.max_cooking_time = filters.maxCookingTime.toString();
      }

      if (filters.minLazinessScore) {
        searchParams.min_laziness_score = filters.minLazinessScore.toString();
      }

      const response = await recipeService.searchRecipes(searchParams);
      
      setRecipes(response.data.recipes);
      setTotal(response.data.total);
    } catch (err) {
      console.error('Recipe search error:', err);
      setError(err instanceof Error ? err.message : 'レシピの検索に失敗しました');
      setRecipes([]);
      setTotal(0);
    } finally {
      setLoading(false);
    }
  }, []);

  const clearError = useCallback(() => {
    setError(null);
  }, []);

  return {
    recipes,
    total,
    loading,
    error,
    searchRecipes,
    clearError,
  };
};