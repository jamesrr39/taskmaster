package db

type DBConn interface {
	Get(dest interface{}, query string, args ...interface{}) error
}
