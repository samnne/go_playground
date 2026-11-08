package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"urlshortner"
)

func main() {

	pathToUrls := map[string]string{
		"/yt": "https://www.youtube.com",
		"/se": "https://samechefu.com",
	}
	fallback := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
	http.HandleFunc("/shorten", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pathToUrls)

	})
	http.HandleFunc("/", urlshortner.MapHandler(pathToUrls, fallback))
	fmt.Println("Starting the server on 8080")
	http.ListenAndServe(":8080", nil)

}
