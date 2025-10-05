package database

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (db *DB) RunMigrations(migrationsPath string) error {
	// Créer la table de suivi des migrations si elle n'existe pas
	createMigrationsTable := `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(255) PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        );
    `

	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Lister les fichiers de migration
	files, err := filepath.Glob(filepath.Join(migrationsPath, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}

	sort.Strings(files)

	for _, file := range files {
		version := strings.TrimSuffix(filepath.Base(file), ".up.sql")

		// Vérifier si la migration a déjà été appliquée
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			continue // Migration déjà appliquée
		}

		// Lire et exécuter le fichier de migration
		content, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Exécuter les requêtes une par une
		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		var query strings.Builder

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "--") {
				continue
			}

			query.WriteString(line + " ")

			if strings.HasSuffix(line, ";") {
				if _, err := db.Exec(query.String()); err != nil {
					return fmt.Errorf("failed to execute migration %s: %w", version, err)
				}
				query.Reset()
			}
		}

		// Marquer la migration comme appliquée
		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			return fmt.Errorf("failed to mark migration as applied: %w", err)
		}

		fmt.Printf("Applied migration: %s\n", version)
	}

	return nil
}
