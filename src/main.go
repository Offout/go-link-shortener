package main

import (
	"github.com/Offout/go-link-shortener/src/auth"
	"github.com/Offout/go-link-shortener/src/squeeze"
	"net/http"
)

func main() {
	http.HandleFunc("/register", auth.Register)
	http.HandleFunc("/login", auth.Login)
	http.HandleFunc("/squeeze", squeeze.Squeeze)
	http.HandleFunc("/s/", squeeze.Redirect)
	http.HandleFunc("/statistics", squeeze.Statistics)
	err := http.ListenAndServe(":9980", nil)
	if err != nil {
		return
	}
}
