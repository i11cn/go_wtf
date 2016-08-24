package wtf

import (
	"github.com/i11cn/go_logger"
	"time"
)

func init() {
	log := logger.GetLogger("wtf")
	log.AddAppender(logger.NewSplittedFileAppender("[%T] [%N-%L] %f@%F.%l: %M", "wtf.log", 24*time.Hour))
	log.SetLevel(logger.ALL)

	log = logger.GetLogger("wtf_access")
	log.AddAppender(logger.NewSplittedFileAppender("[%T] - %m %m %m", "wtf_access.log", 24*time.Hour))
	log.SetLevel(logger.LOG)
}
