-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255),
    email VARCHAR(255) UNIQUE,
    password TEXT,
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now(),
    deleted_at TIMESTAMP NULL
);

-- Create recipes table
CREATE TABLE IF NOT EXISTS recipes (
    id VARCHAR(255) PRIMARY KEY,
    title VARCHAR(255),
    ingredients TEXT[],
    steps TEXT[],
    nutritional_info TEXT,
    allergy_disclaimer TEXT,
    appliances TEXT[],
    embedding REAL[],
    created_at TIMESTAMP DEFAULT now(),
    updated_at TIMESTAMP DEFAULT now()
); 