package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileServerHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

// func (cfg *apiConfig) displayServerHits() http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w, "Hits: %d", cfg.fileServerHits.Load())
// 	})
// }

// func (cfg *apiConfig) resetServerHitCount() http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprintf(w, "Hits: %d", cfg.fileServerHits.Swap(0))
// 	})
// }

func main() {
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}
	rootHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(rootHandler))
	mux.Handle("/assets/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("./assets/"))))

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(fmt.Sprintf("<html><body> <h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", apiCfg.fileServerHits.Load())))
	})
	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		apiCfg.fileServerHits.Swap(0)
		w.Write([]byte(fmt.Sprintf("Hits: %d", 0)))
	})

	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		type chirp struct {
			Body string `json:"body"`
		}
		type errStruct struct {
			Error string `json:"error"`
		}
		type validMsg struct {
			Valid bool `json:"valid"`
		}
		decoder := json.NewDecoder(r.Body)
		params := chirp{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(500)
			return
		}
		if len(params.Body) > 140 {
			log.Printf("Chirp too long")
			errorMsg := errStruct{
				Error: "Chirp is too long",
			}
			data, err := json.Marshal(errorMsg)  
			if err != nil {
				log.Printf("Error marshalling json: %s", err)
			}
			w.WriteHeader(400)
			w.Write(data)
			return
		}
		log.Printf("Valid chirp")
		validation := validMsg {
			Valid: true,
		}
		data, err := json.Marshal(validation)
		if err != nil {
			log.Printf("Error marshalling json: %s", err)
		}
		w.WriteHeader(200)
		w.Write(data)
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