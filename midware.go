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
		w        *gzip.Writer
		mime_ok  bool
		mime_zip bool
		do_zip   *bool
		total    int
		buf      *bytes.Buffer
		set_head bool
	}

	GzipMid struct {
		config   gzip_config
		level    int
		mime     map[string]string
		min_size int
	}

	CorsMid struct {
		domains map[string]string
		headers map[string]string
	}

	wtf_statuscode_ctx struct {
		Context
		handle map[int]func(Context)
		code   int
		total  int
	}

	StatusCodeMid struct {
		handle map[int]func(Context)
	}
)

func (gc *wtf_gzip_ctx) check_data_mime(data []byte) (string, []byte) {
	if gc.buf == nil {
		return http.DetectContentType(data), data
	}
	w_len := 512 - gc.buf.Len()
	gc.buf.Write(data[:w_len])
	return http.DetectContentType(gc.buf.Bytes()), data[w_len:]
}

/* check_zip
 * 检查是否需要zip
 * 1 表示需要zip
 * 0 表示不需要zip
 * -1 表示结果还未检测出来
 */
func (gc *wtf_gzip_ctx) check_zip(total int, data []byte) int {
	if gc.mime_ok {
		if gc.mime_zip {
			return 1
		} else {
			return 0
		}
	}
	ct := gc.Header().Get("Content-Type")
	if len(ct) > 0 {
		gc.mime_ok = true
		mime := strings.Trim(strings.Split(ct, ";")[0], " ")
		if gc.config.mime == nil || len(gc.config.mime) == 0 {
			gc.mime_zip = MimeIsText(mime)
			if gc.mime_zip {
				return 1
			} else {
				return 0
			}
		} else {
			_, gc.mime_zip = gc.config.mime[mime]
		}
		if gc.mime_zip {
			return 1
		} else {
			return 0
		}
	}
	if total < 512 {
		return -1
	}
	return 0
}

func (gc *wtf_gzip_ctx) need_zip(data []byte) bool {
	if gc.mime_ok {
		return gc.mime_zip
	}
	ct := gc.Header().Get("Content-Type")
	if ct == "" {
		ct, data = gc.check_data_mime(data)
		gc.Header().Set("Content-Type", ct)
	}
	gc.mime_ok = true
	mime := strings.Trim(strings.Split(ct, ";")[0], " ")
	if gc.config.mime == nil || len(gc.config.mime) == 0 {
		if !MimeIsText(mime) {
			return gc.mime_zip
		}
		gc.mime_zip = true
		return gc.mime_zip
	}
	_, gc.mime_zip = gc.config.mime[mime]
	return gc.mime_zip
}

func (gc *wtf_gzip_ctx) flush_buffer(out io.Writer) (int, error) {
	if gc.buf == nil {
		return 0, nil
	}
	if gc.buf.Len() > 0 {
		n, err := gc.buf.WriteTo(out)
		gc.buf = nil
		return int(n), err
	} else {
		gc.buf = nil
		return 0, nil
	}
}

func (gc *wtf_gzip_ctx) write_buffer_data(out io.Writer, data ...[]byte) (int, error) {
	if gc.buf == nil {
		if len(data) > 0 {
			return out.Write(data[0])
		}
		return 0, nil
	}
	if gc.buf.Len() > 0 {
		_, err := gc.buf.WriteTo(out)
		gc.buf = nil
		if err != nil {
			return 0, err
		}
		if len(data) > 0 {
			return out.Write(data[0])
		}
		return 0, nil
	} else {
		gc.buf = nil
		if len(data) > 0 {
			return out.Write(data[0])
		}
		return 0, nil
	}
}

func (gc *wtf_gzip_ctx) set_header(zip bool) {
	if !gc.set_head {
		if zip {
			gc.Header().Del("Content-Length")
			gc.Header().Set("Content-Encoding", "gzip")
		}
	}
	gc.set_head = true
}

func (gc *wtf_gzip_ctx) write_and_check(data []byte) (int, error) {
	// 检查写入的数据大小是否超过最小限制，如果超过最小限制，则需要创建gzip缓冲区，把数据转入gzip缓冲区
	if gc.total < 512 {
		if gc.buf == nil {
			gc.buf = &bytes.Buffer{}
		}
		return gc.buf.Write(data)
	}
	if gc.w != nil {
		return gc.w.Write(data)
	}
	if gc.total < gc.config.min_size {
		return gc.buf.Write(data)
	}
	gc.set_header(true)
	if gc.w == nil {
		w, err := gzip.NewWriterLevel(gc.Context, gc.config.level)
		if err != nil {
			w = gzip.NewWriter(gc.Context)
		}
		gc.w = w
	}
	_, err := gc.flush_buffer(gc.w)
	if err != nil {
		return 0, err
	}
	return gc.w.Write(data)
}

func (gc *wtf_gzip_ctx) Write(data []byte) (int, error) {
	gc.total += len(data)
	if gc.do_zip == nil {
		return gc.write_and_check(data)
	}
	// 如果不需要gzip，则直接输出
	if !(*gc.do_zip) {
		// 直接输出到Context
		return gc.Context.Write(data)
	}
	// if !gc.need_zip(data) {
	// 	gc.set_header(false)
	// 	_, err := gc.flush_buffer(gc.Context)
	// 	if err != nil {
	// 		return 0, err
	// 	}
	// 	return gc.Context.Write(data)
	// }
	return gc.w.Write(data)
}

func (gc *wtf_gzip_ctx) WriteString(str string) (n int, err error) {
	return gc.Write([]byte(str))
}

func (gc *wtf_gzip_ctx) Flush() error {
	gc.set_header(gc.w != nil)
	if gc.w == nil {
		_, err := gc.flush_buffer(gc.Context)
		return err
	}
	_, err := gc.flush_buffer(gc.w)
	if err != nil {
		return err
	}
	return gc.w.Flush()
}

func NewGzipMidware(level ...int) *GzipMid {
	ret := &GzipMid{}
	ret.config.level = gzip.DefaultCompression
	if len(level) > 0 {
		ret.config.level = level[0]
		ret.level = level[0]
	}
	ret.config.min_size = 512
	ret.min_size = 1024
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
	gm.level = level
	return gm
}

func (gm *GzipMid) SetMime(ms []string) *GzipMid {
	if len(ms) > 0 {
		use := make(map[string]string)
		for _, m := range ms {
			use[strings.ToUpper(m)] = m
		}
		gm.config.mime = use
		gm.mime = use
	}
	return gm
}

func (gm *GzipMid) SetMinSize(size int) *GzipMid {
	gm.config.min_size = size
	gm.min_size = size
	return gm
}

func (gm *GzipMid) Priority() int {
	return -10
}

func (gm *GzipMid) Proc(ctx Context) Context {
	// 如果压缩率设置为不压缩，则直接返回原来的Context
	if gm.config.level == gzip.NoCompression {
		return ctx
	}
	// 检查对方是否接受压缩
	ecs := ctx.Request().Header.Get("Accept-Encoding")
	for _, ec := range strings.Split(ecs, ",") {
		ec = strings.Trim(strings.ToUpper(ec), " ")
		if ec == "GZIP" {
			ret := &wtf_gzip_ctx{
				Context: ctx,
				config:  gm.config,
				buf:     &bytes.Buffer{},
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
	origin := ctx.Request().Header.Get("Origin")
	if origin == "" {
		return ctx
	}
	if cm.domains != nil {
		if _, ok := cm.domains[origin]; !ok {
			return ctx
		}
	}
	ctx.Header().Set("Access-Control-Allow-Origin", origin)
	if cm.headers == nil {
		ctx.Header().Set("Access-Control-Allow-Credentialls", "true")
		ctx.Header().Set("Access-Control-Allow-Method", "GET, POST")
	} else {
		for k, v := range cm.headers {
			ctx.Header().Set(k, v)
		}
	}
	return ctx
}

func (sc *wtf_statuscode_ctx) WriteHeader(code int) {
	sc.code = code
}

func (sc *wtf_statuscode_ctx) Write(data []byte) (int, error) {
	sc.total += len(data)
	return sc.Context.Write(data)
}

func (sc *wtf_statuscode_ctx) Flush() error {
	if sc.total == 0 {
		sc.Context.WriteHeader(sc.code)
		if sc.handle != nil {
			if h, ok := sc.handle[sc.code]; ok {
				h(sc.Context)
			}
		}
		sc.total = -1
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
		total:   0,
	}
	return ret
}
