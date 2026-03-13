package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
	"todo/cmd/internal/lists"
	"todo/common/storage"
	"todo/common/tasks"
)

var port int

func main() {
	port = *flag.Int("port", 8080, "Port number")

	db, err := storage.NewSqlite("tmp/test.db")
	if err != nil {
		log.Fatalf("could not create database: %v", err)
		return
	}

	repo, err := tasks.NewRepository(db)
	if err != nil {
		log.Fatalf("could not create repository: %v", err)
		return
	}
	tc := tasks.NewController(repo)
	api := lists.NewController(tc)

	target, err := url.Parse("http://localhost:1313")
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(target)

	mux := http.NewServeMux()

	// Registering a handler for a specific method and path with a variable
	mux.HandleFunc("/{path...}", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
	})

	mux.HandleFunc("GET /api", api.Index)

	mux.HandleFunc("GET /api/{list}", api.List)

	mux.HandleFunc("POST /api/{list}/tasks", api.AddTask)

	mux.HandleFunc("PUT /api/{list}/tasks/{id}", api.MoveTask)

	mux.HandleFunc("POST /api/lists", api.AddList)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), LoggingMiddleware(mux))
	if err != nil {
		log.Println(err)
	}
}
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("[%s] %s %s took %v\n", r.Method, r.URL.Path, r.Proto, time.Since(start))
	})
}
