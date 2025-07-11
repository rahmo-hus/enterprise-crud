-- Remove admin user migration rollback
-- This migration removes the default admin user from the system

-- Remove the admin user's role assignments first
DELETE FROM user_roles 
WHERE user_id = (
    SELECT id FROM users WHERE email = 'admin@enterprise-crud.com'
);

-- Remove the admin user
DELETE FROM users 
WHERE email = 'admin@enterprise-crud.com';