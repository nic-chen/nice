# Nice 日志

Nice 中的日志默认使用的是标准包中的 `log`，可以通过 `DI` 来替换全局的日志器。

新的日志器要求实现 `nice.Logger` 接口，并且注册的 `DI` 名称为 `logger`。

## 日志接口

```
type Logger interface {
	Print(v ...interface{})
	Printf(format string, v ...interface{})
	Println(v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
	Fatalln(v ...interface{})
	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
	Panicln(v ...interface{})
}
```

这个接口其实是对标准包`log`的抽象，最基础的日志接口。

## 日志方法

假如你实现了新的日志管理，使用的姿势像这样：

```
app := nice.New()
app.SetDI("logger", newLogger.New())
app.Get("/", func(c *nice.Context) {
    lg := c.DI("logger").(*newLogger.Logger)
    lg.Println("log line")
})
```

其中 `newLogger` 意为你实现的新的日志器。

### 记录日志

`func (b *Nice) Logger() Logger`

Nice 提供的全局日志器可以通过`app.Logger()` 获得到。

举个例子：

```
app := nice.New()
app.Get("/", func(c *nice.Context) {
    lg := c.Nice().Logger()
    lg.Println("log line")
})
```

除了 `Println` 你可以使用日志接口中的所有方法。
