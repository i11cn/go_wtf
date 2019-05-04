package wtf

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type (
	gzip_config struct {
		level    int
		min_size int
		mime     map[string]string
	}

	wtf_gzip_ctx struct {
		Context
		config   gzip_config
		req      Request
		w        *gzip.Writer
		mime_zip *bool
		do_zip   *bool
		total    int
		buf      *bytes.Buffer
	}

	wtf_gzip_writer struct {
		writer WriterWrapper
		config gzip_config
		w      *gzip.Writer
		mime   string
		total  int
		buf    *bytes.Buffer
	}

	wtf_gzip_ctx2 struct {
		Context
		writer WriterWrapper
		resp   Response
	}

	GzipMid struct {
		config gzip_config
	}

	CorsMid struct {
		domains map[string]string
		headers map[string]string
	}

	wtf_statuscode_ctx struct {
		Context
		handle map[int]func(Context)
		code   int
		do     *bool
	}

	StatusCodeMid struct {
		handle map[int]func(Context)
	}
)

func (gw *wtf_gzip_writer) Header() http.Header {
	return gw.Header()
}

func (gw *wtf_gzip_writer) get_mime() string {
	if gw.mime == "" {
		if ct := gw.writer.Header().Get("Content-Type"); ct != "" {

		}
		data := gw.buf.Bytes()
		gw.mime = http.DetectContentType(data)
		if gw.total < 512 && gw.mime == "application/octet-stream" {
			gw.mime = ""
		}
	}
	return gw.mime
}

func (gw *wtf_gzip_writer) is_mime_need_zip(mime string) (ret bool) {
	if gw.config.mime == nil || len(gw.config.mime) == 0 {
		ret = MimeIsText(mime)
	} else {
		_, ret = gw.config.mime[mime]
	}
	return
}

func (gw *wtf_gzip_writer) Write(in []byte) (int, error) {
	if gw.w != nil {
		return gw.w.Write(in)
	}
	// TODO: 此处需要缓存起来，检查之后进行gzip，然后再写入writer
	gw.total += len(in)
	return 0, nil
}

func (gw *wtf_gzip_writer) WriteHeader(code int) {
	gw.WriteHeader(code)
}

func (gw *wtf_gzip_writer) GetWriteInfo() WriteInfo {
	return gw.GetWriteInfo()
}

func (gw *wtf_gzip_writer) Flush() error {
	return nil
}

func (gc *wtf_gzip_ctx2) HttpResponse() http.ResponseWriter {
	return gc.writer
}

func (gc *wtf_gzip_ctx2) Response() Response {
	return gc.resp
}

func (gc *wtf_gzip_ctx) check_data_mime(data []byte) (string, []byte) {
	if gc.buf == nil {
		return http.DetectContentType(data), data
	}
	w_len := 512 - gc.buf.Len()
	gc.buf.Write(data[:w_len])
	return http.DetectContentType(gc.buf.Bytes()), data[w_len:]
}

func (gc *wtf_gzip_ctx) is_mime_need_zip(mime string) (ret bool) {
	if gc.config.mime == nil || len(gc.config.mime) == 0 {
		ret = MimeIsText(mime)
	} else {
		_, ret = gc.config.mime[mime]
	}
	return
}

func (gc *wtf_gzip_ctx) check_content_type() (ret *bool) {
	ct := gc.Request().GetHeader("Content-Type")
	if ct != "" {
		ret = new(bool)
		mime := strings.Trim(strings.Split(ct, ";")[0], " ")
		*ret = gc.is_mime_need_zip(mime)
	}
	return
}

func (gc *wtf_gzip_ctx) write_buffer_data(out io.Writer, data ...[]byte) (int, error) {
	if gc.buf != out && gc.buf != nil {
		if gc.buf.Len() > 0 {
			if _, err := gc.buf.WriteTo(out); err != nil {
				return 0, err
			}
		}
		gc.buf = nil
	}
	if len(data) > 0 {
		return out.Write(data[0])
	}
	return 0, nil
}

func (gc *wtf_gzip_ctx) write_and_check(data []byte) (int, error) {
	gc.mime_zip = gc.check_content_type()
	if gc.mime_zip == nil {
		// 检查写入的数据大小是否超过最小限制，如果超过最小限制，则需要创建gzip缓冲区，把数据转入gzip缓冲区
		if gc.total < 512 {
			// 数据太少，先缓存起来
			if gc.buf == nil {
				gc.buf = &bytes.Buffer{}
			}
			return gc.buf.Write(data)
		}
		// 检查数据是否支持压缩
		gc.mime_zip = new(bool)
		var mime string
		mime, data = gc.check_data_mime(data)
		gc.Response().SetHeader("Content-Type", mime)
		*gc.mime_zip = gc.is_mime_need_zip(mime)
	}
	var out io.Writer
	if gc.do_zip == nil {
		if *gc.mime_zip {
			// 需要gzip，但是还需要检查是否达到min_size
			if gc.total < gc.config.min_size {
				if gc.buf == nil {
					gc.buf = &bytes.Buffer{}
				}
			} else {
				gc.do_zip = new(bool)
				*gc.do_zip = true
			}
		} else {
			// 不需要gzip，直接输出
			gc.do_zip = new(bool)
			*gc.do_zip = false
		}
	}
	if gc.do_zip == nil {
		out = gc.buf
	} else if *gc.do_zip {
		// 创建gzip的Buffer，写入原来的所有数据，写入data
		if gc.w == nil {
			use, err := gzip.NewWriterLevel(gc.Response(), gc.config.level)
			if err != nil {
				use = gzip.NewWriter(gc.Response())
			}
			gc.w = use
			gc.Response().Header().Del("Content-Length")
			gc.Response().SetHeader("Content-Encoding", "gzip")
		}
		out = gc.w
	} else {
		// 把原来的数据和data写入Context
		out = gc.Response()
	}
	return gc.write_buffer_data(out, data)
}

func (gc *wtf_gzip_ctx) Write(data []byte) (int, error) {
	gc.total += len(data)
	if gc.do_zip == nil {
		return gc.write_and_check(data)
	}
	if *gc.do_zip {
		// 需要gzip，输出到gzip的Buffer
		return gc.write_buffer_data(gc.w, data)
	} else {
		// 直接输出到Context
		return gc.write_buffer_data(gc.Response(), data)

	}
}

func (gc *wtf_gzip_ctx) WriteString(str string) (n int, err error) {
	return gc.Write([]byte(str))
}

func (gc *wtf_gzip_ctx) WriteStream(in io.Reader) (int64, error) {
	// 先用最low的方法实现，以后再优化
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, in); err != nil {
		return 0, err
	}
	ret, err := gc.Write(buf.Bytes())
	return int64(ret), err
}

func (gc *wtf_gzip_ctx) Flush() error {
	if gc.do_zip == nil {
		gc.do_zip = new(bool)
		*gc.do_zip = false
	}
	if *gc.do_zip {
		if _, err := gc.write_buffer_data(gc.w); err != nil {
			return err
		}
		return gc.w.Flush()
	} else {
		if _, err := gc.write_buffer_data(gc.Response()); err != nil {
			return err
		}
	}
	return nil
}

func NewGzipMidware(level ...int) *GzipMid {
	ret := &GzipMid{}
	ret.config.level = gzip.DefaultCompression
	if len(level) > 0 {
		ret.config.level = level[0]
	}
	ret.config.min_size = 512
	return ret
}

func (gm *GzipMid) SetLevel(level int) *GzipMid {
	if level < gzip.BestSpeed {
		level = gzip.NoCompression
	}
	if level > gzip.BestCompression {
		level = gzip.BestCompression
	}
	gm.config.level = level
	return gm
}

func (gm *GzipMid) SetMime(ms []string) *GzipMid {
	if len(ms) > 0 {
		use := make(map[string]string)
		for _, m := range ms {
			use[strings.ToUpper(m)] = m
		}
		gm.config.mime = use
	}
	return gm
}

func (gm *GzipMid) AppendMime(ms string, more ...string) *GzipMid {
	if len(ms) > 0 {
		use := make(map[string]string)
		for _, m := range gm.config.mime {
			use[strings.ToUpper(m)] = m
		}
		use[strings.ToUpper(ms)] = ms
		for _, m := range more {
			use[strings.ToUpper(m)] = m
		}
		gm.config.mime = use
	}
	return gm
}

func (gm *GzipMid) SetMinSize(size int) *GzipMid {
	gm.config.min_size = size
	return gm
}

func (gm *GzipMid) Priority() int {
	return -9
}

func (gm *GzipMid) Proc(ctx Context) Context {
	// 如果压缩率设置为不压缩，则直接返回原来的Context
	if gm.config.level == gzip.NoCompression {
		return ctx
	}
	// 检查对方是否接受压缩
	ecs := ctx.Request().GetHeader("Accept-Encoding")
	for _, ec := range strings.Split(ecs, ",") {
		ec = strings.Trim(strings.ToUpper(ec), " ")
		if ec == "GZIP" {
			ret := &wtf_gzip_ctx{
				Context: ctx,
				config:  gm.config,
			}
			return ret
		}
	}
	return ctx
}

func NewCrossOriginMidware() *CorsMid {
	return &CorsMid{}
}

func (cm *CorsMid) SetDomains(domains []string) *CorsMid {
	if domains != nil && len(domains) > 0 {
		use := make(map[string]string)
		for _, domain := range domains {
			domain = strings.ToUpper(domain)
			use[domain] = domain
		}
		cm.domains = use
	}
	return cm
}

func (cm *CorsMid) AddDomains(domain string, others ...string) *CorsMid {
	if cm.domains == nil {
		cm.domains = make(map[string]string)
	}
	domain = strings.ToUpper(domain)
	cm.domains[domain] = domain
	for _, d := range others {
		d = strings.ToUpper(d)
		cm.domains[d] = d
	}
	return cm
}

func (cm *CorsMid) AddHeader(key, value string) *CorsMid {
	if cm.headers == nil {
		cm.headers = make(map[string]string)
	}
	cm.headers[key] = value
	return cm
}

func (cm *CorsMid) Priority() int {
	return 100
}

func (cm *CorsMid) Proc(ctx Context) Context {
	origin := ctx.Request().GetHeader("Origin")
	if origin == "" {
		return ctx
	}
	if cm.domains != nil {
		if _, ok := cm.domains[origin]; !ok {
			return ctx
		}
	}
	ctx.Response().SetHeader("Access-Control-Allow-Origin", origin)
	if cm.headers == nil {
		ctx.Response().SetHeader("Access-Control-Allow-Credentialls", "true")
		ctx.Response().SetHeader("Access-Control-Allow-Method", "GET, POST, OPTION")
	} else {
		for k, v := range cm.headers {
			ctx.Response().SetHeader(k, v)
		}
	}
	return ctx
}

func (sc *wtf_statuscode_ctx) WriteHeader(code int) {
	sc.code = code
}

func (sc *wtf_statuscode_ctx) Write(data []byte) (int, error) {
	if sc.do == nil {
		sc.do = new(bool)
		*sc.do = true
	}
	return sc.Response().Write(data)
}

func (sc *wtf_statuscode_ctx) WriteString(str string) (int, error) {
	if sc.do == nil {
		sc.do = new(bool)
		*sc.do = true
	}
	return sc.Response().WriteString(str)
}

func (sc *wtf_statuscode_ctx) WriteStream(in io.Reader) (int64, error) {
	if sc.do == nil {
		sc.do = new(bool)
		*sc.do = true
	}
	return sc.WriteStream(in)
}

func (sc *wtf_statuscode_ctx) Flush() error {
	if sc.do != nil && *sc.do {
		sc.Response().WriteHeader(sc.code)
		if sc.handle != nil {
			if h, ok := sc.handle[sc.code]; ok {
				h(sc.Context)
			}
		}
		*sc.do = false
	}
	return nil
}

func NewStatusCodeMidware() *StatusCodeMid {
	return &StatusCodeMid{}
}

func (sc *StatusCodeMid) Handle(code int, h func(Context)) *StatusCodeMid {
	if sc.handle == nil {
		sc.handle = make(map[int]func(Context))
	}
	sc.handle[code] = h
	return sc
}

func (sc *StatusCodeMid) Priority() int {
	return 99
}

func (sc *StatusCodeMid) Proc(ctx Context) Context {
	ret := &wtf_statuscode_ctx{
		Context: ctx,
		handle:  sc.handle,
		code:    http.StatusOK,
	}
	return ret
}
