package main

import (
	"embed"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

//go:embed index.html
var content embed.FS

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		data, _ := content.ReadFile("index.html")
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(data)
	})

	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api")
		url := "http://localhost:8080" + path

		req, _ := http.NewRequestWithContext(r.Context(), r.Method, url, r.Body)
		for k, v := range r.Header {
			req.Header[k] = v
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		for k, v := range resp.Header {
			w.Header()[k] = v
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})

	log.Println("Demo: http://localhost:3000")

	const readHeaderTimeout = 10 * time.Second
	server := &http.Server{
		Addr:              ":3000",
		ReadHeaderTimeout: readHeaderTimeout,
	}
	log.Fatal(server.ListenAndServe())
}
