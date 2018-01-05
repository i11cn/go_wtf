package wtf

import (
	"strings"
)

var (
	all_methods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "TRACE"}
)

func AllSupportMethods() []string {
	return all_methods
}

func ValidMethod(m string) bool {
	if len(m) > 0 {
		m = strings.ToUpper(m)
		for _, s := range AllSupportMethods() {
			if m == s {
				return true
			}
		}
	}
	return false
}
