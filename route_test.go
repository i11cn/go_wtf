package wtf

import (
	"fmt"
	"testing"
)

func TestAddEntry(t *testing.T) {
	router := new(default_router)
    router.AddEntry("/", "GET", func(*Context) {})
}

func TestDumpEntry(t *testing.T) {
	router := new(default_router)
	for i, r := range router.router {
		fmt.Println(i, r.pattern)
		for k, _ := range r.entry {
			fmt.Println("    ", k)
		}
	}
}
