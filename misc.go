package wtf

import (
	"fmt"
	"github.com/i11cn/go_logger"
	"os"
	"strconv"
	"strings"
)

type (
	Convert string
)

func (s Convert) ToInt() (int, error) {
	ret, err := strconv.ParseInt((string)(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return int(ret), nil
}

func (s Convert) ToInt64() (int64, error) {
	ret, err := strconv.ParseInt((string)(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func (s Convert) ToUInt() (uint, error) {
	ret, err := strconv.ParseUint((string)(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(ret), nil
}

func (s Convert) ToUInt64() (uint64, error) {
	ret, err := strconv.ParseUint((string)(s), 10, 64)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func (s Convert) ToFloat() (float64, error) {
	ret, err := strconv.ParseFloat((string)(s), 64)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func (s Convert) ToBool() (bool, error) {
	ret, err := strconv.ParseBool((string)(s))
	if err != nil {
		return false, err
	}
	return ret, nil
}

func file_exist(name string) bool {
	_, err := os.Stat(name)
	if err == nil {
		return true
	} else if os.IsExist(err) {
		return true
	} else {
		return false
	}
}

func log_access(vhost, client, method, url, ua string, code, length int) {
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
	log.Log(client, method, url, code, length, ua)
}
