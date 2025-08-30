import React, { useState, useEffect, useRef } from 'react';
import Input from '../common/Input';

export interface SearchFilters {
  query: string;
  tags: string[];
  ingredients: string[];
  maxCookingTime: number | null;
  minLazinessScore: number | null;
}

interface RecipeSearchProps {
  onSearch: (filters: SearchFilters) => void;
  loading?: boolean;
}

const COMMON_TAGS = [
  '簡単', '時短', '節約', '一人前', 'ヘルシー', 
  '和食', '洋食', '中華', '麺類', 'ご飯もの'
];

const COMMON_INGREDIENTS = [
  '豚肉', '鶏肉', '牛肉', '卵', '豆腐',
  'じゃがいも', '玉ねぎ', 'にんじん', 'キャベツ', 'もやし'
];

const RecipeSearch: React.FC<RecipeSearchProps> = ({ onSearch, loading = false }) => {
  const [filters, setFilters] = useState<SearchFilters>({
    query: '',
    tags: [],
    ingredients: [],
    maxCookingTime: null,
    minLazinessScore: null,
  });

  const [showAdvanced, setShowAdvanced] = useState(false);
  const initialLoadRef = useRef(false);
  const searchTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  // リアルタイム検索 - フィルター変更時に自動実行（300msのデバウンス）
  useEffect(() => {
    if (!initialLoadRef.current) {
      // 初回のみすぐに検索
      initialLoadRef.current = true;
      onSearch(filters);
      return;
    }

    // 2回目以降はデバウンスして検索
    if (searchTimeoutRef.current) {
      clearTimeout(searchTimeoutRef.current);
    }
    searchTimeoutRef.current = setTimeout(() => {
      onSearch(filters);
    }, 300);

    // クリーンアップ
    return () => {
      if (searchTimeoutRef.current) {
        clearTimeout(searchTimeoutRef.current);
      }
    };
  }, [filters, onSearch]);

  // リアルタイム検索のためsubmitハンドラーは不要（useEffectで自動実行）

  const handleTagToggle = (tag: string) => {
    setFilters(prev => ({
      ...prev,
      tags: prev.tags.includes(tag)
        ? prev.tags.filter(t => t !== tag)
        : [...prev.tags, tag]
    }));
  };

  const handleIngredientToggle = (ingredient: string) => {
    setFilters(prev => ({
      ...prev,
      ingredients: prev.ingredients.includes(ingredient)
        ? prev.ingredients.filter(i => i !== ingredient)
        : [...prev.ingredients, ingredient]
    }));
  };

  const handleReset = () => {
    const resetFilters: SearchFilters = {
      query: '',
      tags: [],
      ingredients: [],
      maxCookingTime: null,
      minLazinessScore: null,
    };
    setFilters(resetFilters);
    onSearch(resetFilters);
  };

  return (
    <div className="card mb-6">
      <div className="space-y-4">
        {/* 基本検索 */}
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            レシピを検索 {loading && <span className="text-gray-500 text-xs">（検索中...）</span>}
          </label>
          <Input
            type="text"
            placeholder="料理名やキーワードを入力..."
            value={filters.query}
            onChange={(e) => setFilters(prev => ({ ...prev, query: e.target.value }))}
            className="w-full"
          />
        </div>

        {/* 詳細検索トグル */}
        <div className="flex justify-between items-center">
          <button
            type="button"
            onClick={() => setShowAdvanced(!showAdvanced)}
            className="text-blue-600 hover:text-blue-800 text-sm font-medium"
          >
            {showAdvanced ? '▲ 詳細検索を閉じる' : '▼ 詳細検索を開く'}
          </button>
          <button
            type="button"
            onClick={handleReset}
            className="text-gray-600 hover:text-gray-800 text-sm"
          >
            🗑️ リセット
          </button>
        </div>

        {/* 詳細検索 */}
        {showAdvanced && (
          <div className="space-y-4 pt-4 border-t border-gray-200">
            {/* タグ選択 */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                タグで絞り込み
              </label>
              <div className="flex flex-wrap gap-2">
                {COMMON_TAGS.map(tag => (
                  <button
                    key={tag}
                    type="button"
                    onClick={() => handleTagToggle(tag)}
                    className={`px-3 py-1 rounded-full text-sm font-medium transition-colors ${
                      filters.tags.includes(tag)
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    }`}
                  >
                    #{tag}
                  </button>
                ))}
              </div>
            </div>

            {/* 材料選択 */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                使いたい材料
              </label>
              <div className="flex flex-wrap gap-2">
                {COMMON_INGREDIENTS.map(ingredient => (
                  <button
                    key={ingredient}
                    type="button"
                    onClick={() => handleIngredientToggle(ingredient)}
                    className={`px-3 py-1 rounded-full text-sm font-medium transition-colors ${
                      filters.ingredients.includes(ingredient)
                        ? 'bg-green-600 text-white'
                        : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                    }`}
                  >
                    🥘 {ingredient}
                  </button>
                ))}
              </div>
            </div>

            {/* 調理時間フィルター */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  最大調理時間
                </label>
                <select
                  value={filters.maxCookingTime || ''}
                  onChange={(e) => setFilters(prev => ({
                    ...prev,
                    maxCookingTime: e.target.value ? parseInt(e.target.value) : null
                  }))}
                  className="input-field"
                >
                  <option value="">指定しない</option>
                  <option value="5">5分以内</option>
                  <option value="10">10分以内</option>
                  <option value="15">15分以内</option>
                  <option value="30">30分以内</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  ずぼらレベル
                </label>
                <select
                  value={filters.minLazinessScore || ''}
                  onChange={(e) => setFilters(prev => ({
                    ...prev,
                    minLazinessScore: e.target.value ? parseInt(e.target.value) : null
                  }))}
                  className="input-field"
                >
                  <option value="">指定しない</option>
                  <option value="8">とっても簡単 (8点以上)</option>
                  <option value="6">まあまあ簡単 (6点以上)</option>
                  <option value="4">普通 (4点以上)</option>
                </select>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default RecipeSearch;