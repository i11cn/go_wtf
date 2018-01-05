// WTF的目标：简洁，组件化。
// 因此WTF的各部件，都定义为接口，可以随意替换。
//
//
package wtf

import (
	"io"
	"net/http"
)

type (
	// WTF专有的错误结构，相比标准库里的error，多了Code字段，可以设置自己需要的错误码，同时兼容系统error
	Error interface {
		Error() string
		Code() int
		Message() string
	}

	// WTF所使用的日志接口，如果要替换WTF内置的Logger，只需要实现该接口即可
	Logger interface {
		Trace(...interface{})
		Tracef(string, ...interface{})
		Debug(...interface{})
		Debugf(string, ...interface{})
		Info(...interface{})
		Infof(string, ...interface{})
		Log(...interface{})
		Logf(string, ...interface{})
		Warn(...interface{})
		Warnf(string, ...interface{})
		Error(...interface{})
		Errorf(string, ...interface{})
		Fatal(...interface{})
		Fatalf(string, ...interface{})
	}

	// HTML的模板处理接口
	Template interface {
		// 绑定管道函数
		BindPipe(string, interface{})

		// 加载字符串作为模板
		LoadText(string)

		// 加载文件作为模板，可以同时加载多个文件
		LoadFiles(...string)

		// 执行模板，注意模板名称和加载的文件名相同(不包括路径)
		Execute(string, interface{}) ([]byte, error)
	}

	// 以REST方式的请求，在URI中定义的参数将会被解析成该结构
	RESTParam struct {
		name  string
		value string
	}

	// REST方式的请求，URI中定义的参数集合
	RESTParams []RESTParam

	// 定义了Context的一些处理数据，在处理完成后，输出日志时会从该结构中获取所需的数据
	ContextInfo interface {
		RespCode() int
		WriteBytes() int
	}

	// WTF专用的输出结构接口，注意，区别于http.Response，其中定义了一些常用的便利接口。同时Context里也定义了一些接口，因此除非必须，可以仅使用Context接口即可
	Response interface {
		// 向客户端返回状态码, 如果调用时带了body，则忽略WTF默认的状态码对应的body，而返回此处带的body
		StatusCode(int, ...string)

		// 返回状态码404，如果调用时带了body，则忽略WTF默认的body，而返回此处带的body
		NotFound(...string)

		// 向客户端发送重定向状态码
		Redirect(string)

		// 通知客户端，继续请求指定的url，如果有body，可以在调用时指定
		Follow(string, ...string)

		// 允许跨域请求，如果还允许客户端发送cookie，可以由第二个参数指定，默认为false
		CrossOrigin(string, ...bool)

		// 将参数格式化成Json，发送给客户端
		WriteJson(interface{}) (int, error)

		// 将参数格式化成XML，发送给客户端
		WriteXml(interface{}) (int, error)
	}

	// Context接口整合了很多处理所需的上下文环境，例如用户的请求Request、输出的接口Response、HTML处理模板Template等
	Context interface {
		// 获取日志对象
		Logger() Logger

		// 获取客户端发送的请求
		Request() *http.Request

		// 获取向客户端发送数据的响应对象
		Response() Response

		// 执行模板，并且返回执行完成后的数据
		Execute(string, interface{}) ([]byte, Error)

		// 设置REST请求的URI参数
		SetRESTParams(RESTParams)

		// 获取REST请求的URI参数
		RESTParams() RESTParams

		// 获取客户端请求发送来的Body
		GetBody() ([]byte, Error)

		// 将客户端请求发送来的Body解析为Json对象
		GetJsonBody(interface{}) Error

		// 以下三个方法完全兼容http.ResponseWriter，因此该Context可以直接作为http.ResponseWriter使用
		//
		// 向客户端发送的HTTP头
		Header() http.Header

		// 向客户端发送StatusCode
		WriteHeader(int)

		// 向客户端发送数据
		Write([]byte) (int, error)

		// 向客户端发送字符串
		WriteString(string) (int, error)

		// 向客户端发送数据流中的所有数据
		WriteStream(io.Reader) (int, error)

		// 获取Context的处理信息
		GetContextInfo() ContextInfo
	}

	Handler interface {
		Proc(Context)
	}

	ResponseCode interface {
		Handle(int, func(Context))
		StatusCode(Context, int, ...string)
	}

	// Mux接口
	Mux interface {
		// 三个参数依次为处理接口、匹配的模式和匹配的HTTP方法
		Handle(Handler, string, ...string) Error

		// 检查Request是否有匹配的Handler，如果有，则返回Handler，以及对应模式解析后的URI参数
		Match(*http.Request) ([]Handler, RESTParams)
	}

	Chain interface {
		Proc(Context) bool
	}

	MuxBuilder     func() Mux
	ContextBuilder func(Logger, http.ResponseWriter, *http.Request, ResponseCode, Template) Context

	// 服务的主体类，是所有功能的入口
	Server interface {
		http.Handler

		// 更改Mux的创建方法，如果需要用自己实现的Mux替换WTF默认Mux，需要调用该方法替换Mux的Builder。
		// 注意，该方法必须在所有的Handle方法调用之前调用，否则Mux已经创建完成，再替换Builder已经没有任何效果了。
		SetMuxBuilder(func() Mux)

		// 更改创建Context的方法，注意创建方法需要接收并处理的参数
		SetContextBuilder(func(Logger, http.ResponseWriter, *http.Request, ResponseCode, Template) Context)

		// 设置Server所使用的Logger
		SetLogger(Logger)

		// 设置Server所使用的模板
		SetTemplate(Template)

		// 获取该Server正在使用的模板
		Template() Template

		// 直接设置一个完成状态的Mux
		SetMux(Mux, ...string)

		// 对于Mux的Handle方法的代理，其中做了一些强化，比如参数中可以混合Method和Host
		Handle(Handler, string, ...string) Error

		// 对于Mux的Handle方法的代理，在func之外包装了一层Wrapper
		HandleFunc(func(Context), string, ...string) Error

		// 对于Context和Response要回复给客户端的StatusCode，可以在此处设置专门针对某一StatusCode的处理方法，例如404、500啥的
		HandleStatusCode(int, func(Context))
	}
)

func init() {
}

// 获取命名了的URI参数
//
// 例如：/test/:foo，则命名参数为foo
//
// 又如：/test/(?P<name>\d+)，则命名参数为name
func (p RESTParams) Get(name string) string {
	for _, i := range []RESTParam(p) {
		if i.name == name {
			return i.value
		}
	}
	return ""
}

// 按索引获取URI参数
//
// 例如：/test/:foo/(\d+)，第一个参数命名为foo，第二个参数没有命名，只能通过索引取得
func (p RESTParams) GetIndex(i int) string {
	pa := []RESTParam(p)
	if len(pa) > i {
		return pa[i].value
	}
	return ""
}

// 增加URI参数
//
// 对于重名的问题，不在此处考虑，那是使用者需要考虑的事
func (p RESTParams) Append(name, value string) RESTParams {
	ret := []RESTParam(p)
	ret = append(ret, RESTParam{name, value})
	return RESTParams(ret)
}
