CREATE TABLE IF NOT EXISTS recipes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    ingredients JSONB NOT NULL,
    steps JSONB NOT NULL,
    nutritional_info TEXT,
    allergy_disclaimer TEXT,
    embedding JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    images JSONB,
    average_rating DECIMAL(3,2) DEFAULT 0,
    rating_count INTEGER DEFAULT 0,
    difficulty VARCHAR(50),
    prep_time INTEGER,
    cook_time INTEGER,
    approved BOOLEAN DEFAULT FALSE
);

-- Create junction tables for many-to-many relationships
CREATE TABLE IF NOT EXISTS recipe_cuisines (
    recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
    cuisine_id UUID REFERENCES cuisines(id) ON DELETE CASCADE,
    PRIMARY KEY (recipe_id, cuisine_id)
);

CREATE TABLE IF NOT EXISTS recipe_diets (
    recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
    diet_id UUID REFERENCES diets(id) ON DELETE CASCADE,
    PRIMARY KEY (recipe_id, diet_id)
);

CREATE TABLE IF NOT EXISTS recipe_appliances (
    recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
    appliance_id UUID REFERENCES appliances(id) ON DELETE CASCADE,
    PRIMARY KEY (recipe_id, appliance_id)
);

CREATE TABLE IF NOT EXISTS recipe_tags (
    recipe_id UUID REFERENCES recipes(id) ON DELETE CASCADE,
    tag_id UUID REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (recipe_id, tag_id)
); 