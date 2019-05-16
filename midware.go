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

	wtf_gzip_writer struct {
		WriterWrapper
		config gzip_config
		w      io.Writer
		mime   string
		total  int
		buf    *bytes.Buffer
	}

	GzipMid struct {
		config gzip_config
	}

	CorsMid struct {
		domains map[string]string
		headers map[string]string
	}

	wtf_statuscode_writer struct {
		WriterWrapper
		ctx    Context
		handle map[int]func(Context)
		total  int
	}

	StatusCodeMid struct {
		handle map[int]func(Context)
	}
)

func (gw *wtf_gzip_writer) get_mime() string {
	if gw.mime == "" {
		if ct := gw.Header().Get("Content-Type"); ct != "" {
			p := strings.Split(ct, ";")
			gw.mime = strings.TrimSpace(p[0])
		}
	}
	if gw.mime == "" {
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
	ret, err := gw.buf.Write(in)
	if err == nil {
		gw.total += ret
		mime := gw.get_mime()
		if gw.total >= gw.config.min_size && mime != "" {
			if gw.is_mime_need_zip(mime) {
				if gw.Header().Get("Content-Type") == "" {
					gw.Header().Set("Content-Type", mime)
				}
				gw.Header().Del("Content-Length")
				gw.Header().Set("Content-Encoding", "gzip")
				gw.w, err = gzip.NewWriterLevel(gw.WriterWrapper, gw.config.level)
				if err != nil {
					gw.w = gzip.NewWriter(gw.WriterWrapper)
				}
			} else {
				gw.w = gw.WriterWrapper
			}
			gw.buf.WriteTo(gw.w)
		}
	}
	return ret, nil
}

func (gw *wtf_gzip_writer) Flush() (err error) {
	if gw.w == nil {
		if gw.buf.Len() > 0 {
			gw.w = gw.WriterWrapper
			_, err = gw.buf.WriteTo(gw.w)
		}
	} else {
		if gzipw, ok := gw.w.(*gzip.Writer); ok {
			err = gzipw.Flush()
		}
	}
	return
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
		ec = strings.ToUpper(strings.TrimSpace(ec))
		if ec == "GZIP" {
			writer := &wtf_gzip_writer{
				WriterWrapper: ctx.HttpResponse(),
				config:        gm.config,
				buf:           &bytes.Buffer{},
			}
			ret := ctx.Clone(writer)
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
	return -1
}

func (cm *CorsMid) Proc(ctx Context) Context {
	origin := ctx.Request().GetHeader("Origin")
	if origin == "" {
		return ctx
	}
	if cm.domains != nil {
		if _, ok := cm.domains[strings.ToUpper(origin)]; !ok {
			return ctx
		}
	}
	resp := ctx.Response()
	resp.SetHeader("Access-Control-Allow-Origin", origin)
	if cm.headers == nil {
		resp.SetHeader("Access-Control-Allow-Credentialls", "true")
		resp.SetHeader("Access-Control-Allow-Method", "GET, POST, OPTION")
	} else {
		for k, v := range cm.headers {
			resp.SetHeader(k, v)
		}
	}
	return ctx
}

func (sw *wtf_statuscode_writer) Write(in []byte) (int, error) {
	sw.total += len(in)
	return sw.WriterWrapper.Write(in)
}

func (sw *wtf_statuscode_writer) Flush() error {
	info := sw.WriterWrapper.GetWriteInfo()
	if sw.total == 0 && sw.handle != nil {
		if h, exist := sw.handle[info.RespCode()]; exist {
			h(sw.ctx)
		}
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
	return 1
}

func (sc *StatusCodeMid) Proc(ctx Context) Context {
	writer := &wtf_statuscode_writer{
		WriterWrapper: ctx.HttpResponse(),
		ctx:           ctx,
		handle:        sc.handle,
	}
	ret := ctx.Clone(writer)
	return ret
}
