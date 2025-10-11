package db

import (
	"database/sql"
	"embed"
	"path/filepath"

	"github.com/jamesrr39/go-errorsx"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

const migrationsDir = "migrations"

func RunMigrations(db *sql.DB) errorsx.Error {
	var err error

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

func OpenDB(filePath string) (*sqlx.DB, errorsx.Error) {
	dbFilePath := filepath.Join(filePath, "data", "taskmaster-db.sqlite3")

	db, err := sqlx.Open("sqlite", dbFilePath)
	if err != nil {
		return nil, errorsx.Wrap(err)
	}

	return db, nil
}
