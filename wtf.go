package wtf

import (
	"github.com/i11cn/go_logger"
	"net/http"
	"time"
)

type (
	Config interface {
	}
	Server interface {
	}
	Request interface {
		Method() string
		AuthInfo() (user, pass string, ok bool)
		Cookie(name string) (*http.Cookie, error)
		Cookies() []*http.Cookie
		Referer() string
		Host() string
		Uri() string
		Body() []byte
	}
	Response interface {
	}
	Session interface {
	}
	Application interface {
	}
	Context2 interface {
	}
)

func init() {
	log := logger.GetLogger("wtf")
	log.AddAppender(logger.NewSplittedFileAppender("[%T] [%N-%L] %f@%F.%l: %M", "wtf.log", 24*time.Hour))
	log.SetLevel(logger.ALL)

	log = logger.GetLogger("wtf_access")
	log.AddAppender(logger.NewSplittedFileAppender("[%T] - %m %m %m", "wtf_access.log", 24*time.Hour))
	log.SetLevel(logger.LOG)
}

func log_access(method, url string, code int) {
	log := logger.GetLogger("wtf_access")
	log.Log(code, method, url)
}
