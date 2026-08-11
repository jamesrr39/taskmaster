package db

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

var (
	// Compile time check on interface satisfaction
	_ DBConn = &sqlx.DB{}
)

type DBConn interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...any) (sql.Result, error)
}
