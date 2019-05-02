// WTF的目标：简洁，组件化。
// 因此WTF的各部件，都定义为接口，可以随意替换。
//
//
package wtf

import (
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"reflect"
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
		BindPipe(key string, fn interface{})

		// 加载字符串作为模板
		LoadText(string)

		// 加载文件作为模板，可以同时加载多个文件
		LoadFiles(files ...string)

		// 执行模板，注意模板名称和加载的文件名相同(不包括路径)
		Execute(key string, data interface{}) ([]byte, error)
	}

	// 以REST方式的请求，在URI中定义的参数将会被解析成该结构
	RESTParam struct {
		name  string
		value string
	}

	// REST方式的请求，URI中定义的参数集合
	RESTParams []RESTParam

	// Rest 定义了REST参数相关的操作
	Rest interface {
		// 增加URI参数
		//
		// 对于重名的问题，不在此处考虑，那是使用者需要保证的事
		Append(name, value string) RESTParams

		// 获取命名了的URI参数，没有获取到则返回空字符串
		//
		// 例如：/test/:foo，则命名参数为foo
		//
		// 又如：/test/(?P<name>\d+)，则命名参数为name
		Get(name string) string

		// 按索引获取URI参数，没有获取到则返回空字符串
		//
		// 例如：/test/:foo/(\d+)，第一个参数命名为foo，第二个参数没有命名，只能通过索引取得
		GetIndex(i int) string
	}

	// 定义了Context的一些处理数据，在处理完成后，输出日志时会从该结构中获取所需的数据
	ContextInfo interface {
		RespCode() int
		WriteBytes() int64
	}

	// 定义了Response之后的一些处理数据，在处理完成后，输出日志时会从该结构中获取所需的数据
	ResponseInfo interface {
		RespCode() int
		WriteBytes() int64
	}

	UploadFile interface {
		io.Reader
		io.ReaderAt
		io.Seeker
		io.Closer

		Filename() string
		Size() int64
		ContentType() string
		Header() textproto.MIMEHeader
	}

	// Request 封装了http.Request，去掉了http.Request和Client相关的操作函数，增加了一些优化过的方法
	Request interface {
		// BasicAuth 代理http.Request中的BasicAuth，返回请求头中的验证信息
		BasicAuth() (username, password string, ok bool)

		// Cookie 代理http.Request中的Cookie，返回指定的Cookie
		Cookie(name string) (*http.Cookie, error)

		// Cookkies 代理http.Request中的Cookies，返回所有Cookie
		Cookies() []*http.Cookie

		// MultipartReader 代理http.Request中的MultipartReader，以Reader的形式读取Multipart内容
		MultipartReader() (*multipart.Reader, error)

		// ParseMultipartForm 代理http.Request中的ParseMultipartForm，不过增加可选设置，不设置maxMemory时，默认为16M，并且会自己检测是否需要Parse
		ParseMultipartForm(maxMemory ...int64) error

		// Proto 返回HTTP的版本号
		Proto() (int, int)

		// Referer 返回http.Request.Referer
		Referer() string

		// UserAgent 返回http.Request.UserAgent
		UserAgent() string

		// Method 返回http.Request.Method
		Method() string

		// URL 返回http.Request.URL
		URL() *url.URL

		// GetHeader 返回指定key的Header项，如果key不存在，则返回空字符串
		GetHeader(key string) string

		// ContentLength 返回http.Request.ContentLength，如果为0，表示没有Body，或者不能获取到ContentLength
		ContentLength() int64

		// Host 返回http.Request.Host
		Host() string

		// Forms 返回url请求参数、Post请求参数和Multipart请求参数三个map，并且如果没有解析过，会先解析之后再返回，多次调用不会多次Parse
		Forms() (url.Values, url.Values, *multipart.Form)

		// RemoteAddr 返回http.Request.RemoteAddr
		RemoteAddr() string

		// 获取客户端请求发送来的Body，可以重复获取而不影响已有的数据
		GetBody() (io.Reader, Error)

		// GetBodyData 获取已经传输完成的请求体
		GetBodyData() ([]byte, Error)

		// 将客户端请求发送来的Body解析为Json对象
		GetJsonBody(interface{}) Error

		//
		GetUploadFile(key string) ([]UploadFile, Error)
	}

	// WTF专用的输出结构接口，注意，区别于http.Response，其中定义了一些常用的便利接口。同时Context里也定义了一些接口，因此除非必须，可以仅使用Context接口即可
	Response interface {
		// Header 函数兼容http.ResponseWriter
		Header() http.Header
		// Write 函数兼容http.ResponseWriter
		Write([]byte) (int, error)
		// WriteHeader 函数兼容http.ResponseWriter
		WriteHeader(int)

		// GetResponseInfo 获取输出的一些信息
		GetResponseInfo() ResponseInfo

		// WriteString 输出字符串到客户端
		WriteString(string) (int, error)

		// 向客户端发送数据流中的所有数据
		WriteStream(io.Reader) (int64, error)

		// 将参数格式化成Json，发送给客户端
		WriteJson(interface{}) (int, error)

		// 将参数格式化成XML，发送给客户端
		WriteXml(interface{}) (int, error)

		// SetHeader 设置Header中的项
		SetHeader(key, value string)

		// 向客户端返回状态码, 如果调用时带了body，则忽略WTF默认的状态码对应的body，而返回此处带的body
		StatusCode(code int, body ...string)

		// 返回状态码404，如果调用时带了body，则忽略WTF默认的body，而返回此处带的body
		NotFound(body ...string)

		// 向客户端发送重定向状态码
		Redirect(url string)

		// 通知客户端，继续请求指定的url，如果有body，可以在调用时指定
		Follow(url string, body ...string)

		// 允许跨域请求，如果还允许客户端发送cookie，可以由第二个参数指定，默认为false
		CrossOrigin(Request, ...string)
	}

	Flushable interface {
		// 将缓冲区中的数据写入网络
		Flush() error
	}

	// Context接口整合了很多处理所需的上下文环境，例如用户的请求Request、输出的接口Response、HTML处理模板Template等
	Context interface {
		// 获取日志对象
		Logger() Logger

		// 获取客户端发送的原始请求
		// HttpRequest() *http.Request

		// Request 获取封装后的客户端请求数据
		Request() Request

		// HttpResponse 获取原始的http.ResponseWriter
		HttpResponse() http.ResponseWriter

		// Response 获取封装后的Response
		Response() Response

		// 执行模板，并且返回执行完成后的数据
		Execute(string, interface{}) ([]byte, Error)

		// 设置REST请求的URI参数
		SetRESTParams(RESTParams)

		// 获取REST请求的URI参数
		RESTParams() RESTParams

		// // 向客户端发送StatusCode
		// WriteHeader(int)

		// // 向客户端发送数据
		// Write([]byte) (int, error)

		// // 向客户端发送字符串
		// WriteString(string) (int, error)

		// // 向客户端发送数据流中的所有数据
		// WriteStream(io.Reader) (int64, error)

		// 获取Context的处理信息
		GetContextInfo() ContextInfo
	}

	// Mux接口
	Mux interface {
		// 三个参数依次为处理接口、匹配的模式和匹配的HTTP方法
		Handle(func(Context), string, ...string) Error

		// 检查Request是否有匹配的Handler，如果有，则返回Handler，以及对应模式解析后的URI参数
		Match(*http.Request) (func(Context), RESTParams)
	}

	Midware interface {
		// 插件的优先级，从1-10，数字越低优先级越高,相同优先级的，顺序不保证
		Priority() int

		// 插件的处理函数，并且返回一个Context，作为插件链中下一个插件的输入
		//
		// 如果返回nil，则表示不再继续执行后续的插件了
		Proc(Context) Context
	}

	Builder interface {
		SetRequestBuilder(fn func(Logger, *http.Request) Request) Builder
		SetResponseBuilder(fn func(Logger, http.ResponseWriter, Template) Response) Builder
		SetContextBuilder(fn func(Logger, *http.Request, http.ResponseWriter, Template, Builder) Context) Builder
		SetMuxBuilder(fn func() Mux) Builder

		BuildRequest(Logger, *http.Request) Request
		BuildResponse(Logger, http.ResponseWriter, Template) Response
		BuildContext(Logger, *http.Request, http.ResponseWriter, Template, Builder) Context
		BuildMux() Mux
	}

	// 服务的主体类，是所有功能的入口
	Server interface {
		http.Handler

		// SetBuilder 设置各个组件的Builder方法
		SetBuilder(Builder)

		// GetBuilder 获取Server中当前Builder，可以在获取之后，修改自定义的Builder，再设置回去
		GetBuilder() Builder

		// 设置Server所使用的Logger
		SetLogger(Logger)

		// 设置Server所使用的模板
		SetTemplate(Template)

		// 绑定Handler函数里自定义参数的构造方法
		ArgBuilder(typ string, fn func(Context) reflect.Value)

		// 获取该Server正在使用的模板
		Template() Template

		// 直接设置一个完成状态的Mux
		SetMux(Mux, ...string)

		// 向链条中插入一个Midware
		AddMidware(Midware)

		// 定义一种更灵活的Handle方法，可以根据handler的参数内容调整输入参数，取消Handler结构
		Handle(interface{}, string, ...string) Error
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
	if len(name) > 0 {
		for _, i := range []RESTParam(p) {
			if i.name == name {
				return i.value
			}
		}
	}
	return ""
}

// 按索引获取URI参数
//
// 例如：/test/:foo/(\d+)，第一个参数命名为foo，第二个参数没有命名，只能通过索引取得
func (p RESTParams) GetIndex(i int) string {
	pa := []RESTParam(p)
	if i >= 0 && i < len(pa) {
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
