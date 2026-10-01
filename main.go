package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Client struct {
	Count       int
	ResetTime   time.Time
}

var (
	clients = make(map[string]*Client)
	mu      sync.Mutex
)

const (
	Limit  = 5
	Window = time.Minute
)

func rateLimiter(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ip := r.RemoteAddr

		mu.Lock()
		defer mu.Unlock()

		client, exists := clients[ip]

		if !exists || time.Now().After(client.ResetTime) {
			clients[ip] = &Client{
				Count:     1,
				ResetTime: time.Now().Add(Window),
			}

			next(w, r)
			return
		}

		if client.Count >= Limit {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}

		client.Count++

		next(w, r)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Request allowed!")
}

func main() {

	http.HandleFunc("/", rateLimiter(homeHandler))

	fmt.Println("Server running on http://localhost:8080")

	http.ListenAndServe(":8080", nil)
}