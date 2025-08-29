-- LazyChef Database Schema
-- SQLite with JSON1 extension for flexible data storage

-- Enable foreign key support
PRAGMA foreign_keys = ON;

-- Drop tables if they exist (for development)
DROP TABLE IF EXISTS duplicate_detection_results;
DROP TABLE IF EXISTS recipe_embeddings;
DROP TABLE IF EXISTS recipe_generation_jobs;
DROP TABLE IF EXISTS meal_plans;
DROP TABLE IF EXISTS user_preferences; 
DROP TABLE IF EXISTS recipes;

-- Recipes table with JSON data storage
CREATE TABLE recipes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    data JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Virtual columns for efficient querying
    title TEXT GENERATED ALWAYS AS (json_extract(data, '$.title')) STORED,
    cooking_time INTEGER GENERATED ALWAYS AS (json_extract(data, '$.cooking_time')) STORED,
    laziness_score REAL GENERATED ALWAYS AS (json_extract(data, '$.laziness_score')) STORED,
    season TEXT GENERATED ALWAYS AS (json_extract(data, '$.season')) STORED,
    
    -- Constraints
    CHECK (json_valid(data)),
    CHECK (cooking_time > 0),
    CHECK (laziness_score >= 1.0 AND laziness_score <= 10.0)
);

-- Weekly meal plans table
CREATE TABLE meal_plans (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    week_data JSON NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Virtual columns for querying
    start_date DATE GENERATED ALWAYS AS (json_extract(week_data, '$.start_date')) STORED,
    total_cost_estimate REAL GENERATED ALWAYS AS (json_extract(week_data, '$.total_cost_estimate')) STORED,
    
    -- Constraints
    CHECK (json_valid(week_data)),
    CHECK (total_cost_estimate >= 0)
);

-- User preferences table (for future personalization)
CREATE TABLE user_preferences (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT DEFAULT 'default_user', -- For single-user MVP
    preferences JSON NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Virtual columns
    max_cooking_time INTEGER GENERATED ALWAYS AS (json_extract(preferences, '$.max_cooking_time')) STORED,
    
    -- Constraints
    CHECK (json_valid(preferences)),
    UNIQUE(user_id)
);

-- Indexes for performance
-- Recipe indexes
CREATE INDEX idx_recipes_title ON recipes(title);
CREATE INDEX idx_recipes_cooking_time ON recipes(cooking_time);
CREATE INDEX idx_recipes_laziness_score ON recipes(laziness_score);
CREATE INDEX idx_recipes_season ON recipes(season);
CREATE INDEX idx_recipes_created_at ON recipes(created_at);

-- Tag search index (for JSON array search)
CREATE INDEX idx_recipes_tags ON recipes(json_extract(data, '$.tags'));

-- Meal plan indexes
CREATE INDEX idx_meal_plans_start_date ON meal_plans(start_date);
CREATE INDEX idx_meal_plans_created_at ON meal_plans(created_at);

-- User preferences index
CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);

-- Phase 1: Batch API & Embedding Tables

-- Batch job management table
CREATE TABLE recipe_generation_jobs (
    id TEXT PRIMARY KEY,                    -- UUID
    batch_type TEXT NOT NULL,               -- 'sync', 'batch_api'
    config JSON NOT NULL,                   -- generation parameters
    model_info JSON,                        -- model, seed, system_fingerprint
    cost_data JSON,                         -- tokens, $ spent
    status TEXT NOT NULL,                   -- 'pending', 'submitted', 'completed', 'failed'
    batch_id TEXT,                          -- OpenAI Batch API ID
    submitted_at DATETIME,
    completed_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CHECK (json_valid(config)),
    CHECK (json_valid(model_info)),
    CHECK (json_valid(cost_data)),
    CHECK (status IN ('pending', 'submitted', 'completed', 'failed', 'cancelled')),
    CHECK (batch_type IN ('sync', 'batch_api'))
);

-- Recipe embeddings for similarity detection
CREATE TABLE recipe_embeddings (
    recipe_id INTEGER PRIMARY KEY,
    embedding_version TEXT NOT NULL,        -- 'v3', etc
    content_hash TEXT NOT NULL,             -- content change detection
    embedding BLOB NOT NULL,                -- vector data
    dimensions INTEGER NOT NULL,            -- 1536 for v3
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE CASCADE,
    CHECK (dimensions > 0)
);

-- Duplicate detection results
CREATE TABLE duplicate_detection_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    recipe_id INTEGER NOT NULL,
    similar_recipe_id INTEGER NOT NULL,
    similarity_score REAL NOT NULL,         -- cosine similarity [0,1]
    jaccard_score REAL,                     -- ingredient overlap [0,1]
    detection_method TEXT NOT NULL,         -- 'embedding', 'jaccard', 'combined'
    detected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (recipe_id) REFERENCES recipes(id) ON DELETE CASCADE,
    FOREIGN KEY (similar_recipe_id) REFERENCES recipes(id) ON DELETE CASCADE,
    CHECK (similarity_score >= 0 AND similarity_score <= 1),
    CHECK (jaccard_score IS NULL OR (jaccard_score >= 0 AND jaccard_score <= 1)),
    CHECK (detection_method IN ('embedding', 'jaccard', 'combined')),
    CHECK (recipe_id != similar_recipe_id)
);

-- Phase 1: Indexes for new tables

-- Batch job indexes
CREATE INDEX idx_batch_jobs_status ON recipe_generation_jobs(status);
CREATE INDEX idx_batch_jobs_batch_type ON recipe_generation_jobs(batch_type);
CREATE INDEX idx_batch_jobs_batch_id ON recipe_generation_jobs(batch_id);
CREATE INDEX idx_batch_jobs_created_at ON recipe_generation_jobs(created_at);

-- Embedding indexes
CREATE INDEX idx_embeddings_version ON recipe_embeddings(embedding_version);
CREATE INDEX idx_embeddings_hash ON recipe_embeddings(content_hash);
CREATE INDEX idx_embeddings_created_at ON recipe_embeddings(created_at);

-- Duplicate detection indexes
CREATE INDEX idx_duplicates_recipe_id ON duplicate_detection_results(recipe_id);
CREATE INDEX idx_duplicates_similar_id ON duplicate_detection_results(similar_recipe_id);
CREATE INDEX idx_duplicates_similarity_score ON duplicate_detection_results(similarity_score);
CREATE INDEX idx_duplicates_method ON duplicate_detection_results(detection_method);
CREATE INDEX idx_duplicates_detected_at ON duplicate_detection_results(detected_at);

-- Phase 2: Recipe Diversity System Tables (Issue #65)

-- レシピ次元定義テーブル
CREATE TABLE recipe_dimensions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    dimension_type TEXT NOT NULL,            -- 'meal_type', 'staple', 'protein', etc.
    dimension_value TEXT NOT NULL,           -- '朝食', '米', '鶏肉', etc.
    weight REAL DEFAULT 1.0,                 -- Generation priority weight
    is_active BOOLEAN DEFAULT 1,             -- Enable/disable dimension
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(dimension_type, dimension_value),
    CHECK (weight >= 0),
    CHECK (is_active IN (0, 1))
);

-- カバレッジ追跡テーブル
CREATE TABLE dimension_coverage (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    dimension_combo TEXT NOT NULL,           -- JSON string of dimension combination
    current_count INTEGER DEFAULT 0,         -- Current recipe count for this combo
    target_count INTEGER DEFAULT 5,          -- Target recipe count
    priority_score REAL DEFAULT 1.0,         -- Coverage priority (lower = higher priority)
    last_generated_at DATETIME,              -- Last generation timestamp
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(dimension_combo),
    CHECK (current_count >= 0),
    CHECK (target_count > 0),
    CHECK (priority_score >= 0),
    CHECK (json_valid(dimension_combo))
);

-- 生成プロファイル（パフォーマンス追跡）
CREATE TABLE generation_profiles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    profile_name TEXT NOT NULL,              -- 'diverse', 'targeted', 'random', etc.
    config JSON NOT NULL,                    -- Generation parameters
    performance_data JSON,                   -- Success rates, costs, etc.
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    UNIQUE(profile_name),
    CHECK (json_valid(config)),
    CHECK (json_valid(performance_data)),
    CHECK (is_active IN (0, 1))
);

-- Phase 2: Indexes for diversity system

-- Dimension indexes
CREATE INDEX idx_dimensions_type ON recipe_dimensions(dimension_type);
CREATE INDEX idx_dimensions_value ON recipe_dimensions(dimension_value);
CREATE INDEX idx_dimensions_active ON recipe_dimensions(is_active);
CREATE INDEX idx_dimensions_weight ON recipe_dimensions(weight);

-- Coverage indexes  
CREATE INDEX idx_coverage_combo ON dimension_coverage(dimension_combo);
CREATE INDEX idx_coverage_current_count ON dimension_coverage(current_count);
CREATE INDEX idx_coverage_priority ON dimension_coverage(priority_score);
CREATE INDEX idx_coverage_last_generated ON dimension_coverage(last_generated_at);

-- Profile indexes
CREATE INDEX idx_profiles_name ON generation_profiles(profile_name);
CREATE INDEX idx_profiles_active ON generation_profiles(is_active);

-- Insert default user preferences
INSERT INTO user_preferences (user_id, preferences) VALUES (
    'default_user',
    json('{
        "max_cooking_time": 15,
        "exclude_ingredients": [],
        "preferred_tags": ["簡単", "10分以内", "ずぼら"],
        "budget_per_week": 3000,
        "household_size": 1,
        "dietary_restrictions": [],
        "preferred_seasons": ["all"]
    }')
);

-- Insert initial recipe dimensions (Issue #65 Phase 1)

-- 食事タイプ 
INSERT INTO recipe_dimensions (dimension_type, dimension_value, weight) VALUES
    ('meal_type', '朝食', 1.2),
    ('meal_type', '昼食', 1.0),
    ('meal_type', '夕食', 1.0),
    ('meal_type', 'おやつ', 0.8),
    ('meal_type', '夜食', 0.6);

-- 主食
INSERT INTO recipe_dimensions (dimension_type, dimension_value, weight) VALUES
    ('staple', '米', 1.5),
    ('staple', 'うどん', 1.2),
    ('staple', 'そば', 1.0),
    ('staple', 'パン', 1.0),
    ('staple', 'なし', 0.8);

-- タンパク質
INSERT INTO recipe_dimensions (dimension_type, dimension_value, weight) VALUES
    ('protein', '鶏肉', 1.3),
    ('protein', '豚肉', 1.2),
    ('protein', '卵', 1.5),
    ('protein', '豆腐', 1.1),
    ('protein', 'ツナ缶', 1.0),
    ('protein', 'なし', 0.7);

-- 調理法
INSERT INTO recipe_dimensions (dimension_type, dimension_value, weight) VALUES
    ('cooking_method', '電子レンジ', 1.5),
    ('cooking_method', '炒める', 1.2),
    ('cooking_method', '煮る', 1.0),
    ('cooking_method', '焼く', 1.0),
    ('cooking_method', '和えるだけ', 1.8);

-- 味付け
INSERT INTO recipe_dimensions (dimension_type, dimension_value, weight) VALUES
    ('seasoning', '醤油系', 1.3),
    ('seasoning', '味噌系', 1.1),
    ('seasoning', '塩系', 1.0),
    ('seasoning', '甘辛', 1.2),
    ('seasoning', 'さっぱり', 1.0);

-- ラジネス度 (3段階)
INSERT INTO recipe_dimensions (dimension_type, dimension_value, weight) VALUES
    ('laziness_level', '1_超簡単', 2.0),
    ('laziness_level', '2_簡単', 1.2),
    ('laziness_level', '3_ちょい手間', 0.8);

-- Insert default generation profiles
INSERT INTO generation_profiles (profile_name, config, performance_data) VALUES
    ('diverse', json('{
        "strategy": "coverage_first",
        "batch_size": 10,
        "max_similarity": 0.85,
        "quality_threshold": 7.0,
        "dimension_weights": {
            "meal_type": 1.0,
            "staple": 1.0, 
            "protein": 1.0,
            "cooking_method": 1.2,
            "seasoning": 0.8,
            "laziness_level": 1.5
        }
    }'), json('{"success_rate": 0.0, "avg_cost_per_recipe": 0.0, "total_generated": 0}')),
    
    ('targeted', json('{
        "strategy": "priority_first",
        "batch_size": 5,
        "max_similarity": 0.90,
        "quality_threshold": 8.0,
        "focus_dimensions": ["cooking_method", "laziness_level"]
    }'), json('{"success_rate": 0.0, "avg_cost_per_recipe": 0.0, "total_generated": 0}')),
    
    ('random', json('{
        "strategy": "random_sample",
        "batch_size": 20,
        "max_similarity": 0.80,
        "quality_threshold": 6.0
    }'), json('{"success_rate": 0.0, "avg_cost_per_recipe": 0.0, "total_generated": 0}'));

-- Create trigger to update updated_at timestamp
CREATE TRIGGER update_recipes_timestamp 
    AFTER UPDATE ON recipes
    FOR EACH ROW
BEGIN
    UPDATE recipes SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_meal_plans_timestamp 
    AFTER UPDATE ON meal_plans
    FOR EACH ROW
BEGIN
    UPDATE meal_plans SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER update_user_preferences_timestamp 
    AFTER UPDATE ON user_preferences
    FOR EACH ROW
BEGIN
    UPDATE user_preferences SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Diversity system triggers (Issue #65)

-- Update dimension_coverage timestamp
CREATE TRIGGER update_dimension_coverage_timestamp 
    AFTER UPDATE ON dimension_coverage
    FOR EACH ROW
BEGIN
    UPDATE dimension_coverage SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Update generation_profiles timestamp
CREATE TRIGGER update_generation_profiles_timestamp 
    AFTER UPDATE ON generation_profiles
    FOR EACH ROW
BEGIN
    UPDATE generation_profiles SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;