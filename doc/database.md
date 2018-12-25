# 数据库

Nice 内置实现了MySQL操作，如果有需要可以按接口实现其他数据库的操作

初始化：

```
	
	app := nice.Instance(config.APP_NAME);
	app.SetDI("db", nice.NewMysql(MYSQL_HOST, MYSQL_DB, MYSQL_USER, MYSQL_PWD, CHARSET, CONN_MAX_OPEN, CONN_MAX_IDLE))
	//CONN_MAX_OPEN 为最大打开连接数
	//CONN_MAX_IDLE 为最大闲置连接数
	
```
查询sql
```

	db := app.Db()
	val := "a"
	sql := "SELECT * FROM tbl WHERE col =?"
	res, err := db.Query(sql, val) // 执行语句并返回

```

执行sql
```

	db := app.Db()
	val := "a"
	sql := "DELETE FROM tbl WHERE col =?"
	res, err := db.Exec(sql, val) // 执行语句并返回

```

事务
```

	db := app.Db()

	t := db.Begin()

	val := "a"
	sql := "DELETE FROM tbl WHERE col =?"
	res, err := t.Exec(sql, val) // 执行语句并返回

	t.Commit();//t.Rollback

```