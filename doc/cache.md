# 缓存

Nice 内置实现了Redis操作，如果有需要可以按接口实现其他缓存数据库的操作

初始化：

```
	
	app := nice.Instance(config.APP_NAME);
	n.SetDI("cache", nice.NewRedis(REDIS_HOST, REDIS_PWD, REDIS_DB, CONN_MAX_OPEN, CONN_MAX_IDLE));
	
	//CONN_MAX_OPEN 为最大打开连接数
	//CONN_MAX_IDLE 为最大闲置连接数
	
```
执行命令
```

	cache := app.Cache()
	key   := "member_1"
	_, err := cache.Do("DEL", key);

```


> 其他命令请参考 https://godoc.org/github.com/gomodule/redigo/redis