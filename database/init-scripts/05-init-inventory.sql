-- Create inventory database and tables
-- Connect to the inventory database
\c taaza_inventory;

-- Inventory products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    type VARCHAR(20) NOT NULL, -- 'buffalo', 'cow'
    price_per_liter DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Inventory stock table
CREATE TABLE IF NOT EXISTS inventory_stock (
    id SERIAL PRIMARY KEY,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    time_slot VARCHAR(10) NOT NULL, -- 'morning', 'evening'
    total_stock DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    booked_stock DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    available_stock DECIMAL(10,2) GENERATED ALWAYS AS (total_stock - booked_stock) STORED,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(product_id, time_slot)
);

-- Stock history for tracking changes
CREATE TABLE IF NOT EXISTS stock_history (
    id SERIAL PRIMARY KEY,
    product_id INTEGER REFERENCES products(id) ON DELETE CASCADE,
    time_slot VARCHAR(10) NOT NULL,
    change_type VARCHAR(20) NOT NULL, -- 'stock_add', 'stock_remove', 'booking_add', 'booking_remove'
    quantity DECIMAL(10,2) NOT NULL,
    previous_value DECIMAL(10,2) NOT NULL,
    new_value DECIMAL(10,2) NOT NULL,
    reason TEXT,
    created_by VARCHAR(50) DEFAULT 'system',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Notification requests table
CREATE TABLE IF NOT EXISTS notification_requests (
    id SERIAL PRIMARY KEY,
    customer_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    milk_type VARCHAR(20) NOT NULL, -- 'buffalo', 'cow'
    time_slot VARCHAR(10) NOT NULL,
    quantity DECIMAL(10,2) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'notified', 'expired'
    notes TEXT,
    notified_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default products
INSERT INTO products (name, type, price_per_liter) VALUES 
('Buffalo Milk', 'buffalo', 75.00),
('Cow Milk', 'cow', 50.00)
ON CONFLICT (name) DO NOTHING;

-- Insert default stock for each product and time slot
INSERT INTO inventory_stock (product_id, time_slot, total_stock, booked_stock) 
SELECT p.id, ts.slot, 0.00, 0.00
FROM products p
CROSS JOIN (VALUES ('morning'), ('evening')) AS ts(slot)
ON CONFLICT (product_id, time_slot) DO NOTHING;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_inventory_stock_product_time ON inventory_stock(product_id, time_slot);
CREATE INDEX IF NOT EXISTS idx_stock_history_product ON stock_history(product_id);
CREATE INDEX IF NOT EXISTS idx_stock_history_created_at ON stock_history(created_at);
CREATE INDEX IF NOT EXISTS idx_notification_requests_status ON notification_requests(status);
CREATE INDEX IF NOT EXISTS idx_notification_requests_created_at ON notification_requests(created_at);

-- Create update trigger for inventory_stock
CREATE OR REPLACE FUNCTION update_inventory_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_inventory_timestamp
    BEFORE UPDATE ON inventory_stock
    FOR EACH ROW
    EXECUTE FUNCTION update_inventory_timestamp();