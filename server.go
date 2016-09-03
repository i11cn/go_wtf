package wtf

import (
	"net/http"
)

type (
	Server struct {
		Mux
		creator_req     func(*http.Request) (*Request, error)
		creator_resp    func() Response
		creator_session func() Session
		starter         func() error
	}
)

func NewServer() *Server {
	ret := &Server{}
	ret.Mux = NewRegexMux()
	ret.starter = func() error {
		return http.ListenAndServe(":80", ret)
	}
	return ret
}

func (s *Server) SetMux(m Mux) {
	s.Mux = m
}

func (s *Server) SetRequestCreator(fn func(*http.Request) (*Request, error)) {
	s.creator_req = fn
}

func (s *Server) SetResponseCreator(fn func() Response) {
	s.creator_resp = fn
}

func (s *Server) SetSessionCreator(fn func() Session) {
	s.creator_session = fn
}

func (s *Server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	c_resp := s.creator_resp()
	c_req, err := s.creator_req(req)
	if err != nil {
		c_resp.WriteHeader(500)
		return
	}
	ctx := &Context{c_req, c_resp, s.creator_session()}
	fn := s.Match(ctx)
	if fn != nil {
		fn(ctx)
	} else {
		resp.WriteHeader(404)
	}
}

func (s *Server) Listen(addr string) {
	s.starter = func() error {
		return http.ListenAndServe(addr, s)
	}
}

func (s *Server) ListenTLS(addr, cert, key string) {
	s.starter = func() error {
		return http.ListenAndServeTLS(addr, cert, key, s)
	}
}
