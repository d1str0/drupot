package main

import (
	"net/http"
)

func staticHandler(app App) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[1:]
		if path == "CHANGELOG.txt" {
			http.ServeFile(w, r, app.ChangelogFilepath)
		} else {
			http.ServeFile(w, r, "static/index.html")
		}
	}
}

func routes(app App) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", staticHandler(app))
	return mux
}
