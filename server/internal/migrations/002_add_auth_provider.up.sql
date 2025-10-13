-- Ajouter la colonne auth_provider avec valeur par défaut 'local'
ALTER TABLE users ADD COLUMN auth_provider VARCHAR(50) DEFAULT 'local';

-- Créer un index pour accélérer les recherches par auth_provider
CREATE INDEX idx_users_auth_provider ON users(auth_provider);

-- Optionnel : rendre le password nullable pour les comptes OAuth
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
