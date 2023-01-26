package main

import (
	"fmt"
	"github.com/Offout/go-link-shortener/src/auth"
	"github.com/Offout/go-link-shortener/src/squeeze"
	"net/http"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if "" != origin {
			w.Header().Add("Access-Control-Allow-Origin", "http://"+r.Host)
			w.Header().Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Authorization")
		}
		next.ServeHTTP(w, r)
	})
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userName = auth.CheckSession(r)
		if "" == userName {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("Starting")
	authRegisterHandler := http.HandlerFunc(auth.Register)
	http.Handle("/register", corsMiddleware(authRegisterHandler))
	authLoginHandler := http.HandlerFunc(auth.Login)
	http.Handle("/login", corsMiddleware(authLoginHandler))
	squeezeSqueezeHandler := http.HandlerFunc(squeeze.Squeeze)
	http.Handle("/squeeze", corsMiddleware(authMiddleware(squeezeSqueezeHandler)))
	squeezeRedirectHandler := http.HandlerFunc(squeeze.Redirect)
	http.Handle("/s/", corsMiddleware(squeezeRedirectHandler))
	squeezeStatisticsHandler := http.HandlerFunc(squeeze.Statistics)
	http.Handle("/statistics", corsMiddleware(authMiddleware(squeezeStatisticsHandler)))
	err := http.ListenAndServe(":9980", nil)
	if err != nil {
		return
	}
}
