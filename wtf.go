package wtf

import (
	"github.com/i11cn/go_logger"
	"time"
)

type (
	Config interface {
	}
	Session interface {
	}
	Application interface {
	}
	WTF struct {
		servers []*Server
	}
)

func init() {
	log := logger.GetLogger("wtf")
	log.AddAppender(logger.NewSplittedFileAppender("[%T] [%N-%L] %f@%F.%l: %M", "wtf.log", 24*time.Hour))
	log.SetLevel(logger.ALL)

	log = logger.GetLogger("access")
	log.AddAppender(logger.NewSplittedFileAppender("%m [%T] %m %m %m %m %m %m", "wtf_access.log", 24*time.Hour))
	log.SetLevel(logger.LOG)
}

func NewWTF() *WTF {
	return &WTF{make([]*Server, 0, 10)}
}

func (w *WTF) AddServer(s *Server) {
	w.servers = append(w.servers, s)
}

func (w *WTF) Start() error {
	if len(w.servers) > 0 {
		total := len(w.servers)
		quit := make(chan error)
		for _, s := range w.servers {
			go func(q chan<- error) {
				q <- s.starter()
			}(quit)
		}
		count := 0
		for err := range quit {
			if err != nil {
				return err
			}
			count++
			if count == total {
				return nil
			}
		}
	}
	return nil
}

func (w *WTF) StartServer(s *Server) error {
	w.AddServer(s)
	return w.Start()
}
