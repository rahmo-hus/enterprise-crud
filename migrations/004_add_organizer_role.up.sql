-- Add ORGANIZER role
-- This migration adds the ORGANIZER role for event management
INSERT INTO roles (name, description) VALUES 
('ORGANIZER', 'Event organizer with event management permissions')
ON CONFLICT (name) DO NOTHING;