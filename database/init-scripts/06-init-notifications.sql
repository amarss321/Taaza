-- Notification requests table
CREATE TABLE IF NOT EXISTS notification_requests (
    id SERIAL PRIMARY KEY,
    customer_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(15) NOT NULL,
    milk_type VARCHAR(20) NOT NULL CHECK (milk_type IN ('buffalo', 'cow')),
    quantity DECIMAL(3,2) NOT NULL,
    time_slot VARCHAR(10) NOT NULL CHECK (time_slot IN ('morning', 'evening')),
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'notified', 'cancelled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    notified_at TIMESTAMP NULL,
    notes TEXT
);

-- Index for faster queries
CREATE INDEX IF NOT EXISTS idx_notification_status ON notification_requests(status);
CREATE INDEX IF NOT EXISTS idx_notification_milk_time ON notification_requests(milk_type, time_slot);