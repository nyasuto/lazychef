import React from 'react';

const RecipeListPage: React.FC = () => {
  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold text-gray-900">レシピ一覧</h1>
      </div>
      
      <div className="card">
        <div className="text-center py-12">
          <div className="text-6xl mb-4">🔍</div>
          <h3 className="text-xl font-medium text-gray-900 mb-2">レシピ検索機能</h3>
          <p className="text-gray-600 mb-6">
            Phase 2で実装予定です<br />
            レシピの検索・フィルター・表示機能を追加します
          </p>
          <div className="text-sm text-gray-500">
            実装予定機能：
            <ul className="mt-2 space-y-1">
              <li>• タグ・材料による検索</li>
              <li>• 調理時間・ずぼらスコアフィルター</li>
              <li>• レシピカード表示</li>
              <li>• ページネーション</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
};

export default RecipeListPage;