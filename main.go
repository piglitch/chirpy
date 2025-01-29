package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"main.go/internal/databases"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	dbQueries *databases.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func displayServerHits(w http.ResponseWriter, r *http.Request) {
	apiCfg := &apiConfig{}
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(fmt.Sprintf("<html><body> <h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>",apiCfg.fileServerHits.Load())))
}

func resetServerHitCount(w http.ResponseWriter, r *http.Request) {
	apiCfg := &apiConfig{}
	apiCfg.fileServerHits.Swap(0)
	w.Write([]byte(fmt.Sprintf("Hits: %d", 0)))
}

func validateChirp(w http.ResponseWriter, r *http.Request) {

	type chirp struct {
		Body string `json:"body"`
	}
	type profaneChirp struct {
		Cleaned_Body string `json:"cleaned_body"`
	}
	type errStruct struct {
		Error string `json:"error"`
	}
	// type validMsg struct {
	// 	Valid bool `json:"valid"`
	// }
	decoder := json.NewDecoder(r.Body)
	params := chirp{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}
	msg := params.Body
	if len(msg) > 140 {
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
	cleanedArr := strings.Split(msg, " ")
	
	for i := 0; i < len(cleanedArr); i++ {
		if strings.ToLower(cleanedArr[i]) == "kerfuffle" || strings.ToLower(cleanedArr[i]) == "sharbert" || strings.ToLower(cleanedArr[i]) == "fornax" {
			cleanedArr[i] = "****"
		}
	}
	
	validation := profaneChirp {
		Cleaned_Body: strings.Join(cleanedArr, " "),
	}
	validdata, err := json.Marshal(validation)
	if err != nil {
		log.Printf("Error marshalling json: %s", err)
	}
	w.WriteHeader(200)
	w.Write(validdata)
}

func healthRoute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("unable to connect to the database: ", err)
	}
	dbQueries := databases.New(db)
	mux := http.NewServeMux()
	apiCfg := &apiConfig{dbQueries: dbQueries}
	rootHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(rootHandler))
	mux.Handle("/assets/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("./assets/"))))

	mux.HandleFunc("GET /api/healthz", healthRoute)
	mux.HandleFunc("GET /admin/metrics", displayServerHits)
	mux.HandleFunc("POST /admin/reset", resetServerHitCount)
	mux.HandleFunc("POST /api/validate_chirp", validateChirp)

	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	err = server.ListenAndServe() 
    if err != nil {
        fmt.Println("Error starting server:", err)
    }
}