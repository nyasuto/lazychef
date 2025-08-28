import React from 'react';

const WeeklyPlanPage: React.FC = () => {
  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold text-gray-900">週間献立</h1>
      </div>
      
      <div className="card">
        <div className="text-center py-12">
          <div className="text-6xl mb-4">📅</div>
          <h3 className="text-xl font-medium text-gray-900 mb-2">週間献立機能</h3>
          <p className="text-gray-600 mb-6">
            Phase 4で実装予定です<br />
            1週間分の献立作成と買い物リスト生成を行います
          </p>
          <div className="text-sm text-gray-500">
            実装予定機能：
            <ul className="mt-2 space-y-1">
              <li>• 7日×3食のグリッドレイアウト</li>
              <li>• 自動献立生成</li>
              <li>• 材料使い回し最適化</li>
              <li>• 買い物リスト自動作成</li>
              <li>• 食費概算表示</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
};

export default WeeklyPlanPage;