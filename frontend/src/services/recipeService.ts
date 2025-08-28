import api from './api';
import type { 
  Recipe, 
  GenerateRecipeRequest, 
  RecipeSearchParams, 
  ApiResponse,
  SearchRecipesResponse
} from '../types';

export const recipeService = {
  // Search recipes
  async searchRecipes(params: RecipeSearchParams = {}): Promise<SearchRecipesResponse> {
    const response = await api.get<ApiResponse<{ recipes: Recipe[]; total: number }>>('/recipes/search', {
      params,
    });
    
    return {
      success: true,
      data: {
        recipes: response.data.data?.recipes || [],
        total: response.data.data?.total || 0
      }
    };
  },

  // Generate single recipe
  async generateRecipe(request: GenerateRecipeRequest): Promise<Recipe> {
    const response = await api.post<ApiResponse<Recipe>>('/recipes/generate', request);
    
    if (!response.data.data) {
      throw new Error('Failed to generate recipe');
    }
    
    return response.data.data;
  },

  // Generate batch recipes
  async generateBatchRecipes(request: GenerateRecipeRequest & { count: number }): Promise<Recipe[]> {
    const response = await api.post<ApiResponse<Recipe[]>>('/recipes/generate-batch', request);
    
    if (!response.data.data) {
      throw new Error('Failed to generate recipes');
    }
    
    return response.data.data;
  },

  // Get recipe generator health
  async getGeneratorHealth(): Promise<{ status: string; openai_configured: boolean }> {
    const response = await api.get('/recipes/health');
    return response.data;
  },

  // Test recipe generation (for debugging)
  async testRecipeGeneration(): Promise<unknown> {
    const response = await api.get('/recipes/test');
    return response.data;
  },

  // Clear recipe cache
  async clearCache(): Promise<void> {
    await api.post('/recipes/clear-cache');
  },
};