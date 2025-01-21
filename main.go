package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	mux.Handle("/assets/", http.StripPrefix("/app/assets", http.FileServer(http.Dir("./assets/logo.png"))))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	err := server.ListenAndServe() 
    if err != nil {
        fmt.Println("Error starting server:", err)
    }
}