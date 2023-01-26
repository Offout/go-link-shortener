package main

import (
	"fmt"
	"github.com/Offout/go-link-shortener/src/auth"
	"github.com/Offout/go-link-shortener/src/squeeze"
	"github.com/rs/cors"
	"net/http"
)

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
	http.Handle("/register", cors.Default().Handler(authRegisterHandler))
	authLoginHandler := http.HandlerFunc(auth.Login)
	http.Handle("/login", cors.Default().Handler(authLoginHandler))
	squeezeSqueezeHandler := http.HandlerFunc(squeeze.Squeeze)
	http.Handle("/squeeze", cors.Default().Handler(authMiddleware(squeezeSqueezeHandler)))
	squeezeRedirectHandler := http.HandlerFunc(squeeze.Redirect)
	http.Handle("/s/", cors.Default().Handler(squeezeRedirectHandler))
	squeezeStatisticsHandler := http.HandlerFunc(squeeze.Statistics)
	http.Handle("/statistics", cors.Default().Handler(authMiddleware(squeezeStatisticsHandler)))
	err := http.ListenAndServe(":9980", nil)
	if err != nil {
		return
	}
}
