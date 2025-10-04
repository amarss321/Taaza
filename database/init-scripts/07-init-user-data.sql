\c taaza_users;

-- User subscriptions table
CREATE TABLE IF NOT EXISTS user_subscriptions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    subscription_type VARCHAR(50) NOT NULL DEFAULT 'milk',
    morning_enabled BOOLEAN DEFAULT FALSE,
    morning_milk_type VARCHAR(20),
    morning_quantity DECIMAL(4,2),
    morning_frequency VARCHAR(20),
    morning_time_slot VARCHAR(20),
    morning_days JSONB,
    evening_enabled BOOLEAN DEFAULT FALSE,
    evening_milk_type VARCHAR(20),
    evening_quantity DECIMAL(4,2),
    evening_frequency VARCHAR(20),
    evening_time_slot VARCHAR(20),
    evening_days JSONB,
    address_data JSONB,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User preferences table for storing various user settings
CREATE TABLE IF NOT EXISTS user_preferences (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    preference_key VARCHAR(100) NOT NULL,
    preference_value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, preference_key)
);

-- Stock data table (for admin)
CREATE TABLE IF NOT EXISTS stock_data (
    id SERIAL PRIMARY KEY,
    milk_type VARCHAR(20) NOT NULL,
    quantity DECIMAL(10,2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(milk_type)
);

-- Delivery schedule table
CREATE TABLE IF NOT EXISTS delivery_schedule (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    delivery_date DATE NOT NULL,
    time_slot VARCHAR(20),
    milk_type VARCHAR(20),
    quantity DECIMAL(4,2),
    status VARCHAR(20) DEFAULT 'scheduled',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, delivery_date, time_slot)
);

-- Insert default stock data
INSERT INTO stock_data (milk_type, quantity) VALUES 
('buffalo', 100.0),
('cow', 100.0)
ON CONFLICT (milk_type) DO NOTHING;

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_user_subscriptions_user_id ON user_subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_preferences_user_id ON user_preferences(user_id);
CREATE INDEX IF NOT EXISTS idx_delivery_schedule_user_date ON delivery_schedule(user_id, delivery_date);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(token);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires ON user_sessions(expires_at);