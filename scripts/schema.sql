-- Bonpreu Database Schema
-- Run this script to create the necessary tables

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    product_id INTEGER PRIMARY KEY,
    product_type VARCHAR(255),
    product_name VARCHAR(500) NOT NULL,
    product_description TEXT,
    product_brand VARCHAR(255),
    product_pack_size_description VARCHAR(255),
    product_price_amount DECIMAL(10,2),
    product_currency VARCHAR(10),
    product_unit_price_amount DECIMAL(10,2),
    product_unit_price_currency VARCHAR(10),
    product_unit_price_unit VARCHAR(50),
    product_available BOOLEAN DEFAULT false,
    product_alcohol BOOLEAN DEFAULT false,
    product_cooking_guidelines TEXT,
    product_categories TEXT[], -- Array of category strings
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create product_nutritional_data table
CREATE TABLE IF NOT EXISTS product_nutritional_data (
    id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL,
    product_nutritional_value VARCHAR(255),
    product_nutritional_quantity VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE CASCADE
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_products_product_id ON products(product_id);
CREATE INDEX IF NOT EXISTS idx_products_product_name ON products(product_name);
CREATE INDEX IF NOT EXISTS idx_products_product_brand ON products(product_brand);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);

CREATE INDEX IF NOT EXISTS idx_product_nutritional_data_product_id ON product_nutritional_data(product_id);
CREATE INDEX IF NOT EXISTS idx_product_nutritional_data_created_at ON product_nutritional_data(created_at);

-- Create updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Create trigger for products table
CREATE TRIGGER update_products_updated_at 
    BEFORE UPDATE ON products 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE products IS 'Stores product information from Bonpreu API';
COMMENT ON TABLE product_nutritional_data IS 'Stores nutritional information for products';
COMMENT ON COLUMN products.product_categories IS 'Array of category strings for the product';
COMMENT ON COLUMN products.created_at IS 'Timestamp when the record was created';
COMMENT ON COLUMN products.updated_at IS 'Timestamp when the record was last updated'; 