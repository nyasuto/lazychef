import React from 'react';

const SettingsPage: React.FC = () => {
  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h1 className="text-3xl font-bold text-gray-900">設定</h1>
      </div>
      
      <div className="card">
        <div className="text-center py-12">
          <div className="text-6xl mb-4">⚙️</div>
          <h3 className="text-xl font-medium text-gray-900 mb-2">設定機能</h3>
          <p className="text-gray-600 mb-6">
            Phase 4で実装予定です<br />
            ユーザーの好みや制限事項を設定できます
          </p>
          <div className="text-sm text-gray-500">
            実装予定機能：
            <ul className="mt-2 space-y-1">
              <li>• 食材の好み・嫌い</li>
              <li>• アレルギー・食事制限</li>
              <li>• 調理スキルレベル</li>
              <li>• 予算設定</li>
              <li>• 通知設定</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SettingsPage;