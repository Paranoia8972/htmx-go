package main

import (
	"net/http"
	"path/filepath"
	"runtime"
)

func main() {
	http.Handle("/resources/", http.StripPrefix("/resources", http.FileServer(http.Dir(getPath("./resources")))))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, getPath("index.html"))
	})

	http.HandleFunc("/dynamic-content", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<p>This is dynamically loaded content!</p>"))
	})

	http.ListenAndServe(":8080", nil)
}

func getPath(elem ...string) string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Join(filepath.Dir(b), "..") // Adjusted to navigate up one level
	return filepath.Join(basepath, filepath.Join(elem...))
}
