import axios from 'axios';

// Create axios instance with base configuration
const api = axios.create({
  baseURL: 'http://localhost:8080/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    // Add any auth headers here if needed in the future
    console.log(`Making ${config.method?.toUpperCase()} request to: ${config.url}`);
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    console.error('API Error:', error.response?.data || error.message);
    
    // Handle different error types
    if (error.response) {
      // Server responded with error status
      const { status, data } = error.response;
      
      switch (status) {
        case 400:
          throw new Error(data.message || 'Bad request');
        case 401:
          throw new Error('Unauthorized access');
        case 403:
          throw new Error('Access forbidden');
        case 404:
          throw new Error('Resource not found');
        case 429:
          throw new Error('Too many requests. Please try again later.');
        case 500:
          throw new Error('Server error. Please try again later.');
        case 503:
          throw new Error(data.message || 'Service unavailable');
        default:
          throw new Error(data.message || 'An unexpected error occurred');
      }
    } else if (error.request) {
      // Request made but no response received
      throw new Error('Network error. Please check your connection.');
    } else {
      // Error in request configuration
      throw new Error(error.message || 'Request failed');
    }
  }
);

// Import types
import type { 
  Recipe, 
  SearchRecipesResponse, 
  RecipeSearchParams,
  MealPlan,
  CreateMealPlanRequest,
  GenerateRecipeRequest 
} from '../types';

// API Functions

// Recipe APIs
export const searchRecipes = async (params: RecipeSearchParams): Promise<SearchRecipesResponse> => {
  const response = await api.get('/recipes/search', { params });
  return response.data;
};

export const generateRecipe = async (request: GenerateRecipeRequest): Promise<Recipe> => {
  const response = await api.post('/recipes/generate', request);
  return response.data.recipe;
};

export const generateRecipes = async (request: GenerateRecipeRequest & { count: number }): Promise<Recipe[]> => {
  const response = await api.post('/recipes/generate-batch', request);
  return response.data.recipes;
};

// Meal Plan APIs
export const createMealPlan = async (request: CreateMealPlanRequest): Promise<MealPlan> => {
  const response = await api.post('/meal-plans/create', request);
  return response.data.meal_plan;
};

// Health check
export const healthCheck = async (): Promise<{ status: string }> => {
  const response = await api.get('/health');
  return response.data;
};

export default api;