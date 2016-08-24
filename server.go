package wtf

import ()

type (
	Server struct {
		creator_req     func() Request
		creator_resp    func() Response
		creator_session func() Session
	}
)

func (s *Server) SetRequestCreator(fn func() Request) {
	s.creator_req = fn
}

func (s *Server) SetResponseCreator(fn func() Response) {
	s.creator_resp = fn
}

func (s *Server) SetSessionCreator(fn func() Session) {
	s.creator_session = fn
}
