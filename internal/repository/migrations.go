package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func RunMigrations(databaseURI string, migrationsPath string) error {
	resolvedPath, err := resolveMigrationsPath(migrationsPath)
	if err != nil {
		return err
	}

	m, err := migrate.New(
		"file://"+filepath.ToSlash(resolvedPath),
		databaseURI,
	)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}

		return err
	}

	return nil
}

func resolveMigrationsPath(migrationsPath string) (string, error) {
	if migrationsPath == "" {
		return "", fmt.Errorf("migrations path is empty")
	}

	candidates := make([]string, 0, 8)
	if filepath.IsAbs(migrationsPath) {
		candidates = append(candidates, migrationsPath)
	} else {
		if wd, err := os.Getwd(); err == nil {
			candidates = append(candidates,
				filepath.Join(wd, migrationsPath),
				filepath.Join(wd, "..", migrationsPath),
			)
		}

		if exe, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exe)
			candidates = append(candidates,
				filepath.Join(exeDir, migrationsPath),
				filepath.Join(exeDir, "..", migrationsPath),
				filepath.Join(exeDir, "..", "..", migrationsPath),
			)
		}
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

	return "", fmt.Errorf("migrations directory %q not found", migrationsPath)
}
