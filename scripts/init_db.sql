-- LazyChef Database Schema
-- SQLite with JSON1 extension for flexible data storage

-- Enable foreign key support
PRAGMA foreign_keys = ON;

-- Drop tables if they exist (for development)
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