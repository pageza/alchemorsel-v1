-- Drop junction tables first (due to foreign key constraints)
DROP TABLE IF EXISTS recipe_tags;
DROP TABLE IF EXISTS recipe_appliances;
DROP TABLE IF EXISTS recipe_diets;
DROP TABLE IF EXISTS recipe_cuisines;

-- Drop main tables
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS appliances;
DROP TABLE IF EXISTS diets;
DROP TABLE IF EXISTS cuisines; 