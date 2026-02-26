package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"tracker/internal/infra/config"
)

const (
	defaultMigrationsDir = "migrations"
)

func main() {
	databaseURL, err := resolveDatabaseURL()
	if err != nil {
		log.Fatalf("resolve database url: %v", err)
	}

	migrationsDir, err := resolveMigrationsDir()
	if err != nil {
		log.Fatalf("resolve migrations dir: %v", err)
	}

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("set goose dialect: %v", err)
	}

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("ping db: %v", err)
	}

	command, args := parseCommand()
	if err := run(command, db, migrationsDir, args); err != nil {
		log.Fatalf("goose %s failed: %v", command, err)
	}
}

func parseCommand() (string, []string) {
	if len(os.Args) < 2 {
		return "up", nil
	}
	return os.Args[1], os.Args[2:]
}

func resolveDatabaseURL() (string, error) {
	if databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); databaseURL != "" {
		return databaseURL, nil
	}

	cfg, err := config.NewForMigrate()
	if err != nil {
		return "", err
	}
	return cfg.Database.DSN(), nil
}

func resolveMigrationsDir() (string, error) {
	if dir := strings.TrimSpace(os.Getenv("MIGRATIONS_DIR")); dir != "" {
		return dir, nil
	}

	if st, err := os.Stat(defaultMigrationsDir); err == nil && st.IsDir() {
		return defaultMigrationsDir, nil
	}

	return "", fmt.Errorf(
		"migrations directory %q not found; run command from backend/ or set MIGRATIONS_DIR",
		defaultMigrationsDir,
	)
}

func run(command string, db *sql.DB, dir string, args []string) error {
	switch command {
	case "up":
		return goose.Up(db, dir)
	case "down":
		return goose.Down(db, dir)
	case "up-by-one":
		return goose.UpByOne(db, dir)
	case "redo":
		return goose.Redo(db, dir)
	case "reset":
		return goose.Reset(db, dir)
	case "status":
		return goose.Status(db, dir)
	case "version":
		return goose.Version(db, dir)
	case "fix":
		return goose.Fix(dir)
	case "create":
		if len(args) < 2 {
			return errors.New("usage: go run ./cmd/migrate create <name> <sql|go>")
		}
		return goose.Create(db, dir, args[0], args[1])
	case "up-to":
		version, err := parseVersionArg(args)
		if err != nil {
			return err
		}
		return goose.UpTo(db, dir, version)
	case "down-to":
		version, err := parseVersionArg(args)
		if err != nil {
			return err
		}
		return goose.DownTo(db, dir, version)
	default:
		return fmt.Errorf("unsupported command %q", command)
	}
}

func parseVersionArg(args []string) (int64, error) {
	if len(args) < 1 {
		return 0, errors.New("version argument is required")
	}
	version, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid version %q: %w", args[0], err)
	}
	return version, nil
}
