-- 階層的材料分類システム用スキーマ
-- Issue #87: 材料検索の精度向上のための階層構造

-- 材料グループテーブル（大分類・中分類）
CREATE TABLE ingredient_groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,           -- グループ名（例：肉類、鶏肉、野菜）
    display_name TEXT NOT NULL,          -- 表示名（日本語）
    parent_id INTEGER,                   -- 親グループID（階層構造用）
    level INTEGER NOT NULL DEFAULT 1,    -- 階層レベル（1=大分類、2=中分類）
    sort_order INTEGER DEFAULT 0,        -- 表示順序
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- 外部キー制約
    FOREIGN KEY (parent_id) REFERENCES ingredient_groups(id),
    
    -- チェック制約
    CHECK (level >= 1 AND level <= 3),
    CHECK (sort_order >= 0)
);

-- 具体的材料テーブル
CREATE TABLE specific_ingredients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,           -- 具体的材料名（例：鶏胸肉、豚こま肉）
    display_name TEXT NOT NULL,          -- 表示名
    aliases TEXT,                        -- 別名（JSON配列形式）
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- バリデーション
    CHECK (name != ''),
    CHECK (display_name != ''),
    CHECK (aliases IS NULL OR json_valid(aliases))
);

-- 材料グループと具体的材料のマッピング
CREATE TABLE ingredient_group_mappings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ingredient_id INTEGER NOT NULL,      -- specific_ingredients.id
    group_id INTEGER NOT NULL,           -- ingredient_groups.id
    primary_group BOOLEAN DEFAULT FALSE, -- 主要分類かどうか
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- 外部キー制約
    FOREIGN KEY (ingredient_id) REFERENCES specific_ingredients(id) ON DELETE CASCADE,
    FOREIGN KEY (group_id) REFERENCES ingredient_groups(id) ON DELETE CASCADE,
    
    -- ユニーク制約（同じ材料と同じグループの組み合わせは1つまで）
    UNIQUE (ingredient_id, group_id)
);

-- インデックス作成
CREATE INDEX idx_ingredient_groups_parent ON ingredient_groups(parent_id);
CREATE INDEX idx_ingredient_groups_level ON ingredient_groups(level);
CREATE INDEX idx_specific_ingredients_name ON specific_ingredients(name);
CREATE INDEX idx_ingredient_mappings_ingredient ON ingredient_group_mappings(ingredient_id);
CREATE INDEX idx_ingredient_mappings_group ON ingredient_group_mappings(group_id);
CREATE INDEX idx_ingredient_mappings_primary ON ingredient_group_mappings(primary_group) WHERE primary_group = TRUE;

-- 初期データ投入

-- 大分類（レベル1）
INSERT INTO ingredient_groups (name, display_name, level, sort_order) VALUES
    ('meat', '肉類', 1, 1),
    ('vegetables', '野菜', 1, 2),
    ('seafood', '魚介類', 1, 3),
    ('grains', '穀物・麺類', 1, 4),
    ('dairy_eggs', '卵・乳製品', 1, 5),
    ('seasonings', '調味料', 1, 6),
    ('others', 'その他', 1, 7);

-- 中分類（レベル2）- 肉類
INSERT INTO ingredient_groups (name, display_name, parent_id, level, sort_order) VALUES
    ('beef', '牛肉', (SELECT id FROM ingredient_groups WHERE name = 'meat'), 2, 1),
    ('pork', '豚肉', (SELECT id FROM ingredient_groups WHERE name = 'meat'), 2, 2),
    ('chicken', '鶏肉', (SELECT id FROM ingredient_groups WHERE name = 'meat'), 2, 3);

-- 中分類（レベル2）- 野菜
INSERT INTO ingredient_groups (name, display_name, parent_id, level, sort_order) VALUES
    ('root_vegetables', '根菜類', (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), 2, 1),
    ('leafy_vegetables', '葉菜類', (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), 2, 2),
    ('fruit_vegetables', '果菜類', (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), 2, 3);

-- 中分類（レベル2）- 魚介類
INSERT INTO ingredient_groups (name, display_name, parent_id, level, sort_order) VALUES
    ('fish', '魚類', (SELECT id FROM ingredient_groups WHERE name = 'seafood'), 2, 1),
    ('canned_seafood', '缶詰・加工品', (SELECT id FROM ingredient_groups WHERE name = 'seafood'), 2, 2);

-- 中分類（レベル2）- 穀物・麺類
INSERT INTO ingredient_groups (name, display_name, parent_id, level, sort_order) VALUES
    ('rice', '米・ご飯', (SELECT id FROM ingredient_groups WHERE name = 'grains'), 2, 1),
    ('noodles', '麺類', (SELECT id FROM ingredient_groups WHERE name = 'grains'), 2, 2);

-- 中分類（レベル2）- 調味料
INSERT INTO ingredient_groups (name, display_name, parent_id, level, sort_order) VALUES
    ('basic_seasonings', '基本調味料', (SELECT id FROM ingredient_groups WHERE name = 'seasonings'), 2, 1),
    ('oils', '油類', (SELECT id FROM ingredient_groups WHERE name = 'seasonings'), 2, 2);

-- 具体的材料データ
INSERT INTO specific_ingredients (name, display_name, aliases) VALUES
    -- 肉類
    ('牛切り落とし', '牛切り落とし', '["牛肉切り落とし", "牛こま肉"]'),
    ('豚こま肉', '豚こま肉', '["豚切り落とし", "豚肉切り落とし"]'),
    ('鶏胸肉', '鶏胸肉', '["鶏むね肉", "チキンブレスト"]'),
    
    -- 野菜
    ('玉ねぎ', '玉ねぎ', '["タマネギ", "オニオン"]'),
    ('人参', '人参', '["にんじん", "ニンジン", "キャロット"]'),
    ('じゃがいも', 'じゃがいも', '["ジャガイモ", "ポテト"]'),
    ('もやし', 'もやし', '["モヤシ", "豆もやし"]'),
    ('キャベツ', 'キャベツ', '["きゃべつ"]'),
    ('白菜', '白菜', '["はくさい", "ハクサイ"]'),
    ('レタス', 'レタス', '["れたす"]'),
    ('きゅうり', 'きゅうり', '["キュウリ", "胡瓜"]'),
    ('トマト', 'トマト', '["とまと"]'),
    ('ピーマン', 'ピーマン', '["ぴーまん"]'),
    ('ねぎ', 'ねぎ', '["ネギ", "青ねぎ", "青ネギ", "長ねぎ"]'),
    ('にんにく', 'にんにく', '["ニンニク", "ガーリック"]'),
    
    -- 魚介類
    ('鮭', '鮭', '["さけ", "サケ", "サーモン"]'),
    ('ツナ缶', 'ツナ缶', '["ツナ", "まぐろ缶"]'),
    ('わかめ', 'わかめ', '["ワカメ", "若布"]'),
    
    -- 穀物・麺類
    ('ご飯', 'ご飯', '["米", "白米", "ライス"]'),
    ('パスタ', 'パスタ', '["スパゲッティ", "スパゲティ"]'),
    ('うどん', 'うどん', '["ウドン"]'),
    
    -- 卵・乳製品
    ('卵', '卵', '["たまご", "玉子", "エッグ"]'),
    ('豆腐', '豆腐', '["とうふ", "トウフ"]'),
    
    -- 調味料
    ('塩', '塩', '["しお", "食塩"]'),
    ('胡椒', '胡椒', '["こしょう", "コショウ", "ペッパー"]'),
    ('しょうゆ', 'しょうゆ', '["醤油", "ショウユ", "しょう油"]'),
    ('味噌', '味噌', '["みそ", "ミソ"]'),
    ('砂糖', '砂糖', '["さとう", "サトウ", "シュガー"]'),
    ('みりん', 'みりん', '["ミリン", "本みりん"]'),
    ('酒', '酒', '["日本酒", "料理酒", "清酒"]'),
    ('サラダ油', 'サラダ油', '["サラダあぶら", "植物油"]'),
    ('ごま油', 'ごま油', '["ゴマ油", "胡麻油", "セサミオイル"]');

-- 材料グループマッピング
-- 肉類のマッピング
INSERT INTO ingredient_group_mappings (ingredient_id, group_id, primary_group) VALUES
    -- 牛肉
    ((SELECT id FROM specific_ingredients WHERE name = '牛切り落とし'), 
     (SELECT id FROM ingredient_groups WHERE name = 'beef'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = '牛切り落とし'), 
     (SELECT id FROM ingredient_groups WHERE name = 'meat'), FALSE),
    
    -- 豚肉
    ((SELECT id FROM specific_ingredients WHERE name = '豚こま肉'), 
     (SELECT id FROM ingredient_groups WHERE name = 'pork'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = '豚こま肉'), 
     (SELECT id FROM ingredient_groups WHERE name = 'meat'), FALSE),
    
    -- 鶏肉
    ((SELECT id FROM specific_ingredients WHERE name = '鶏胸肉'), 
     (SELECT id FROM ingredient_groups WHERE name = 'chicken'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = '鶏胸肉'), 
     (SELECT id FROM ingredient_groups WHERE name = 'meat'), FALSE);

-- 野菜のマッピング
INSERT INTO ingredient_group_mappings (ingredient_id, group_id, primary_group) VALUES
    -- 根菜類
    ((SELECT id FROM specific_ingredients WHERE name = '玉ねぎ'), 
     (SELECT id FROM ingredient_groups WHERE name = 'root_vegetables'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = '玉ねぎ'), 
     (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = '人参'), 
     (SELECT id FROM ingredient_groups WHERE name = 'root_vegetables'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = '人参'), 
     (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'じゃがいも'), 
     (SELECT id FROM ingredient_groups WHERE name = 'root_vegetables'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'じゃがいも'), 
     (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), FALSE),
    
    -- 葉菜類
    ((SELECT id FROM specific_ingredients WHERE name = 'キャベツ'), 
     (SELECT id FROM ingredient_groups WHERE name = 'leafy_vegetables'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'キャベツ'), 
     (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = '白菜'), 
     (SELECT id FROM ingredient_groups WHERE name = 'leafy_vegetables'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = '白菜'), 
     (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'レタス'), 
     (SELECT id FROM ingredient_groups WHERE name = 'leafy_vegetables'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'レタス'), 
     (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'もやし'), 
     (SELECT id FROM ingredient_groups WHERE name = 'leafy_vegetables'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'もやし'), 
     (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), FALSE),
    
    -- 果菜類
    ((SELECT id FROM specific_ingredients WHERE name = 'トマト'), 
     (SELECT id FROM ingredient_groups WHERE name = 'fruit_vegetables'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'トマト'), 
     (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'ピーマン'), 
     (SELECT id FROM ingredient_groups WHERE name = 'fruit_vegetables'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'ピーマン'), 
     (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'きゅうり'), 
     (SELECT id FROM ingredient_groups WHERE name = 'fruit_vegetables'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'きゅうり'), 
     (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'ねぎ'), 
     (SELECT id FROM ingredient_groups WHERE name = 'fruit_vegetables'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'ねぎ'), 
     (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'にんにく'), 
     (SELECT id FROM ingredient_groups WHERE name = 'fruit_vegetables'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'にんにく'), 
     (SELECT id FROM ingredient_groups WHERE name = 'vegetables'), FALSE);

-- 魚介類のマッピング
INSERT INTO ingredient_group_mappings (ingredient_id, group_id, primary_group) VALUES
    ((SELECT id FROM specific_ingredients WHERE name = '鮭'), 
     (SELECT id FROM ingredient_groups WHERE name = 'fish'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = '鮭'), 
     (SELECT id FROM ingredient_groups WHERE name = 'seafood'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'ツナ缶'), 
     (SELECT id FROM ingredient_groups WHERE name = 'canned_seafood'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'ツナ缶'), 
     (SELECT id FROM ingredient_groups WHERE name = 'seafood'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'わかめ'), 
     (SELECT id FROM ingredient_groups WHERE name = 'seafood'), TRUE);

-- 穀物・麺類のマッピング
INSERT INTO ingredient_group_mappings (ingredient_id, group_id, primary_group) VALUES
    ((SELECT id FROM specific_ingredients WHERE name = 'ご飯'), 
     (SELECT id FROM ingredient_groups WHERE name = 'rice'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'ご飯'), 
     (SELECT id FROM ingredient_groups WHERE name = 'grains'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'パスタ'), 
     (SELECT id FROM ingredient_groups WHERE name = 'noodles'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'パスタ'), 
     (SELECT id FROM ingredient_groups WHERE name = 'grains'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'うどん'), 
     (SELECT id FROM ingredient_groups WHERE name = 'noodles'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'うどん'), 
     (SELECT id FROM ingredient_groups WHERE name = 'grains'), FALSE);

-- 卵・乳製品のマッピング
INSERT INTO ingredient_group_mappings (ingredient_id, group_id, primary_group) VALUES
    ((SELECT id FROM specific_ingredients WHERE name = '卵'), 
     (SELECT id FROM ingredient_groups WHERE name = 'dairy_eggs'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = '豆腐'), 
     (SELECT id FROM ingredient_groups WHERE name = 'dairy_eggs'), TRUE);

-- 調味料のマッピング
INSERT INTO ingredient_group_mappings (ingredient_id, group_id, primary_group) VALUES
    ((SELECT id FROM specific_ingredients WHERE name = '塩'), 
     (SELECT id FROM ingredient_groups WHERE name = 'basic_seasonings'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = '塩'), 
     (SELECT id FROM ingredient_groups WHERE name = 'seasonings'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = '胡椒'), 
     (SELECT id FROM ingredient_groups WHERE name = 'basic_seasonings'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = '胡椒'), 
     (SELECT id FROM ingredient_groups WHERE name = 'seasonings'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'しょうゆ'), 
     (SELECT id FROM ingredient_groups WHERE name = 'basic_seasonings'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'しょうゆ'), 
     (SELECT id FROM ingredient_groups WHERE name = 'seasonings'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = '味噌'), 
     (SELECT id FROM ingredient_groups WHERE name = 'basic_seasonings'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = '味噌'), 
     (SELECT id FROM ingredient_groups WHERE name = 'seasonings'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = '砂糖'), 
     (SELECT id FROM ingredient_groups WHERE name = 'basic_seasonings'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = '砂糖'), 
     (SELECT id FROM ingredient_groups WHERE name = 'seasonings'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'みりん'), 
     (SELECT id FROM ingredient_groups WHERE name = 'basic_seasonings'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'みりん'), 
     (SELECT id FROM ingredient_groups WHERE name = 'seasonings'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = '酒'), 
     (SELECT id FROM ingredient_groups WHERE name = 'basic_seasonings'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = '酒'), 
     (SELECT id FROM ingredient_groups WHERE name = 'seasonings'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'サラダ油'), 
     (SELECT id FROM ingredient_groups WHERE name = 'oils'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'サラダ油'), 
     (SELECT id FROM ingredient_groups WHERE name = 'seasonings'), FALSE),
    ((SELECT id FROM specific_ingredients WHERE name = 'ごま油'), 
     (SELECT id FROM ingredient_groups WHERE name = 'oils'), TRUE),
    ((SELECT id FROM specific_ingredients WHERE name = 'ごま油'), 
     (SELECT id FROM ingredient_groups WHERE name = 'seasonings'), FALSE);