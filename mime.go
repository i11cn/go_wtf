package wtf

import (
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
)

func DetectMime(stream io.ReadSeeker, name ...string) string {
	var buf [512]byte
	n, _ := io.ReadFull(stream, buf[:])
	ret := http.DetectContentType(buf[:n])
	stream.Seek(0, io.SeekStart)
	if ret == "application/octet-stream" {
		if len(name) > 0 && name[0] != "" {
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
