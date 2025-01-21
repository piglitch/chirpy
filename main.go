package main

import (
	"fmt"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	// mux.Handle("/", apihandler{}
	// mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request){

	// })
	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	server.ListenAndServe();
}