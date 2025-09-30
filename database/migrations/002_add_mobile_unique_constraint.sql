-- Add unique constraint to mobile column
ALTER TABLE users ADD CONSTRAINT unique_mobile UNIQUE (mobile);

-- Create index for better performance
CREATE INDEX idx_users_mobile ON users(mobile) WHERE mobile IS NOT NULL;