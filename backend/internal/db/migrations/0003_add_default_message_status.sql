-- Add default status for existing messages
UPDATE messages 
SET status = 'sent' 
WHERE status IS NULL;

-- Make status non-nullable with default value
ALTER TABLE messages 
ALTER COLUMN status SET DEFAULT 'sent',
ALTER COLUMN status SET NOT NULL; 