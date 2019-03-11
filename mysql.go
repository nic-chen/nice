package nice

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/mitchellh/mapstructure"
	"log"
	"os"
	"strconv"
	"time"
)

// mysql is DB pool struct
type Mysql struct {
	DriverName string
	Master     *sql.DB
	Slave      *sql.DB
	Loger      Logger
}

type MysqlConf struct {
	//map类型
	Master      MysqlNode   `yaml: "master"`
	Slave       []MysqlNode `yaml: "slaves"`
	Database    string      `yaml: "database"`
	Charset     string      `yaml: "charset"`
	MaxLifetime string      `yaml: "maxlifetime"`
}

type MysqlNode struct {
	Host     string `yaml: "host"`
	User     string `yaml: "user"`
	Password string `yaml: "password"`
	MaxOpen  int    `yaml: "maxopen"` //maxOpenConn
	MaxIdle  int    `yaml: "maxidle"` //maxIdleConn
}

// Init DB pool
func NewMysql(config interface{}) *Mysql {
	var err error
	p := &Mysql{
		DriverName: "mysql",
		Loger:      log.New(os.Stderr, "[Nice] ", log.LstdFlags),
	}

	p.Loger.Printf("mysql config:%v", config)
	conf := MysqlConf{}
	err = mapstructure.Decode(config.(map[string]interface{}), &conf)

	p.Master, err = p.Connect(conf.Master, conf.Database, conf.Charset, conf.MaxLifetime)
	if err != nil {
		p.Loger.Panicln("Init mysql master pool failed.", err.Error())
	}

	// p.Slave, err = p.Connect(conf.Slave, conf.Database, conf.Charset, conf.MaxLifetime)
	// if err != nil {
	// 	return nil
	// }

	return p
}

func (p *Mysql) Connect(node MysqlNode, database, charset, MaxLifetime string) (*sql.DB, error) {
	var err error

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=%s&autocommit=true", node.User, node.Password, node.Host, database, charset)

	p.Loger.Println("dsn: %s", dsn)

	conn, err := sql.Open(p.DriverName, dsn)
	if err != nil {
		return nil, err
	}
	if err = conn.Ping(); err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(node.MaxOpen)
	conn.SetMaxIdleConns(node.MaxIdle)

	mlt, _ := time.ParseDuration(MaxLifetime)

	conn.SetConnMaxLifetime(mlt)

	return conn, err
}

// Close pool
func (p *Mysql) Close() error {
	return p.Master.Close()
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
	rows, err := p.Master.Query(sqlStr, args...)
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
	return p.Master.QueryRow(sqlStr, args...)
}

func (p *Mysql) Exec(sqlStr string, args ...interface{}) (sql.Result, error) {
	res, err := p.Master.Exec(sqlStr, args...)
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
	if pingErr := p.Master.Ping(); pingErr == nil {
		oneSQLConnTransaction.SQLTX, err = p.Master.Begin()
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
