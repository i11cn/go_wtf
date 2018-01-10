package wtf

import (
	"compress/gzip"
	"net/http"
	"strings"
)

type (
	wtf_gzip_ctx struct {
		Context
		w     *gzip.Writer
		check bool
		zip   bool
	}

	WTFGzipMidware struct {
		level int
		mime  []string
	}

	wtf_cors_midware struct {
	}
)

func (gc *wtf_gzip_ctx) Write(data []byte) (int, error) {
	if gc.check {
		// TODO: 检查是否需要压缩
		ct := gc.Header().Get("Content-Type")
		if ct == "" {
			ct = http.DetectContentType(data)
			gc.Header().Set("Content-Type", ct)
		}
		mime := strings.Trim(strings.Split(ct, ";")[0], " ")
		if MimeIsText(mime) {
			gc.zip = true
			gc.Header().Set("Content-Encoding", "gzip")
		}
		gc.check = false
	}
	if !gc.zip {
		return gc.Context.Write(data)
	}
	ret, err := gc.w.Write(data)
	err = gc.w.Flush()
	return ret, err
}

func NewGzipMidware(l int, mime []string) *WTFGzipMidware {
	return nil
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
			return &wtf_gzip_ctx{ctx, w, true, false}
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
