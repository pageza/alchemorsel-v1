-- Drop triggers
DROP TRIGGER IF EXISTS update_recipes_updated_at ON recipes;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_recipes_title;
DROP INDEX IF EXISTS idx_recipes_created_by;
DROP INDEX IF EXISTS idx_users_email;

-- Drop tables
DROP TABLE IF EXISTS recipes;
DROP TABLE IF EXISTS users; 