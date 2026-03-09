package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jiazaiwanbi/second-hand-platform/internal/platform/config"
)

type migrationFile struct {
	Version   int
	Direction string
	Path      string
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./cmd/migrate [up|down|version|force <version>]")
	}

	cfg, err := config.LoadForMigrate()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", cfg.Postgres.URL())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	if err := ensureSchemaMigrationsTable(db); err != nil {
		log.Fatal(err)
	}

	switch os.Args[1] {
	case "up":
		if err := migrateUp(db); err != nil {
			log.Fatal(err)
		}
	case "down":
		if err := migrateDown(db); err != nil {
			log.Fatal(err)
		}
	case "version":
		version, dirty, err := currentVersion(db)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("version=%d dirty=%t\n", version, dirty)
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("usage: go run ./cmd/migrate force <version>")
		}
		version, err := strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("parse version: %v", err)
		}
		if err := forceVersion(db, version, false); err != nil {
			log.Fatal(err)
		}
	default:
		log.Fatalf("unsupported command: %s", os.Args[1])
	}
}

func migrateUp(db *sql.DB) error {
	version, dirty, err := currentVersion(db)
	if err != nil {
		return err
	}
	if dirty {
		return fmt.Errorf("dirty database version %d", version)
	}

	files, err := migrationFiles("up")
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Version <= version {
			continue
		}
		if err := applyMigration(db, file); err != nil {
			return err
		}
		return nil
	}

	return nil
}

func migrateDown(db *sql.DB) error {
	version, dirty, err := currentVersion(db)
	if err != nil {
		return err
	}
	if dirty {
		return fmt.Errorf("dirty database version %d", version)
	}
	if version == 0 {
		return nil
	}

	files, err := migrationFiles("down")
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Version == version {
			return applyMigration(db, file)
		}
	}

	return fmt.Errorf("down migration for version %d not found", version)
}

func applyMigration(db *sql.DB, file migrationFile) error {
	content, err := os.ReadFile(file.Path)
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(string(content)); err != nil {
		_ = setVersionTx(tx, file.Version, true)
		return fmt.Errorf("apply migration %s: %w", filepath.Base(file.Path), err)
	}

	nextVersion := file.Version
	if file.Direction == "down" {
		nextVersion = file.Version - 1
	}

	if err := setVersionTx(tx, nextVersion, false); err != nil {
		return err
	}

	return tx.Commit()
}

func migrationFiles(direction string) ([]migrationFile, error) {
	dir, err := migrationsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]migrationFile, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		parts := strings.Split(name, "_")
		if len(parts) < 2 {
			continue
		}
		version, err := strconv.Atoi(parts[0])
		if err != nil {
			continue
		}
		if !strings.HasSuffix(name, "."+direction+".sql") {
			continue
		}
		files = append(files, migrationFile{
			Version:   version,
			Direction: direction,
			Path:      filepath.Join(dir, name),
		})
	}

	sort.Slice(files, func(i, j int) bool {
		if direction == "down" {
			return files[i].Version > files[j].Version
		}
		return files[i].Version < files[j].Version
	})

	return files, nil
}

func migrationsDir() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	candidates := []string{
		filepath.Join(cwd, "db", "migrations"),
		filepath.Join(cwd, "..", "..", "db", "migrations"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("db/migrations not found from %s", cwd)
}

func ensureSchemaMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT NOT NULL PRIMARY KEY,
			dirty BOOLEAN NOT NULL
		)
	`)
	return err
}

func currentVersion(db *sql.DB) (int, bool, error) {
	row := db.QueryRow(`SELECT version, dirty FROM schema_migrations LIMIT 1`)
	var version int
	var dirty bool
	if err := row.Scan(&version, &dirty); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, false, nil
		}
		return 0, false, err
	}
	return version, dirty, nil
}

func forceVersion(db *sql.DB, version int, dirty bool) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := setVersionTx(tx, version, dirty); err != nil {
		return err
	}
	return tx.Commit()
}

func setVersionTx(tx *sql.Tx, version int, dirty bool) error {
	if _, err := tx.Exec(`TRUNCATE schema_migrations`); err != nil {
		return err
	}
	if version == 0 && !dirty {
		return nil
	}
	_, err := tx.Exec(`INSERT INTO schema_migrations (version, dirty) VALUES ($1, $2)`, version, dirty)
	return err
}
