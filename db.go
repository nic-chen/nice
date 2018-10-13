package nice

import (
	"database/sql"
)

type Db interface {
	Open() error
	Close() error
	Query(sqlStr string, args ...interface{}) ([]map[string]interface{}, error)
	Exec(sqlStr string, args ...interface{}) (sql.Result, error)
}