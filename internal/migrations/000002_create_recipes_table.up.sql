CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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