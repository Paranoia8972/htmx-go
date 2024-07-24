package main

import (
	"fmt"
	"net/http"
	"path/filepath"
)

func main() {
	fs := http.FileServer(http.Dir("./resources"))
	http.Handle("/resources/", http.StripPrefix("/resources/", fs))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		htmlFilePath := filepath.Join("src", "index.html")
		http.ServeFile(w, r, htmlFilePath)
	})

	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, HTMX!")
	})

	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
