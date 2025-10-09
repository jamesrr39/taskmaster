package db

import (
	"database/sql"
	"embed"
	"path/filepath"

	"github.com/jamesrr39/go-errorsx"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

const migrationsDir = "migrations"

func RunMigrations(dataDir string) errorsx.Error {

	dbFilePath := filepath.Join(dataDir, "taskmaster-db.sqlite3")

	db, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		return errorsx.Wrap(err)
	}

	goose.SetBaseFS(embedMigrations)

	err = goose.SetDialect(string(database.DialectSQLite3))
	if err != nil {
		return errorsx.Wrap(err)
	}

	err = goose.Up(db, migrationsDir)
	if err != nil {
		return errorsx.Wrap(err)
	}

	return nil
}
