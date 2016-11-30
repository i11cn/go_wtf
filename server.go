package wtf

import (
	"bytes"
	"fmt"
	"github.com/i11cn/go_logger"
	"net/http"
	"strings"
	"time"
)

type (
	Server struct {
		Mux
		creator_req  func(*http.Request) (*Request, error)
		creator_resp func(http.ResponseWriter) Response
		creator_tpl  func() Template
		starter      func() error
		resp_code    RespCode
		mids         []mid_chain_item
		name         string
		version      string
		ext_info     string
		server_info  string
	}
)

func NewServer() *Server {
	ret := &Server{}
	ret.Mux = NewRegexMux()
	ret.creator_req = NewRequest
	ret.creator_resp = NewResponse
	ret.creator_tpl = func() Template {
		return &default_template{"./templates", nil}
	}
	ret.starter = func() error {
		return http.ListenAndServe(":80", ret)
	}
	ret.resp_code = NewRespCode()
	mid := func(ctx *Context) bool {
		fn := ret.Match(ctx)
		if fn != nil {
			fn(ctx)
		} else {
			ctx.WriteHeader(404)
		}
		return true
	}
	ret.mids = make([]mid_chain_item, 0, 10)
	ret.mids = append(ret.mids, mid_chain_item{"", mid})
	return ret
}

func (s *Server) SetServerInfo(name, version, ext string) *Server {
	s.name = name
	s.version = version
	s.ext_info = ext
	if len(s.name) > 0 {
		var buf bytes.Buffer
		buf.WriteString(s.name)
		if len(s.version) > 0 {
			buf.WriteString("/")
			buf.WriteString(s.version)
		}
		if len(s.ext_info) > 0 {
			buf.WriteString(" (")
			buf.WriteString(s.ext_info)
			buf.WriteString(")")
		}
		s.server_info = buf.String()
	} else {
		s.server_info = ""
	}
	return s
}

func (s *Server) SetMux(m Mux) *Server {
	s.Mux = m
	return s
}

func (s *Server) SetMidware(name string, fn MiddleWare) *Server {
	exist := false
	for _, m := range s.mids {
		if m.name == name {
			m.mid = fn
			exist = true
		}
	}
	if !exist {
		s.mids = append(s.mids, mid_chain_item{name, fn})
	}
	return s
}

func (s *Server) SetRequestCreator(fn func(*http.Request) (*Request, error)) *Server {
	s.creator_req = fn
	return s
}

func (s *Server) SetResponseCreator(fn func(http.ResponseWriter) Response) *Server {
	s.creator_resp = fn
	return s
}

func (s *Server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	start := time.Now()
	forward := req.Header.Get("X-Forwarded-For")
	if len(forward) > 0 {
		addrs := strings.Split(forward, ",")
		req.RemoteAddr = strings.TrimSpace(addrs[0])
	}
	c_resp := s.creator_resp(resp)
	c_tpl := s.creator_tpl()
	c_req, err := s.creator_req(req)
	ctx := &Context{c_req, c_resp, c_tpl}
	if err != nil {
		ctx = nil
		c_resp.WriteHeader(500)
	} else {
		stop := false
		for i := len(s.mids) - 1; !stop && i >= 0; i-- {
			stop = s.mids[i].mid(ctx)
		}
	}
	if c_resp.RespCode() != 200 && c_resp.Empty() {
		if exist, _, fn := s.resp_code.GetResp(c_resp.RespCode()); exist {
			data := fn(ctx)
			_, err := c_resp.Write(data)
			if err != nil {
				logger.GetLogger("wtf").Error("返回响应时发生了错误: ", err.Error())
				return
			}
		}
	}
	if len(s.server_info) > 0 {
		c_resp.Header().Set("Server", s.server_info)
	}
	if len, err := c_resp.Flush(); err != nil {
		logger.GetLogger("wtf").Error("返回响应时发生了错误: ", err.Error())
	} else {
		client := req.RemoteAddr
		pos := strings.Index(client, ":")
		if pos != -1 {
			client = string([]byte(client)[:pos])
		}
		esp := time.Since(start)
		log_access(strings.Replace(strings.ToLower(req.URL.Host), ":", ".", -1), client, req.Method, req.URL.Path, req.Header.Get("User-Agent"), c_resp.RespCode(), len, esp)
		fmt.Println(esp)
	}
}

func (s *Server) Listen(addr string) *Server {
	s.starter = func() error {
		return http.ListenAndServe(addr, s)
	}
	return s
}

func (s *Server) ListenTLS(addr, cert, key string) *Server {
	s.starter = func() error {
		return http.ListenAndServeTLS(addr, cert, key, s)
	}
	return s
}
