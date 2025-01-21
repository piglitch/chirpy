package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(".")))
	mux.Handle("/assests", http.FileServer(http.Dir("./assets/logo.png")))
	// mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request){

	// })
	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	err := server.ListenAndServe() // nil means use DefaultServeMux
    if err != nil {
        fmt.Println("Error starting server:", err)
    }
}