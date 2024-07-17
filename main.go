package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"timstack/internal/env"
	"timstack/internal/flash"

	"github.com/gorilla/mux"
)

var (
	logger *slog.Logger
)

func main() {
	port := env.GetAsIntElseAlt("PORT", 9005)
	mode := env.GetAsStringElseAlt("ENV", "dev")

	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug, // we toggle this if we're in prod
	}
	var handler slog.Handler = slog.NewTextHandler(os.Stdout, opts)
	if mode == "prod" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}
	logger = slog.New(handler)

	r := mux.NewRouter()

	// Set no caching
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
			wr.Header().Set("Cache-Control", "max-age=0, must-revalidate")
			next.ServeHTTP(wr, req)
		})
	})

	// Setup static file handling
	fs := http.FileServer(http.Dir("static"))
	r.PathPrefix("/static").Handler(http.StripPrefix("/static", fs))

	// Setup the Flash package notification handler
	r.Handle("/notifications", flash.HandlerWithLogger(logger))

	r.HandleFunc("/flash", func(w http.ResponseWriter, r *http.Request) {
		flash.Set(w, flash.Success, "Flash Handler Test! ", "High five 🖐️")
		http.Redirect(w, r, "/", http.StatusFound)
	})

	// Entry route
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		helloWorld := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tim Stack</title>
    <link rel="icon" type="image/x-icon" href="static/img/favicon.ico">
    <script
      src="https://unpkg.com/htmx.org@1.9.10"
      integrity="sha384-D1Kt99CQMDuVetoL1lrYwg5t+9QdHe7NLX/SoJYkXDFfX37iInKRy5xLSi8nO7UC"
      crossorigin="anonymous"
    ></script>
    <link href="/static/css/main.css" rel="stylesheet" />
</head>
<body>
 <div hx-get="/notifications" hx-trigger="load" hx-swap="outerHTML">
       <!-- USE THIS DIV FOR FLASH NOTIFICATIONS -->
</div>  
<div class="container mx-auto h-screen flex flex-col justify-center items-center">
  <h1 class="text-6xl">
    Welcome to
    <strong class="bg-clip-text text-transparent bg-gradient-to-r from-blue-500 to-purple-500">
      TimStack
    </strong> 👋
  </h1>
</div>
</body>
`
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, helloWorld)
	})

	host := fmt.Sprintf("0.0.0.0:%d", port)
	logger.Info("Your app is running on", "host", host)
	log.Fatal(http.ListenAndServe(host, r))
}