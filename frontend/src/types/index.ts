// API Response Types
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

// Recipe Types
export interface Recipe {
  id: number;
  title: string;
  ingredients: Ingredient[];
  steps: string[];
  cooking_time: number;
  laziness_score: number;
  tags: string[];
  cuisine?: string;
  servings?: number;
  difficulty?: string;
}

export interface Ingredient {
  name: string;
  amount: string;
  unit?: string;
  category?: string;
}

// Meal Plan Types
export interface MealPlan {
  id: number;
  week_start_date: string;
  daily_recipes: Record<string, DailyRecipe>;
  shopping_list: ShoppingItem[];
  total_cost?: number;
}

export interface DailyRecipe {
  breakfast?: Recipe;
  lunch?: Recipe;
  dinner?: Recipe;
}

export interface ShoppingItem {
  item: string;
  amount: string;
  category?: string;
  cost?: number;
  checked?: boolean;
}

// Request Types
export interface GenerateRecipeRequest {
  ingredients: string[];
  cooking_time?: number;
  meal_type?: 'breakfast' | 'lunch' | 'dinner';
  dietary_restrictions?: string[];
  laziness_level?: number;
}

export interface CreateMealPlanRequest {
  days: number;
  servings: number;
  preferences: MealPlanPreferences;
}

export interface MealPlanPreferences {
  dietary_restrictions?: string[];
  preferred_tags?: string[];
  budget_limit?: number;
  laziness_preference?: number;
}

// Search Types
export interface RecipeSearchParams {
  tag?: string;
  ingredient?: string;
  max_cooking_time?: number;
  min_laziness_score?: number;
  limit?: number;
  offset?: number;
}

// UI State Types
export interface LoadingState {
  isLoading: boolean;
  error?: string | null;
}

export interface FormErrors {
  [key: string]: string | undefined;
}