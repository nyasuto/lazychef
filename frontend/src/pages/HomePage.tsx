import React from 'react';
import { Link } from 'react-router-dom';
import Button from '../components/common/Button';

const HomePage: React.FC = () => {
  return (
    <div className="space-y-8">
      {/* Hero Section */}
      <div className="text-center">
        <h1 className="text-4xl md:text-6xl font-bold text-gray-900 mb-4">
          🍳 LazyChef
        </h1>
        <p className="text-xl text-gray-600 mb-8 max-w-2xl mx-auto">
          ずぼらな人のためのAI料理アシスタント<br />
          簡単で美味しい献立を自動生成します
        </p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Link to="/recipes">
            <Button variant="primary" size="lg" className="w-full sm:w-auto">
              レシピを探す
            </Button>
          </Link>
          <Link to="/meal-plan">
            <Button variant="outline" size="lg" className="w-full sm:w-auto">
              献立を作成
            </Button>
          </Link>
        </div>
      </div>

      {/* Features Section */}
      <div className="grid md:grid-cols-3 gap-6 mt-16">
        <div className="card text-center">
          <div className="text-4xl mb-4">⚡</div>
          <h3 className="text-xl font-semibold mb-2">簡単・時短</h3>
          <p className="text-gray-600">
            最大15分で作れる<br />
            簡単なレシピのみ
          </p>
        </div>
        
        <div className="card text-center">
          <div className="text-4xl mb-4">🛒</div>
          <h3 className="text-xl font-semibold mb-2">買い物リスト</h3>
          <p className="text-gray-600">
            材料を効率よく使い回し<br />
            買い物リストも自動生成
          </p>
        </div>
        
        <div className="card text-center">
          <div className="text-4xl mb-4">🤖</div>
          <h3 className="text-xl font-semibold mb-2">AI生成</h3>
          <p className="text-gray-600">
            お好みの材料から<br />
            オリジナルレシピを作成
          </p>
        </div>
      </div>

      {/* Getting Started Section */}
      <div className="card">
        <h2 className="text-2xl font-bold mb-4">使い方</h2>
        <div className="space-y-4">
          <div className="flex items-start space-x-3">
            <div className="flex-shrink-0 w-8 h-8 bg-primary-100 text-primary-600 rounded-full flex items-center justify-center font-bold">
              1
            </div>
            <div>
              <h4 className="font-medium">レシピを探す or 生成する</h4>
              <p className="text-gray-600">既存のレシピから選ぶか、AIに新しいレシピを作ってもらいましょう</p>
            </div>
          </div>
          
          <div className="flex items-start space-x-3">
            <div className="flex-shrink-0 w-8 h-8 bg-primary-100 text-primary-600 rounded-full flex items-center justify-center font-bold">
              2
            </div>
            <div>
              <h4 className="font-medium">週間献立を作成</h4>
              <p className="text-gray-600">1週間分の献立を自動で組み立てます</p>
            </div>
          </div>
          
          <div className="flex items-start space-x-3">
            <div className="flex-shrink-0 w-8 h-8 bg-primary-100 text-primary-600 rounded-full flex items-center justify-center font-bold">
              3
            </div>
            <div>
              <h4 className="font-medium">買い物リストをゲット</h4>
              <p className="text-gray-600">材料がまとめられた買い物リストで効率的にお買い物</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default HomePage;