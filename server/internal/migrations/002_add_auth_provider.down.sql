-- 002_add_auth_provider.down.sql

-- Supprimer l'index sur auth_provider
DROP INDEX IF EXISTS idx_users_auth_provider;

-- Supprimer la colonne auth_provider
ALTER TABLE users DROP COLUMN IF EXISTS auth_provider;

-- Remettre password_hash obligatoire
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
