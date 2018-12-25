# Nice 路由

nice 基于 http resetfull 模式设计了路由管理器，提供了常规路由，参数路由，文件路由，静态文件路由，还有组路由。

## 常规路由

```
func (b *Nice) Delete(pattern string, h ...HandlerFunc) RouteNode
func (b *Nice) Get(pattern string, h ...HandlerFunc) RouteNode
func (b *Nice) Head(pattern string, h ...HandlerFunc) RouteNode
func (b *Nice) Options(pattern string, h ...HandlerFunc) RouteNode
func (b *Nice) Patch(pattern string, h ...HandlerFunc) RouteNode
func (b *Nice) Post(pattern string, h ...HandlerFunc) RouteNode
func (b *Nice) Put(pattern string, h ...HandlerFunc) RouteNode
```

接受两个参数，一个是URI路径，另一个是HandlerFunc类型，设定匹配到该路径时执行的方法；允许多个，按照设定顺序进行链式处理。

返回一个RouteNode，该Node只有一个方法，`Name(name string)` 用于命名该条路由规则，以备后用。

除了以上几个标准方法，还支持多个method设定的路由姿势：

```
func (b *Nice) Route(pattern, methods string, h ...HandlerFunc) RouteNode
func (b *Nice) Any(pattern string, h ...HandlerFunc) RouteNode
```

## 路由语法

### 静态路由

静态路由语法就是没有任何参数变量，pattern是一个固定的字符串。

使用示例：

```
package main

import (
	"github.com/nic-chen/nice"
)

func main() {
	app := nice.Instance("")
	app.Get("/foo", func(c *nice.Context) {
		c.String(200, c.URL(true))
	})
	app.Get("/bar", func(c *nice.Context) {
		c.String(200, c.URL(true))
	})
	app.Run("8080")
}
```

测试：

```
curl http://127.0.0.1:8080/foo
curl http://127.0.0.1:8080/bar
```

### 参数路由

静态路由是最基础的，但显然满足不了需求的，我们在程序中通常相同的资源访问规则相同，不同的只是资源的编号，这时就该参数路由出场了。

> 参数路由以 `/` 为拆分单位，故每两个斜线区间中只能有一个参数存在，更复杂的规则需要 正则路由。

参数路由以冒号 `:` 后面跟一个字符串作为参数名称，可以通过 `Context`的 `Param` 系列方法获取路由参数的值。

使用示例：

```
package main

import (
	"fmt"
	"github.com/nic-chen/nice"
)

func main() {
	app := nice.Instance("")
	app.Get("/user/:id", func(c *nice.Context) {
		c.String(200, "My user id is: " + c.Param("id"))
	})
	app.Get("/user/:id/project/:pid", func(c *nice.Context) {
		id := c.ParamInt("id")
		pid := c.ParamInt("pid")
		c.String(200, fmt.Sprintf("user id: %d, project id: %d", id, pid))
	})
	app.Run("8080")
}
```

测试：

```
curl http://127.0.0.1:8080/user/101
curl http://127.0.0.1:8080/user/101/project/201
```

## 路由选项

```
func (b *Nice) SetAutoHead(v bool)
```

搜索引擎很喜欢用HEAD方法来检查一个网页是否能正常访问。但我们一般又不会单独写一个HEAD的处理方法，一般行为是GET返回的数据不要内容。

使用 `app.SetAutoHead(true)` 将在设置 `GET` 方法时，自动添加 `HEAD` 路由，绑定和GET一样的处理。

```
func (b *Nice) SetAutoTrailingSlash(v bool)
```

在URL访问中，一个目录要带不带最后的斜线也有很多争议，google站长工具明确表示，带不带斜线将表示不同的URL资源，但是浏览习惯问题，很多时候带不带都能访问到相同的资源目录。

使用 `app.SetAutoTrailingSlash(true)` 将处理最后的斜线，将带和不带都统一行为，自动补全最后一个斜线。

## 组路由

```
func (b *Nice) Group(pattern string, f func(), h ...HandlerFunc)
```

组路由，顾名思义，用来处理一组路由的需求，可以设定统一的前缀，统一的前置方法。

使用示例：

```
package main

import (
	"fmt"
	"github.com/nic-chen/nice"
)

func main() {
	app := nice.Instance("")

	app.Group("/group", func() {
		app.Get("/", func(c *nice.Context) {
			c.String(200, "我是组的首页")
		})
		app.Group("/user", func() {
			app.Get("/", func(c *nice.Context) {
				c.String(200, "我是组的用户")
			})
			app.Get("/:id", func(c *nice.Context) {
				c.String(200, "in group, user id: "+c.Param("id"))
			})
		})
		app.Get("/:gid", func(c *nice.Context) {
			c.String(200, "in group, group id: "+c.Param("gid"))
		})
	}, func(c *nice.Context) {
		// 我是组内的前置检测，过不了我这关休想访问组内的资源
	})

	app.Run("8080")
}
```

测试：

```
curl http://127.0.0.1:8080/group/
curl http://127.0.0.1:8080/group/user/
curl http://127.0.0.1:8080/group/user/101
curl http://127.0.0.1:8080/group/111
```

### 链式处理

一个URL请求可以先处理A，根据A的结果再执行B。

**举个例子：**

一个URL要先判断你登录过才可以访问，就可以设定两个Handler，第一个 判断是否登录，如果没登录就调到登录界面，否则继续执行第二个真正的内容。

使用示例：

```
package main

import (
	"github.com/nic-chen/nice"
)

func main() {
	app := nice.Default()
	app.Get("/", func(c *nice.Context) {
		c.String(200, "Hello, World")
	})
	app.Post("/", func(c *nice.Context) {
		c.String(200, c.Req.Method)
	})
	app.Get("/admin", func(c *nice.Context) {
		if c.GetCookie("login_id") != "admin" {
			c.Redirect(302, "/login")
			c.Break()
		}
	}, func(c *nice.Context) {
		c.String(200, "恭喜你，看到后台了")
	})
	app.Get("/login", func(c *nice.Context) {
		c.Resp.Header().Set("Content-Type", "text/html; charset=utf-8")
		c.SetCookie("login_id", "admin", 3600, "/")
		c.Resp.Write([]byte("登录成功，<a href=\"/admin\">点击进入后台</a>"))
	})
	app.Run(":8080")
}
```

## 命名路由

```
func (n *Node) Name(name string)
func (b *Nice) URLFor(name string, args ...interface{}) string
```

前面可以看到添加路由后，返回了一个 `RouteNode` 说可以做命名路由，有什么用呢？

就是给一个URL起个名字，然后在程序中可以通过 `URLFor`方法来生成这个符合这个路由的URL路径。

举个例子：

```
app := nice.Instance("")
app.Get("/user/:id/project", func(c *nice.Context) {
	c.String(200, c.Nice().URLFor("user_project", c.Param("id")))
}).Name("user_project")
```

执行上面的方法，会输出你当前访问的URL，就是这个姿势。

## 文件路由

```
func (b *Nice) Static(prefix string, dir string, index bool, h HandlerFunc)
func (b *Nice) StaticFile(pattern string, path string) RouteNode
```

在一个完整的应用中，我们除了业务逻辑，还有访问图片/CSS/JS等需求，通过文件路由，可以直接访问文件或文件夹。

`app.StaticFile` 可以让你直接访问一个具体的文件，比如： robots.txt

`app.Static` 可以访问一个目录下所有的资源，甚至列出目录结构，类似文件服务器。

举个例子：

```
app := nice.Instance("")
app.Static("/assets", "/data/www/public/asssets", true, func(c *nice.Context) {
	// 你可以对输出的结果干点啥的
})
app.Static("/robots.txt", "/data/www/public/robots.txt")
```

就是这样，第一条路由就可以列出目录和访问下面的资源了。第二条路由可以直接返回一个静态文件。

## 自定义错误

### 500错误

```
func (b *Nice) SetError(h ErrorHandleFunc)
```

要是运行过程中程序出错了，怎么办，会不会泄露你的隐私，能不能提供点错误日志？

nice 默认在 `debug` 模式下向浏览器发送具体的错误信息，线上运行只显示 `Internal Server Error` 并返回 `500` 错误头。

可以通过 `app.SetError` 来设置错误处理方法，该方法接受一个 ErrorHandleFunc类型。

### 404错误

```
func (b *Nice) SetNotFound(h HandlerFunc)
```

nice默认返回 `Not Found` 和 `404` 错误头，你也可以通过 `app.SetNotFound`来自定义错误处理，该方法接受一个 HandlerFunc类型。


举个例子：

```
app := nice.Instance("")
app.SetError(func(err error, c *nice.Context) {
	c.Nice().Logger().Println("记录日志", err)
	c.String(500, "出错了")
})
app.SetNotFound(func(c *nice.Context) {
	c.String(404, "页面放假了，请稍后再来。")
})
app.Run(":8080")
```


