package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

func main() {
	publicKey, err := getRsaPublicKey()
	if err != nil {
		log.Fatalf("Failed to load RSA public key: %v", err)
	}

	port := os.Getenv("PORT")

	authService := &AuthService{publicKey: publicKey}

	mux := http.NewServeMux()

	mux.Handle("/users/", JWTMiddleware(authService, proxyHandler("http://user-service:8081")))
	mux.Handle("/auth/", proxyHandler("http://user-service:8081"))
	mux.Handle("/chats/", JWTMiddleware(authService, http.StripPrefix("/chats", proxyHandler("http://chat-service:8082"))))
	mux.Handle("/media/", JWTMiddleware(authService, http.StripPrefix("/media", proxyHandler("http://media-service:8083"))))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	log.Printf("API Gateway listening on port %s...", port)
	log.Fatal(server.ListenAndServe())
}

func proxyHandler(target string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url, err := url.Parse(target)
		if err != nil {
			http.Error(w, "Invalid target URL", http.StatusInternalServerError)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, fmt.Sprintf("Proxy error: %v", err), http.StatusBadGateway)
		}
		proxy.ServeHTTP(w, r)
	})
}
