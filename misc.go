package wtf

import (
	"fmt"
	"github.com/i11cn/go_logger"
	"strings"
	"time"
)

func log_access(vhost, client, method, url, ua string, code, length int, esp time.Duration) {
	if len(client) < 1 {
		client = "-"
	}
	if len(ua) < 1 {
		ua = "-"
	}
	if strings.Contains(url, " ") {
		url = fmt.Sprintf("\"%s\"", url)
	}
	if strings.Contains(ua, " ") {
		ua = fmt.Sprintf("\"%s\"", ua)
	}
	name := "access"
	if len(vhost) > 0 {
		name = fmt.Sprintf("%s.access", vhost)
	}
	log := logger.GetLogger(name)
	if log.AppenderCount() < 1 {
		log = logger.GetLogger("access")
	}
	log.Log(client, esp, method, url, code, length, ua)
}
