-- Drop indexes first
DROP INDEX IF EXISTS idx_roles_name;
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_roles_user_id;

-- Drop junction table (has foreign keys to both tables)
DROP TABLE IF EXISTS user_roles;

-- Drop roles table
DROP TABLE IF EXISTS roles;