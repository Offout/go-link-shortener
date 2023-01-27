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
	corsConfig := cors.New(cors.Options{
		AllowedHeaders: []string{"Authorization", "Content-Type"},
	})
	fmt.Println("Starting")
	authRegisterHandler := http.HandlerFunc(auth.Register)
	http.Handle("/register", corsConfig.Handler(authRegisterHandler))
	authLoginHandler := http.HandlerFunc(auth.Login)
	http.Handle("/login", corsConfig.Handler(authLoginHandler))
	squeezeSqueezeHandler := http.HandlerFunc(squeeze.Squeeze)
	http.Handle("/squeeze", corsConfig.Handler(authMiddleware(squeezeSqueezeHandler)))
	squeezeRedirectHandler := http.HandlerFunc(squeeze.Redirect)
	http.Handle("/s/", corsConfig.Handler(squeezeRedirectHandler))
	squeezeStatisticsHandler := http.HandlerFunc(squeeze.Statistics)
	http.Handle("/statistics", corsConfig.Handler(authMiddleware(squeezeStatisticsHandler)))
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
