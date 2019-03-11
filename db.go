package nice

import (
	"database/sql"
)

type Db interface {
	Close() error
	Query(sqlStr string, args ...interface{}) ([]map[string]interface{}, error)
	QueryRow(sqlStr string, args ...interface{}) *sql.Row
	Exec(sqlStr string, args ...interface{}) (sql.Result, error)
}
