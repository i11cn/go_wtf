package wtf

import (
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

func DetectMime(data []byte, name ...string) string {
	ret := http.DetectContentType(data)
	if len(data) > 512 {
		return ret
	}
	if ret == "application/octet-stream" {
		if len(name) > 0 && name[0] != "" {
			p := strings.Split(name[0], ";")
			return strings.TrimSpace(p[0])
		}
		if len(name) > 1 && name[1] != "" {
			tmp := mime.TypeByExtension(filepath.Ext(name[0]))
			if tmp != "" {
				return tmp
			}
		}
	}
	return ret
}

func MimeIsText(m string) bool {
	m = strings.ToUpper(m)
	if strings.HasPrefix(m, "TEXT/") {
		return true
	}
	if strings.HasSuffix(m, "JSON") {
		return true
	}
	if strings.HasSuffix(m, "XML") {
		return true
	}
	if strings.HasSuffix(m, "JAVASCRIPT") {
		return true
	}
	return false
}
