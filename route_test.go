package wtf

import (
	"fmt"
	. "net/http"
	"testing"
)

func TestAddEntry(t *testing.T) {
	router := new(default_router)
	router.Get("test", func(*Context) { fmt.Println("GET") })
	router.Post("/test", func(*Context) { fmt.Println("POST") })
}

func TestDumpEntry(t *testing.T) {
	router := new(default_router)
	router.Get("test", func(*Context) { fmt.Println("GET") })
	router.Post("/test", func(*Context) { fmt.Println("POST") })
	for i, r := range router.router {
		fmt.Println(i, r.pattern)
		for k, _ := range r.entry {
			fmt.Println("    ", k)
		}
	}
}
