# Nice

一个简单高效的Go Web及微服务开发框架。主要Web框架有路由、中间件，依赖注入和HTTP上下文构成，微服务主要有服务发现、分布式追踪等。


## 特性

* 路由
* 中间件
* 依赖注入
* 日志处理
* 数据库操作
* 缓存操作
* 统一的HTTP错误处理
* 微服务
    * 服务发现
    * 分布式追踪
* ……

## 快速上手

### 安装

```
go get -u https://github.com/nic-chen/nice
```

### 代码

```
// nice.go
package main

import (
    "github.com/nic-chen/nice"
)

func main() {
    app := nice.Instance("")
    app.Get("/", func(c *nice.Context) {
        c.String(200, "Hello, World")
    })
    app.Run(":8080")
}
```

### 运行

```
go run nice.go
```

### 浏览

```
http://127.0.0.1:8080/
```

### 使用中间件

```
package main

import (
    "github.com/nic-chen/nice/middleware"
    "github.com/nic-chen/nice/middleware"
    "github.com/nic-chen/nice"
)

func main() {
    app := nice.Instance("")
    app.Use(middleware.Recovery())
    app.Use(middleware.Logger())

    app.Get("/", func(c *nice.Context) {
        c.String(200, "Hello, World")
    })

    app.Run(":8080")
}
```

## 示例

https://github.com/nic-chen/nice-example


## 文档目录

* [Nice核心](https://github.com/nic-chen/nice/tree/master/doc/nice.md)
* [路由](https://github.com/nic-chen/nice/tree/master/doc/router.md)
    * [常规路由](https://github.com/nic-chen/nice/tree/master/doc/router.md#常规路由)
    * [路由语法](https://github.com/nic-chen/nice/tree/master/doc/router.md#路由语法)
        * [静态路由](https://github.com/nic-chen/nice/tree/master/doc/router.md#静态路由)
        * [参数路由](https://github.com/nic-chen/nice/tree/master/doc/router.md#参数路由)
    * [路由选项](https://github.com/nic-chen/nice/tree/master/doc/router.md#路由选项)
    * [组路由](https://github.com/nic-chen/nice/tree/master/doc/router.md#组路由)
    * [链式处理](https://github.com/nic-chen/nice/tree/master/doc/router.md#链式处理)
    * [命名路由](https://github.com/nic-chen/nice/tree/master/doc/router.md#命名路由)
    * [文件路由](https://github.com/nic-chen/nice/tree/master/doc/router.md#文件路由)
    * [自定义错误](https://github.com/nic-chen/nice/tree/master/doc/router.md#自定义错误)
        * [500错误](https://github.com/nic-chen/nice/tree/master/doc/router.md#500错误)
        * [404错误](https://github.com/nic-chen/nice/tree/master/doc/router.md#404错误)
* [中间件](https://github.com/nic-chen/nice/tree/master/doc/middleware.md)
    * [编写中间件](https://github.com/nic-chen/nice/tree/master/doc/middleware.md#编写中间件)
    * [使用中间件](https://github.com/nic-chen/nice/tree/master/doc/middleware.md#使用中间件)
 
* [依赖注入 DI](https://github.com/nic-chen/nice/tree/master/doc/di.md)
    * [注册](https://github.com/nic-chen/nice/tree/master/doc/di.md#注册)
    * [使用](https://github.com/nic-chen/nice/tree/master/doc/di.md#使用)
        * [日志](https://github.com/nic-chen/nice/tree/master/doc/di.md#日志)
        * [路由](https://github.com/nic-chen/nice/tree/master/doc/di.md#路由)
        * [模板](https://github.com/nic-chen/nice/tree/master/doc/di.md#模板)
* [HTTP上下文](https://github.com/nic-chen/nice/tree/master/doc/context.md)
    * [Request](https://github.com/nic-chen/nice/tree/master/doc/context.md#request)
        * [URL参数](https://github.com/nic-chen/nice/tree/master/doc/context.md#url参数)
        * [路由参数](https://github.com/nic-chen/nice/tree/master/doc/context.md#路由参数)
        * [Cookie](https://github.com/nic-chen/nice/tree/master/doc/context.md#cookie)
        * [文件上传](https://github.com/nic-chen/nice/tree/master/doc/context.md#文件上传)
    * [Response](https://github.com/nic-chen/nice/tree/master/doc/context.md#response)
        * [数据存储](https://github.com/nic-chen/nice/tree/master/doc/context.md#数据存储)
        * [内容输出](https://github.com/nic-chen/nice/tree/master/doc/context.md#内容输出)
        * [有用的函数](https://github.com/nic-chen/nice/tree/master/doc/context.md#有用的函数)

* [日志](https://github.com/nic-chen/nice/tree/master/doc/log.md)
    * [日志接口](https://github.com/nic-chen/nice/tree/master/doc/log.md#日志接口)
    * [日志方法](https://github.com/nic-chen/nice/tree/master/doc/log.md#日志方法)

* [数据库](https://github.com/nic-chen/nice/tree/master/doc/database.md)
* [缓存](https://github.com/nic-chen/nice/tree/master/doc/cache.md)


* [项目结构](https://github.com/nic-chen/nice/tree/master/doc/project.md)
    * [目录结构](https://github.com/nic-chen/nice/tree/master/doc/project.md#目录结构)
    * [控制器](https://github.com/nic-chen/nice/tree/master/doc/project.md#控制器)
    * [数据操作](https://github.com/nic-chen/nice/tree/master/doc/project.md#数据模型)
    * [服务](https://github.com/nic-chen/nice/tree/master/doc/project.md#服务)
    * [配置](https://github.com/nic-chen/nice/tree/master/doc/project.md#配置)

* [微服务](https://github.com/nic-chen/nice/tree/master/doc/micro.md)
