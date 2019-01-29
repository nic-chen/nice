package nice

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
	"strconv"
)

// mysql is DB pool struct
type Mysql struct {
	DriverName     string
	DataSourceName string
	MaxOpenConns   int
	MaxIdleConns   int
	DB             *sql.DB
	Loger          Logger
}

// Init DB pool
func NewMysql(host, database, user, password, charset string, maxOpenConns, maxIdleConns int) *Mysql {
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&autocommit=true", user, password, host, database, charset)
	p := &Mysql{
		DriverName:     "mysql",
		DataSourceName: dataSourceName,
		MaxOpenConns:   maxOpenConns,
		MaxIdleConns:   maxIdleConns,
		Loger:          log.New(os.Stderr, "[Nice] ", log.LstdFlags),
	}
	if err := p.Open(); err != nil {
		p.Loger.Panicln("Init mysql pool failed.", err.Error())
	}
	return p
}

func (p *Mysql) Open() error {
	var err error
	p.DB, err = sql.Open(p.DriverName, p.DataSourceName)
	if err != nil {
		return err
	}
	if err = p.DB.Ping(); err != nil {
		return err
	}
	p.DB.SetMaxOpenConns(p.MaxOpenConns)
	p.DB.SetMaxIdleConns(p.MaxIdleConns)

	return err
}

// Close pool
func (p *Mysql) Close() error {
	return p.DB.Close()
}

// Get via pool
func (p *Mysql) Get(queryStr string, args ...interface{}) (map[string]interface{}, error) {
	results, err := p.Query(queryStr, args...)
	if err != nil {
		return map[string]interface{}{}, err
	}
	if len(results) <= 0 {
		return map[string]interface{}{}, sql.ErrNoRows
	}
	if len(results) > 1 {
		return map[string]interface{}{}, errors.New("sql: more than one rows")
	}
	return results[0], nil
}

// Query via pool
func (p *Mysql) Query(sqlStr string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := p.DB.Query(sqlStr, args...)
	if err != nil {
		p.Loger.Printf("query err: %v", err)
		return []map[string]interface{}{}, err
	}
	defer rows.Close()

	columns, err := rows.ColumnTypes()
	scanArgs := make([]interface{}, len(columns))
	values := make([]sql.RawBytes, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	rowsMap := make([]map[string]interface{}, 0, 10)
	for rows.Next() {
		rows.Scan(scanArgs...)
		rowMap := make(map[string]interface{})
		for i, value := range values {
			rowMap[columns[i].Name()] = bytes2RealType(value, columns[i])
		}
		rowsMap = append(rowsMap, rowMap)
	}
	if err = rows.Err(); err != nil {
		return []map[string]interface{}{}, err
	}
	return rowsMap, nil
}

// QueryRow via pool
func (p *Mysql) QueryRow(sqlStr string, args ...interface{}) *sql.Row {
	return p.DB.QueryRow(sqlStr, args...)
}

func (p *Mysql) Exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	res, err := p.DB.Exec(sqlStr, args...)
	if err != nil {
		p.Loger.Printf("exec err: %v", err)
	}
	return res, err
}

// Update via pool
func (p *Mysql) Update(updateStr string, args ...interface{}) (int64, error) {
	result, err := p.Exec(updateStr, args...)
	if err != nil {
		return 0, err
	}
	affect, err := result.RowsAffected()
	return affect, err
}

// Insert via pool
func (p *Mysql) Insert(insertStr string, args ...interface{}) (int64, error) {
	result, err := p.Exec(insertStr, args...)
	if err != nil {
		return 0, err
	}
	lastId, err := result.LastInsertId()
	return lastId, err

}

// Delete via pool
func (p *Mysql) Delete(deleteStr string, args ...interface{}) (int64, error) {
	result, err := p.Exec(deleteStr, args...)
	if err != nil {
		return 0, err
	}
	affect, err := result.RowsAffected()
	return affect, err
}

// SQLConnTransaction is for transaction connection
type SQLConnTransaction struct {
	SQLTX *sql.Tx
	Loger Logger
}

// Begin transaction
func (p *Mysql) Begin() (*SQLConnTransaction, error) {
	var oneSQLConnTransaction = &SQLConnTransaction{}
	var err error
	if pingErr := p.DB.Ping(); pingErr == nil {
		oneSQLConnTransaction.SQLTX, err = p.DB.Begin()
		oneSQLConnTransaction.Loger = p.Loger
	}
	return oneSQLConnTransaction, err
}

// Rollback transaction
func (t *SQLConnTransaction) Rollback() error {
	return t.SQLTX.Rollback()
}

// Commit transaction
func (t *SQLConnTransaction) Commit() error {
	return t.SQLTX.Commit()
}

// Get via transaction
func (t *SQLConnTransaction) Get(queryStr string, args ...interface{}) (map[string]interface{}, error) {
	results, err := t.Query(queryStr, args...)
	if err != nil {
		return map[string]interface{}{}, err
	}
	if len(results) <= 0 {
		return map[string]interface{}{}, sql.ErrNoRows
	}
	if len(results) > 1 {
		return map[string]interface{}{}, errors.New("sql: more than one rows")
	}
	return results[0], nil
}

// Query via transaction
func (t *SQLConnTransaction) Query(queryStr string, args ...interface{}) ([]map[string]interface{}, error) {
	rows, err := t.SQLTX.Query(queryStr, args...)
	if err != nil {
		t.Loger.Printf("t query err: %v", err)
		return []map[string]interface{}{}, err
	}
	defer rows.Close()
	columns, err := rows.ColumnTypes()
	scanArgs := make([]interface{}, len(columns))
	values := make([]sql.RawBytes, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}
	rowsMap := make([]map[string]interface{}, 0, 10)
	for rows.Next() {
		rows.Scan(scanArgs...)
		rowMap := make(map[string]interface{})
		for i, value := range values {
			rowMap[columns[i].Name()] = bytes2RealType(value, columns[i])
		}
		rowsMap = append(rowsMap, rowMap)
	}
	if err = rows.Err(); err != nil {
		return []map[string]interface{}{}, err
	}
	return rowsMap, nil
}

func (t *SQLConnTransaction) Exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	res, err := t.SQLTX.Exec(sqlStr, args...)
	if err != nil {
		t.Loger.Printf("t exec err: %v", err)
	}
	return res, err
}

// Update via transaction
func (t *SQLConnTransaction) Update(updateStr string, args ...interface{}) (int64, error) {
	result, err := t.Exec(updateStr, args...)
	if err != nil {
		return 0, err
	}
	affect, err := result.RowsAffected()
	return affect, err
}

// Insert via transaction
func (t *SQLConnTransaction) Insert(insertStr string, args ...interface{}) (int64, error) {
	result, err := t.Exec(insertStr, args...)
	if err != nil {
		return 0, err
	}
	lastId, err := result.LastInsertId()
	return lastId, err

}

// Delete via transaction
func (t *SQLConnTransaction) Delete(deleteStr string, args ...interface{}) (int64, error) {
	result, err := t.Exec(deleteStr, args...)
	if err != nil {
		return 0, err
	}
	affect, err := result.RowsAffected()
	return affect, err
}

// bytes2RealType is to convert db type to code type
func bytes2RealType(src []byte, column *sql.ColumnType) interface{} {
	srcStr := string(src)
	var result interface{}
	switch column.DatabaseTypeName() {
	case "BIT", "TINYINT", "SMALLINT", "INT":
		result, _ = strconv.ParseInt(srcStr, 10, 64)
	case "BIGINT":
		result, _ = strconv.ParseUint(srcStr, 10, 64)
	case "CHAR", "VARCHAR",
		"TINY TEXT", "TEXT", "MEDIUM TEXT", "LONG TEXT",
		"TINY BLOB", "MEDIUM BLOB", "BLOB", "LONG BLOB",
		"JSON", "ENUM", "SET",
		"YEAR", "DATE", "TIME", "TIMESTAMP", "DATETIME":
		result = srcStr
	case "FLOAT", "DOUBLE", "DECIMAL":
		result, _ = strconv.ParseFloat(srcStr, 64)
	default:
		result = nil
	}
	return result
}
