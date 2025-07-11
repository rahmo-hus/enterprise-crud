-- Remove ORGANIZER role
-- This migration removes the ORGANIZER role
DELETE FROM user_roles WHERE role_id = (SELECT id FROM roles WHERE name = 'ORGANIZER');
DELETE FROM roles WHERE name = 'ORGANIZER';