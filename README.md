# short_link

#### 介绍
&nbsp;&nbsp;&nbsp;&nbsp; Go Gin实现的短链接服务。<br>
&nbsp;&nbsp;&nbsp;&nbsp; 本项目用于短链接服务的学习和原理阐述，请勿直接将该项目运用到生产环境中。

#### 博客
&nbsp;&nbsp;&nbsp;&nbsp; [使用Go语言开发一个短链接服务](https://www.cnblogs.com/ALXPS/p/18066568)

#### 软件架构
- 服务依赖组件包括缓存和数据库
- 项目中缓存不依赖特定组件，实现app/component/cache.go的CacheItf接口即可。默认已实现Redis作为缓存
- 项目中数据库不依赖特定组件，实现app/component/database.go的DatabaseItf接口即可。默认已实现MongoDB作为数据库

#### 实现接口
&nbsp;&nbsp;&nbsp;&nbsp; 见接口文档，【[docs/api_doc.json](./docs/api_doc.json)】
- 跳转：GET /:code
- 添加短链接: POST /api/v1/links
- 获取链接详情：GET /api/v1/links

#### 接口用户认证
&nbsp;&nbsp;&nbsp;&nbsp; 项目未实现用户的登录注册等逻辑，因为认为短链接服务应该为一个子系统，用户服务多数情况下设计为一个单独的服务。其它子服务，仅实现用户认证。<br>
&nbsp;&nbsp;&nbsp;&nbsp; 项目实现了一个基于请求头JsonWebToken的认证中间件，添加和获取短链接两个接口需要认证。<br>
&nbsp;&nbsp;&nbsp;&nbsp; 请求需要认证的接口需要在求头加上key "Authorization"， value "Bearer {jwt}"<br>
&nbsp;&nbsp;&nbsp;&nbsp; 下面为一段生成用来测试jwt代码：<br>
```go
package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/1911860538/short_link/config"
)

func main() {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       "1f70a466-1449-4676-b2d7-2037341c718e",
		"name":     "Tom",
		"username": "Tom01",
		"exp":      time.Date(2035, 10, 10, 12, 0, 0, 0, time.UTC).Unix(),
	})

	tokenStr, err := token.SignedString([]byte(config.Conf.Jwt.SecretKey))
	if err != nil {
		panic(err)
		return
	}
	fmt.Println(tokenStr)
}
```
&nbsp;&nbsp;&nbsp;&nbsp; 下面为一个可用来测试的Authorization:
```text
 Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjIwNzU2MzA0MDAsImlkIjoiMWY3MGE0NjYtMTQ0OS00Njc2LWIyZDctMjAzNzM0MWM3MThlIiwibmFtZSI6IlRvbSIsInVzZXJuYW1lIjoiVG9tMDEifQ.wGEVC-9okRKjCWoxJMWF90LM0gKHSdJHVvvuqQMTwbk 
```

#### 启动目录
&nbsp;&nbsp;&nbsp;&nbsp; [cmd/server/main.go](cmd/server/main.go)
