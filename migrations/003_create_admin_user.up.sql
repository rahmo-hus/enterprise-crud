-- Create admin user migration
-- This migration adds a default admin user to the system

-- First, let's create the admin user with bcrypt hashed password
-- Password: "admin123" (change this after first login!)
-- The hash below is bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
INSERT INTO users (id, email, username, password, created_at, updated_at) VALUES 
(
    uuid_generate_v4(),
    'admin@enterprise-crud.com',
    'admin',
    '$2a$12$ND7Oc0W3sci0H.iyz2NJdOFbCu7Co0LbxgNXyX2b5JOlWNhwsWdUSverif', -- bcrypt hash for "admin123"
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Get the admin user ID and admin role ID for the junction table
-- Then assign ADMIN role to the admin user
INSERT INTO user_roles (user_id, role_id, assigned_at)
SELECT 
    u.id as user_id,
    r.id as role_id,
    NOW() as assigned_at
FROM users u
CROSS JOIN roles r
WHERE u.email = 'admin@enterprise-crud.com' 
  AND r.name = 'ADMIN'
ON CONFLICT (user_id, role_id) DO NOTHING;