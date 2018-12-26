# Nice 项目结构

项目结构建议及相关说明。
一个项目建议只起一个服务及网关接口。

## 目录结构

```
project
  |-- api
    |-- controller
    |-- router.go
  |-- cmd
    |-- srv
    |-- mian.go
  |-- config
  |-- dao
  |-- proto
  |-- services
  |-- README.md
```

结构说明

| 路径                         | 说明          | 备注   |
|-----------------------------|--------------|--------|
| api                         | api网关       | --     |
| api/controller              | 业务控制器目录 | --     |
| cmd                         | 命令行程序     | --     |
| cmd/srv                     | 各服务调起程序 | --     |
| cmd/mian.go                 | 命令行程序入库 | --     |
| config                      | 配置文件目录   | --     |
| dao                         | 数据操作目录   | --     |
| proto                       | pb消息定义目录 | --     |
| services                    | 服务目录      | 应用的主要逻辑都放这里 |
| README.md                   | 应用说明      | --     |

完整结构，参见示例 [example](https://github.com/nic-chen/nice-example)
你也可以按自己需求进行目录的调整

## 控制器

控制器中按业务划分成了不同的文件，不同的操作还应该有不同的方法对应，在实现上有两种考虑：

- 一个控制器中所有方法都是函数，使用控制器的名字作为函数名前置防止多个控制中的命名冲突。
- 将一个控制器视为一个类，所有方法都是类的方法，虽然Go中没有明确的类，但也可以实现面向对象编程。

两种声音都有支持，你可以根据自己喜欢来做，我们选择了第二种姿势，看起来更舒服一些。

最终，一个控制文件可能是这样的：

```
// api/controller/index.go

package controller

import (
  "context"
  "encoding/json"
  "github.com/nic-chen/nice"
  "../../dao"
  "../../config"

  proto "../../proto/member"
)

type member struct{}
var Member = member{}

func (member) Info(c *nice.Context) {
  id := c.ParamInt("id");

  n := nice.Instance(config.APP_NAME);
  d := dao.NewMemberDao();

  m, _ := d.Fetch(id);

  if len(m)>0 {
    delete(m, "password")
    delete(m, "salt")
  }

  RenderJson(c, 0, "", m)
}

....

```

> 该文件来自示例程序 [api](http://github.com/nic-chen/nice-example/tree/master/api)

为了实现面向对象，创建了一个空的结构体作为方法的承载，所有方法都注册给这个结构体。

**需要解释的一句是，为什么还要声明一个 `IndexController` 呢？**

路由注册时需要将每一个URL对应到具体的方法上来，结构体的方法是不能直接用的，需要先声明一个结构体实例才能使用。

在哪儿声明呢？一个是路由注册的时候，一个是控制器定义的时候，我们选择了在控制器定义的时候声明，作为控制器开发的一个规范，路由定义时引入包就可以用了。


## 数据操作

Nice本身不提供数据模型的处理，在 [dao](http://github.com/nic-chen/nice-example/tree/master/dao) 示例中提供了mysql和redis操作参考。

在dao/base.go中定义了最简单的数据操作，fetch（根据主键获取数据，先取缓存，缓存不存在再取数据库并写缓存）、insert（插入一条记录）、update（通过主键更新一条记录并删除缓存）、delete（通过主键删除一条记录并删除缓存）。

举个例子，member表操作如下：

```
// dao/member.go 定义

package dao

func NewMemberDao() *Tbl {
  m := &Tbl{
    Name: "member",   //表名
    Key: "id",        //主键字段名
    //Cols: new(map[string]interface{}),
  }

  return m;
}
```
调用：
```
  
import (
  "../../dao"
)

d := dao.NewMemberDao();
m, _ := d.Fetch(1);

```

**member.go中的定义，表名和主键字段名，fetch、delete、update等依赖这两个配置。**


## 配置文件

应用配置文件，我比较喜欢将配置以独立文件进行硬编码。你也可以自行采用yaml、yml等格式配置然后解析。


## 打包发布

Go程序的一个好处就是，`go build`然后生成一个二进制文件，Copy到服务上就行了。

也可以像例子里写个Makefile，直接make build、make run即可编译或者运行。

至于怎么发布，用什么发布系统都没关系。不过要注意，编译的系统环境和运行的系统环境要一致，mac下编译出来的，linux可不一定能运行。


### 运行

执行 `pathto/builded-file` 即可。需要调试时可以先设置环境变量，如：

```

NICE_ENV=test

```

