# Nice 依赖注入

依赖注入(dependency injection)简称 DI，是 nice 实现的核心，nice 所有组件基于DI组装起来的。

默认的 日志、路由、模板 都是通过 DI 注册进来的，在 [Nice核心#更换内置引擎](https://github.com/nic-chen/nice/tree/master/doc/nice.md#更换内置引擎)一节也介绍过。

Nice的初始化函数是这样写的：

```
// New create a nice application without any config.
func New() *Nice {
	n := new(Nice)
	n.middleware = make([]HandlerFunc, 0)
	n.pool = sync.Pool{
		New: func() interface{} {
			return NewContext(nil, nil, n)
		},
	}
	if Env != PROD {
		n.debug = true
	}
	n.SetDIer(NewDI())
	n.SetDI("router", NewTree(n))
	n.SetDI("logger", log.New(os.Stderr, "[Nice] ", log.LstdFlags))
	n.SetDI("render", newRender())
	// n.SetDI("db", NewMysql())
	// n.SetDI("cache", NewRedis())
	n.SetNotFound(n.DefaultNotFoundHandler)
	return n
}
```

代码出处，[nice.go](https://github.com/nic-chen/nice/tree/master/nice.go)

## 注册

`func (b *Nice) SetDI(name string, h interface{})`

DI 的注册和使用都依赖于注册的名称，比如要更换内置组件必须注册为指定的名称。

### name string

依赖的名称

### h interface{}

依赖的实例，可以是任意类型

使用示例：

```
package main

import (
	"log"
	"os"

	"github.com/nic-chen/nice"
)

func main() {
	app := nice.Instance("")
	app.SetDI("logger", log.New(os.Stderr, "[NiceDI] ", log.LstdFlags))

	app.Get("/", func(c *nice.Context) {
		c.String(200, "Hello, World")
	})

	app.Run(":8080")
}
```

## 使用

`func (b *Nice) GetDI(name string) interface{}`

`func (c *Context) DI(name string) interface{}`


可以通过 `nice.GetDI` 或者 `c.DI` 来获取已经注册的依赖，如果未注册，返回 `nil`

由于注册的依赖可能是任意类型，故返回类型为 `interface{}`，所以获取后，需要做一次类型断言再使用。

使用示例：

```
package main

import (
	"log"
	"os"

	"github.com/nic-chen/nice"
)

func main() {
	app := nice.Instance("")
	app.SetDI("logger", log.New(os.Stderr, "[NiceDI] ", log.LstdFlags))

	app.Get("/", func(c *nice.Context) {
		// use di
		logger := c.DI("logger").(*log.Logger)
		logger.Println("i am use logger di")

		c.String(200, "Hello, World")
	})

	app.Run(":8080")
}
```

### 日志

nice 将日志抽象为 `nice.Logger` 接口，只要实现了该接口，就可以注册为全局日志器。

nice 内置的日志器使用的是标准包的 `log` 实例。

更换全局日志器：

```
nice.SetDI("logger", newLogger)
```

> logger 是内置名称，该命名被用于全局日志器。
> 
> 如果不是要更换全局日志，而是注册一个新的日志器用于其他用途，只需更改注册名称即可，而且也不需要实现 `nice.Logger` 接口。

### 路由

只要实现接口 `nice.Router` 接口即可。

```
nice.SetDI("router", newRouter)
```

> router 是内置名称，该命名被用于全局路由器。


### 数据库

只要实现接口 `nice.Db` 接口即可。

```
nice.SetDI("db", nice.NewMysql(config.MYSQLHOST, config.MYSQLDB, config.MYSQLUSER, config.MYSQLPWD, config.DBCHARSET, config.DBCONNOPEN, config.DBCONNIDLE))
```

> db 是内置名称，该命名被用于数据库操作。

> db 需要配置地址账户等，所以一般放到应用中进行注册，[example](https://github.com/nic-chen/nice-example/blob/master/cmd/srv/api.go)


### 缓存

只要实现接口 `nice.Cache` 接口即可。

```
	n.SetDI("cache", nice.NewRedis(config.REDISHOST, config.REDISPWD, config.REDISDB, config.DBCONNOPEN, config.DBCONNIDLE))
```

> db 是内置名称，该命名被用于数据库操作。
> cache 需要配置地址账户等，所以一般放到应用中进行注册，[example](https://github.com/nic-chen/nice-example/blob/master/cmd/srv/api.go)
