import React, { useState } from 'react';
import type { ShoppingItem } from '../../types';

interface ShoppingListProps {
  items: ShoppingItem[];
  onItemToggle?: (index: number, checked: boolean) => void;
  onItemAdd?: (item: ShoppingItem) => void;
  onItemRemove?: (index: number) => void;
  readonly?: boolean;
  showAddForm?: boolean;
}

const ShoppingList: React.FC<ShoppingListProps> = ({
  items,
  onItemToggle,
  onItemAdd,
  onItemRemove,
  readonly = false,
  showAddForm = false
}) => {
  const [newItem, setNewItem] = useState({
    item: '',
    amount: '',
    category: '',
    cost: 0
  });
  const [showForm, setShowForm] = useState(false);

  // カテゴリー別にアイテムをグループ化
  const groupedItems = items.reduce((groups, item, index) => {
    const category = item.category || 'その他';
    if (!groups[category]) {
      groups[category] = [];
    }
    groups[category].push({ ...item, index });
    return groups;
  }, {} as Record<string, Array<ShoppingItem & { index: number }>>);

  // カテゴリーのアイコン
  const getCategoryIcon = (category: string) => {
    const icons: Record<string, string> = {
      '野菜': '🥬',
      '肉類': '🥩',
      '魚介類': '🐟',
      '乳製品': '🥛',
      '調味料': '🧂',
      '穀物': '🌾',
      '果物': '🍎',
      'パン': '🍞',
      'その他': '📦'
    };
    return icons[category] || '📦';
  };

  const handleItemToggle = (index: number, checked: boolean) => {
    if (onItemToggle && !readonly) {
      onItemToggle(index, checked);
    }
  };

  const handleAddItem = () => {
    if (newItem.item.trim() && onItemAdd) {
      onItemAdd({
        item: newItem.item,
        amount: newItem.amount,
        category: newItem.category || 'その他',
        cost: newItem.cost || undefined,
        checked: false
      });
      setNewItem({ item: '', amount: '', category: '', cost: 0 });
      setShowForm(false);
    }
  };

  const totalCost = items.reduce((sum, item) => sum + (item.cost || 0), 0);
  const completedItems = items.filter(item => item.checked).length;

  return (
    <div className="bg-white rounded-lg shadow-md overflow-hidden">
      {/* ヘッダー */}
      <div className="bg-secondary-600 text-white p-4">
        <div className="flex justify-between items-center">
          <div>
            <h2 className="text-xl font-semibold flex items-center gap-2">
              🛒 買い物リスト
            </h2>
            <p className="text-secondary-100 text-sm">
              {completedItems}/{items.length} 項目完了
            </p>
          </div>
          <div className="text-right">
            <div className="text-2xl font-bold">¥{totalCost.toLocaleString()}</div>
            <div className="text-secondary-100 text-sm">予算目安</div>
          </div>
        </div>

        {/* 進捗バー */}
        <div className="mt-4">
          <div className="bg-secondary-700 rounded-full h-2">
            <div 
              className="bg-white rounded-full h-2 transition-all duration-300"
              style={{ width: `${items.length > 0 ? (completedItems / items.length) * 100 : 0}%` }}
            />
          </div>
        </div>
      </div>

      <div className="p-4">
        {/* カテゴリー別リスト */}
        {Object.entries(groupedItems).map(([category, categoryItems]) => (
          <div key={category} className="mb-6">
            {/* カテゴリーヘッダー */}
            <div className="flex items-center gap-2 mb-3">
              <span className="text-lg">{getCategoryIcon(category)}</span>
              <h3 className="font-semibold text-gray-900">{category}</h3>
              <span className="text-sm text-gray-500">
                ({categoryItems.filter(item => !item.checked).length}件)
              </span>
            </div>

            {/* アイテムリスト */}
            <div className="space-y-2">
              {categoryItems.map((item) => (
                <div
                  key={item.index}
                  className={`flex items-center gap-3 p-3 rounded-lg border transition-all duration-200 ${
                    item.checked
                      ? 'bg-gray-50 border-gray-200 opacity-60'
                      : 'bg-white border-gray-300 hover:border-secondary-300'
                  }`}
                >
                  {/* チェックボックス */}
                  {!readonly && (
                    <input
                      type="checkbox"
                      checked={item.checked || false}
                      onChange={(e) => handleItemToggle(item.index, e.target.checked)}
                      className="w-5 h-5 text-secondary-600 bg-gray-100 border-gray-300 rounded focus:ring-secondary-500 focus:ring-2"
                    />
                  )}

                  {/* アイテム情報 */}
                  <div className="flex-1">
                    <div className="flex items-center justify-between">
                      <div className={`font-medium ${item.checked ? 'line-through text-gray-500' : 'text-gray-900'}`}>
                        {item.item}
                      </div>
                      {item.cost && (
                        <div className={`text-sm font-medium ${item.checked ? 'text-gray-400' : 'text-secondary-600'}`}>
                          ¥{item.cost}
                        </div>
                      )}
                    </div>
                    {item.amount && (
                      <div className={`text-sm ${item.checked ? 'text-gray-400' : 'text-gray-600'}`}>
                        {item.amount}
                      </div>
                    )}
                  </div>

                  {/* 削除ボタン */}
                  {!readonly && onItemRemove && (
                    <button
                      onClick={() => onItemRemove(item.index)}
                      className="text-gray-400 hover:text-red-500 transition-colors duration-200"
                    >
                      ✕
                    </button>
                  )}
                </div>
              ))}
            </div>
          </div>
        ))}

        {/* 空の状態 */}
        {items.length === 0 && (
          <div className="text-center py-8">
            <div className="text-4xl mb-4">📝</div>
            <div className="text-gray-500 mb-4">買い物リストが空です</div>
            {showAddForm && (
              <button
                onClick={() => setShowForm(true)}
                className="btn-secondary"
              >
                アイテムを追加
              </button>
            )}
          </div>
        )}

        {/* 追加フォーム */}
        {showAddForm && !readonly && (
          <div className="mt-6 pt-6 border-t">
            {!showForm ? (
              <button
                onClick={() => setShowForm(true)}
                className="btn-outline w-full"
              >
                + アイテムを追加
              </button>
            ) : (
              <div className="bg-gray-50 p-4 rounded-lg">
                <h4 className="font-medium mb-3">新しいアイテムを追加</h4>
                <div className="space-y-3">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    <input
                      type="text"
                      placeholder="アイテム名"
                      value={newItem.item}
                      onChange={(e) => setNewItem({ ...newItem, item: e.target.value })}
                      className="input-field"
                    />
                    <input
                      type="text"
                      placeholder="数量 (例: 300g, 2個)"
                      value={newItem.amount}
                      onChange={(e) => setNewItem({ ...newItem, amount: e.target.value })}
                      className="input-field"
                    />
                  </div>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                    <select
                      value={newItem.category}
                      onChange={(e) => setNewItem({ ...newItem, category: e.target.value })}
                      className="input-field"
                    >
                      <option value="">カテゴリーを選択</option>
                      <option value="野菜">🥬 野菜</option>
                      <option value="肉類">🥩 肉類</option>
                      <option value="魚介類">🐟 魚介類</option>
                      <option value="乳製品">🥛 乳製品</option>
                      <option value="調味料">🧂 調味料</option>
                      <option value="穀物">🌾 穀物</option>
                      <option value="果物">🍎 果物</option>
                      <option value="パン">🍞 パン</option>
                      <option value="その他">📦 その他</option>
                    </select>
                    <input
                      type="number"
                      placeholder="予想価格 (円)"
                      value={newItem.cost || ''}
                      onChange={(e) => setNewItem({ ...newItem, cost: Number(e.target.value) })}
                      className="input-field"
                    />
                  </div>
                  <div className="flex gap-3">
                    <button
                      onClick={handleAddItem}
                      className="btn-primary flex-1"
                    >
                      追加
                    </button>
                    <button
                      onClick={() => setShowForm(false)}
                      className="btn-outline px-6"
                    >
                      キャンセル
                    </button>
                  </div>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default ShoppingList;