package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type migrationFile struct {
	version int64
	path    string
}

func RunMigrations(databaseURI string, migrationsPath string) error {
	path, err := resolveMigrationsPath(migrationsPath)
	if err != nil {
		return err
	}

	db, err := sql.Open("pgx", databaseURI)
	if err != nil {
		return err
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT NOT NULL PRIMARY KEY,
			dirty BOOLEAN NOT NULL DEFAULT FALSE
		)
	`); err != nil {
		return err
	}

	var currentVersion int64
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations WHERE dirty = FALSE`).Scan(&currentVersion); err != nil {
		return err
	}

	files, err := migrationFiles(path)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.version <= currentVersion {
			continue
		}

		if err := applyMigration(ctx, db, file); err != nil {
			return err
		}
	}

	return nil
}

func applyMigration(ctx context.Context, db *sql.DB, file migrationFile) error {
	body, err := os.ReadFile(file.path)
	if err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, dirty) VALUES ($1, TRUE) ON CONFLICT (version) DO UPDATE SET dirty = TRUE`, file.version); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, string(body)); err != nil {
		return fmt.Errorf("apply migration %s: %w", filepath.Base(file.path), err)
	}

	if _, err := tx.ExecContext(ctx, `UPDATE schema_migrations SET dirty = FALSE WHERE version = $1`, file.version); err != nil {
		return err
	}

	return tx.Commit()
}

func migrationFiles(dir string) ([]migrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]migrationFile, 0)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}

		versionText := strings.SplitN(entry.Name(), "_", 2)[0]
		version, err := strconv.ParseInt(versionText, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse migration version %q: %w", entry.Name(), err)
		}

		files = append(files, migrationFile{
			version: version,
			path:    filepath.Join(dir, entry.Name()),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].version < files[j].version
	})

	return files, nil
}

func resolveMigrationsPath(path string) (string, error) {
	candidates := []string{path}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, path),
			filepath.Join(exeDir, "..", path),
			filepath.Join(exeDir, "..", "..", path),
		)
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			abs, err := filepath.Abs(candidate)
			if err != nil {
				return candidate, nil
			}
			return abs, nil
		}
	}

	return "", fmt.Errorf("migrations directory %q not found", path)
}
