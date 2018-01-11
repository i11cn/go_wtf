# go_wtf

各位看官，不要想歪了，WTF 是指一个小型的Web框架：Web Tiny Framework

## 快速上手
---

* 一个最简单的例子：
```
package main

import (
    "github.com/i11cn/go_wtf"
    "net/http"
)

func main() {
    serv := wtf.NewServer()
    serv.HandleFunc(func(ctx wtf.Context){
        ctx.WriteString("点啥都是这一页")
    }, "/*")
    http.ListenAndServe(":4321", serv)
}
```

* 一个稍微复杂点的例子：
```
package main

import (
    "github.com/i11cn/go_wtf"
    "net/http"
)

type (
    my_server struct {
    }
)

func (s *my_server) Hello(ctx wtf.Context) {
    who := ctx.RESTParams().Get("who")
    ctx.WriteString("Hello，" + who)
}

func main() {
    serv := wtf.NewServer()
    my := &my_server{}
    serv.Handle(my.Hello, "/hello/:who")
    serv.HandleFunc(func(ctx wtf.Context){
        ctx.WriteString("点啥都是这一页")
    }, "/*")
    http.ListenAndServe(":4321", serv)
}
```

## 重点来了
---

WTF有一套非常灵活的路由规则（而且这个路由还可以独立创建，改吧改吧就能给其他框架用），这就要重点对路由进行说明

* ### 支持其他路由的通配符格式

> 例如路由： "/user/:uid" 这样的格式，将会匹配uri: "/user/i11cn"、"/usr/who"、"/user/1234" 等等...

看代码：
```
func Hello(ctx wtf.Context) {
    who := ctx.RESTParams().Get("uid")
    ctx.WriteString("Hello，" + who)
}

.
.
.

serv.HandleFunc(Hello, "/user/:uid")
```
> 从这个例子里，也能看到如果要获取RESTful接口中url里的参数，是怎么操作的了：**Context.RESTParams()**

* ### 支持正则表达式

> 从上面那个例子可以看出，对于参数uid，没有办法限定到底是字符串、数字、或是其他，如果还有这种需求，可以考虑用正则表达式来限定，正则表达式完全采用golang自己的regexp里的格式

> 例如路由： "/user/(?P<uid>\d+)"，将会只匹配 "/user/1234"，而 "/user/i11cn"、"/user/who" 等包含非数字的uri不会被匹配到

看代码：
```
func Hello(ctx wtf.Context) {
    who := ctx.RESTParams().Get("uid")
    ctx.WriteString("Hello，" + who)
}

.
.
.

serv.HandleFunc(Hello, "/user/(?P<uid>\\d+))")
```

* ### 通配符匹配

> 如果想匹配任意内容，可以用星号 '\*' 来代替，不过需要注意的是，在模式中，'\*' 之后的内容会被忽略，也就是说 "/user/\*" 和 "/user/\*/else" 是一样的，之后的 "/else" 被忽略掉了

> 另外需要注意的是，'\*' 在匹配顺序中，是排在最后的，即如果前面没有任何路由匹配到，才会最后匹配到 '\*'

* ### 匹配顺序

> 如果匹配列表里，即有纯字符串式的完全匹配模式，又有正则表达式(或者其他路由的那种名称匹配)，还有通配符模式，那么他们的匹配顺序是怎样的？举一个小栗子，各位看官就应该明白了：

```
serv.HandleFunc(func(ctx wtf.Context){
    ctx.WriteString("任何字符，除了 9999")
}, "/user/:id") // 能够匹配 /user/who，除了 /user/9999

serv.HandleFunc(func(ctx wtf.Context){
    ctx.WriteString("9999")
}, "/user/9999") // 匹配 /user/9999，其他都不匹配

serv.HandleFunc(func(ctx wtf.Context){
    ctx.WriteString("很遗憾，不能匹配，因为排在了 /user/:id 的后面...")
}, "/user/(?P<id>\\d+)") // 匹配不到任何url

serv.HandleFunc(func(ctx wtf.Context){
    ctx.WriteString("/user/* 是没戏了，只能搞定 /user/*/... 这样的了")
}, "/user/*") // 匹配不到任何url

serv.HandleFunc(func(ctx wtf.Context){
    ctx.WriteString("所有以上搞不定的，都在这里")
}, "/*") // 除了/user/... 之外，任何url都能匹配
```

> 好了，详细解释一下，:id 这样的格式能够匹配完整的一级，所以 "/user/:id" 将会匹配所有 "/user"的下一级url，当然如果还有第三级目录， "/user/:id" 就无能为力了，而正则表达式 "/user/(?P<id>\d+)" 由于排序在 "/user/:id" 后面，所有的url都被其给截胡了。所以各位亲，一定要注意，**同样都属于正则表达式的模式，匹配范围小的一定要写在前面啊**。

> "/user/\*" 能够匹配所有以 "/user/" 开头的url，不过由于有两级的url全被 "/user/:id" 吸走了，所以它只能匹配三级url了，即在这里等同于 "/user/.../\*"

> 特别的，有一个不包含正则的完全匹配模式 "/user/9999" , 嗯，这一类模式的优先级是最高的，所以完全忽视 "/user/:id" 的存在

> 小结一下匹配顺序： 完全匹配(不包含正则表达式的模式) > 正则表达式模式(注意匹配范围，小的写在前面，大的写在后面) > 通配符 \*

## RESTful里的Method

大家都知道，RESTful很看重Method，不同的Method需要能够交给不同的方法处理，可是上面的路由里都没写Method，没这功能？NO NO NO，这么重要的功能怎么可能漏掉呢

```
serv.HandleFunc(func(ctx wtf.Context){
    ctx.WriteString("这是GET方法来的")
}, "/*", "geT")

serv.HandleFunc(func(ctx wtf.Context){
    ctx.WriteString(ctx.Request().Method)
}, "/*", "post", "PUT")

serv.HandleFunc(func(ctx wtf.Context){
    ctx.WriteString(ctx.Request().Method)
}, "/user/:id", "all")
```

> 看代码，说故事，增加方法好像很简单，是吧

> 首先能感觉到的，是忽略大小写，geT和GET、get是同样的效果

> 再有，如果想匹配所有方法，就写个"all"好了，或者更简单一点，啥都不写...

> 但是注意，这些Method不要拼错了，因为这里的参数，同时支持vhost，或许不正确的Method会被当成vhost，那就糗大了...

## 中间件

中间件的作用其实还是很大的，比如所有的请求都要先验证token，而接口一共有那么2、3千个... 估计程序员不用被杀来祭天，自己就挂了。这个时候，只需要写个中间件，挂在所有请求处理之前，咻，世界干净了。

惯例，看例子：
```
type (
    AuthMid struct {
    }
)

func (am *AuthMid) Priority() int {
    return 0
}
func (am *AuthMid) Proc(ctx wtf.Context) wtf.Context {
    ctx.WriteHeader(http.StatusUnauthorized)
    ctx.WriteString("爷今天不高兴，谁也不让过")
    return nil
}

.
.
.

serv.AddMidware(&AuthMid{})
```

> 其中那个Priority方法，是用来做中间件的排序的，数字越小的越靠前，越早被执行。这个优先级的取值是0~100，超出范围的会被截取到这个范围内，即小于0的认为是0，大于100的认为是100。同样优先级的，按照加入的顺序来执行。

> Proc方法是需要返回一个Context的，这个Context会用来作为下一个中间件的输入，而如果要终止这个处理链，就像这里的AuthMid一样，只需要返回nil，之后所有的中间件就不会执行了。

> 更复杂一点的，如何自定义Context，交给后续的中间件，甚至处理函数，就不啰嗦了，可以自行参考已经有的几个个实现：GzipMid、CorsMid和StatusCodeMid。不过需要注意的是，Context中如果有Flush的需求，一定要保证多次调用Flush不会造成危害。

## 先写这么多

累了，如果还有需要补充的，再接着写吧。

Ejoy WTF ！！！

Bye
