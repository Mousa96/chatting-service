ALTER TABLE messages
ADD COLUMN status VARCHAR(20) DEFAULT 'sent' NOT NULL,
ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP; 