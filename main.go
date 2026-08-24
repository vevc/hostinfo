package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"hostinfo/internal/api"
)

func main() {
	addr := flag.String("addr", envOr("ADDR", ":8080"), "listen address")
	flag.Parse()

	handler := api.NewHandler()
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", handler.Index)
	mux.HandleFunc("GET /api/health", handler.Health)
	mux.HandleFunc("GET /api/info", handler.Info)
	mux.HandleFunc("GET /api/env", handler.Env)
	mux.HandleFunc("GET /api/env/{key}", handler.EnvKey)
	mux.HandleFunc("GET /api/user", handler.User)
	mux.HandleFunc("GET /api/network", handler.Network)
	mux.HandleFunc("GET /api/public-ip", handler.PublicIP)
	mux.HandleFunc("GET /api/system", handler.System)
	mux.HandleFunc("GET /api/host", handler.Host)
	mux.HandleFunc("GET /api/cpu", handler.CPU)
	mux.HandleFunc("GET /api/memory", handler.Memory)
	mux.HandleFunc("GET /api/disk", handler.Disk)
	mux.HandleFunc("GET /api/process", handler.Process)
	mux.HandleFunc("GET /api/workdir", handler.Workdir)
	mux.HandleFunc("GET /api/hostname", handler.Hostname)
	mux.HandleFunc("GET /api/runtime", handler.Runtime)

	server := &http.Server{
		Addr:              *addr,
		Handler:           loggingMiddleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("hostinfo listening on %s", *addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
