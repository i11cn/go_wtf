package wtf

import (
	"github.com/i11cn/go_logger"
	"time"
)

func init() {
	log := logger.GetLogger("wtf")
	log.AddAppender(logger.NewSplittedFileAppender("[%T] [%N-%L] %f@%F.%l: %M", "wtf.log", 24*time.Hour))
	log.SetLevel(logger.WARN)
    
    log = logger.GetLogger("wtf_access")
	log.AddAppender(logger.NewSplittedFileAppender("[%T] : %M", "wtf_access.log", 24*time.Hour))
	log.SetLevel(logger.INFO)
}

func log_access(method, url string, code int) {
    log := logger.GetLogger("wtf_access")
    log.Info(method, " ", code, " ", url)
}