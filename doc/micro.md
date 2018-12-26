# 微服务

提供很简单的方式来支持微服务架构.具备以下能力

* 服务发现与注册: 采用基于etcd的方式,支持`HOSTIP`环境变量,简化配置
* 负载均衡: GRPC提供,默认采用round_robin
* 消息格式: 目前为默认的GRPC的protobuf格式
* 消息流: 内置的拦截器支持unary和stream方式
* 跟踪服务: uber jeager

##编写proto
```

syntax = "proto3";

service Member {
    rpc Info (Request) returns (Response);
}

message Request {
    int32 id = 1;
}

message Response {
    int32 id = 1;
    string nickname = 2;
    string avatar = 3;
}

```

## 服务构建

可通过`nice/micro`包提供的默认定义初始化
```    
   
	service := membersrv.NewMemberService()

	listen := net.JoinHostPort(config.SrvHost, config.SrvPort)

	var opts = []micro.Option{
		micro.WithRegistry(register, config.MemberSrvName, listen),
		micro.WithTracer(tracer),
	}


	server, err := micro.NewServer(config.MemberSrvName, opts...)

	if err != nil {
		panic(fmt.Errorf("%s server start error:%s", config.MemberSrvName, err))
	}

	rpc := server.BuildGrpcServer()
	member.RegisterMemberServer(rpc, service)
	
	err = server.Run(rpc, listen); 

```

### 客户端调用

```

	conn := newSrvDialer(config.MemberSrvName)

	n := nice.Instance(config.APP_NAME);
	
	n.Logger().Printf("connecting:%s", config.MemberSrvName);

    //grpc client
	client := proto.NewMemberClient(conn);

	req := &proto.Request{
		Id: id,
	}

	res, err := client.Info(context.Background(), req) 

```

> 具体可参考例子：[example](https://github.com/nic-chen/nice-example)

