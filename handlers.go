package main

import "net/http"

func catchAllAndRouteToStatic() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/":
			http.ServeFile(w, r, "static/index.html")
		case r.URL.Path == "/robots.txt":
			http.ServeFile(w, r, "static/robots.txt")
		default:
			http.ServeFile(w, r, "static/"+r.URL.Path+".html")
		}
	}
}
