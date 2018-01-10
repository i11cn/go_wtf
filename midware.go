package wtf

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type (
	wtf_gzip_ctx struct {
		Context
		level    int
		w        *gzip.Writer
		mime_ok  bool
		mime_zip bool
		total    int
		mime     map[string]string
		min_size int
		buf      *bytes.Buffer
		set_head bool
		code     int
	}

	WTFGzipMidware struct {
		level    int
		mime     map[string]string
		min_size int
	}

	wtf_cors_midware struct {
	}
)

func (gc *wtf_gzip_ctx) check_mime(data []byte) string {
	if gc.buf.Len() == 0 {
		return http.DetectContentType(data)
	}
	d := [512]byte{}
	pos := 0
	pos = copy(d[:], gc.buf.Bytes())
	if pos < 512 {
		pos = copy(d[pos:], data)
	}
	return http.DetectContentType(d[:])
}

func (gc *wtf_gzip_ctx) need_zip(data []byte) bool {
	if gc.mime_ok {
		return gc.mime_zip
	}
	ct := gc.Header().Get("Content-Type")
	if ct == "" {
		ct = gc.check_mime(data)
		gc.Header().Set("Content-Type", ct)
	}
	gc.mime_ok = true
	mime := strings.Trim(strings.Split(ct, ";")[0], " ")
	if gc.mime == nil || len(gc.mime) == 0 {
		if !MimeIsText(mime) {
			return gc.mime_zip
		}
		gc.mime_zip = true
		return gc.mime_zip
	}
	_, gc.mime_zip = gc.mime[mime]
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

func (gc *wtf_gzip_ctx) Write(data []byte) (int, error) {
	gc.total += len(data)
	min_size := 512
	if gc.min_size != 0 && gc.min_size > min_size {
		min_size = gc.min_size
	}
	if gc.total < min_size {
		return gc.buf.Write(data)
	}
	if !gc.need_zip(data) {
		_, err := gc.flush_buffer(gc.Context)
		if err != nil {
			return 0, err
		}
		gc.Context.WriteHeader(gc.code)
		return gc.Context.Write(data)
	}
	if !gc.set_head {
		gc.Header().Del("Content-Length")
		gc.Header().Set("Content-Encoding", "gzip")
		gc.set_head = true
		gc.Context.WriteHeader(gc.code)
	}
	if gc.w == nil {
		w, err := gzip.NewWriterLevel(gc.Context, gc.level)
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

func (gc *wtf_gzip_ctx) WriteHeader(c int) {
	gc.code = c
}

func (gc *wtf_gzip_ctx) Flush() error {
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

func NewGzipMidware(level ...int) *WTFGzipMidware {
	ret := &WTFGzipMidware{}
	if len(level) > 0 {
		ret.level = level[0]
	}
	ret.min_size = 1024
	return ret
}

func (gm *WTFGzipMidware) SetLevel(level int) *WTFGzipMidware {
	gm.level = level
	return gm
}

func (gm *WTFGzipMidware) SetMime(ms []string) *WTFGzipMidware {
	if len(ms) > 0 {
		use := make(map[string]string)
		for _, m := range ms {
			use[strings.ToUpper(m)] = m
		}
		gm.mime = use
	}
	return gm
}

func (gm *WTFGzipMidware) SetMinSize(size int) *WTFGzipMidware {
	gm.min_size = size
	return gm
}

func (gm *WTFGzipMidware) Priority() int {
	return -10
}

func (gm *WTFGzipMidware) Proc(ctx Context) Context {
	ecs := ctx.Request().Header.Get("Accept-Encoding")
	for _, ec := range strings.Split(ecs, ",") {
		ec = strings.Trim(strings.ToUpper(ec), " ")
		if ec == "GZIP" {
			w, err := gzip.NewWriterLevel(ctx, gm.level)
			if err != nil {
				w = gzip.NewWriter(ctx)
			}
			ret := &wtf_gzip_ctx{
				Context:  ctx,
				level:    gm.level,
				w:        w,
				mime:     gm.mime,
				min_size: gm.min_size,
				buf:      &bytes.Buffer{},
				code:     http.StatusOK,
			}
			return ret
		}
	}
	return ctx
}

func GetCrossOriginMidware() Midware {
	return &wtf_cors_midware{}
}

func (cm *wtf_cors_midware) Priority() int {
	return 100
}

func (cm *wtf_cors_midware) Proc(ctx Context) Context {
	origin := ctx.Request().Header.Get("Origin")
	if len(origin) > 0 {
		ctx.Header().Set("Access-Control-Allow-Origin", origin)
		ctx.Header().Set("Access-Control-Allow-Credentialls", "true")
		ctx.Header().Set("Access-Control-Allow-Method", "GET, POST")
	}
	return ctx
}
