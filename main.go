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
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"main.go/internal/databases"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	dbQueries *databases.Queries
	platform string
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler{
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func displayServerHits(apiCfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(fmt.Sprintf("<html><body> <h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>",apiCfg.fileServerHits.Load())))
	}
}

func resetDB(apiCfg *apiConfig) http.HandlerFunc {
	type respMsg struct {
		Status string `json:"status"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if apiCfg.platform != "dev" {
				w.WriteHeader(http.StatusForbidden)
				return
		}
		_, err := apiCfg.dbQueries.DeleteUser(r.Context())
		if err != nil {
				log.Printf("Error executing query: %s", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
		}
		responseMsg := respMsg{
			Status: "reset done",
		}
		marshalledMsg, err := json.Marshal(responseMsg)
		if err != nil {
			log.Printf("Error while marshalling: %s", err)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(marshalledMsg)
	}
}

func createChirp(apiCfg *apiConfig) http.HandlerFunc {
	type Chirp struct {
		Id        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserId    uuid.UUID `json:"user_id"`
	}

	type errStruct struct {
		Error string `json:"error"`
	}
	return func(w http.ResponseWriter, r *http.Request) {

		decoder := json.NewDecoder(r.Body)
		params := Chirp{}
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
		
		validation := databases.CreateChirpParams{
			Body: strings.Join(cleanedArr, " "),
			UserID: params.UserId,
		}

		newChirp, err := apiCfg.dbQueries.CreateChirp(r.Context(), validation)
		if err != nil {
			log.Printf("Error executing query: %s", err)
			return
		}
		responseChirp := Chirp{
			Id: newChirp.ID,
			CreatedAt: newChirp.CreatedAt,
			UpdatedAt: newChirp.UpdatedAt,
			Body: newChirp.Body,
			UserId: newChirp.UserID,
		}

		marshalledChirp, err := json.Marshal(responseChirp)
		if err != nil {
			log.Printf("Error marshalling json: %s", err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write(marshalledChirp)
	}
}

func healthRoute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func createUser(apiCfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type User struct{
			ID uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Email string `json:"email"`
		}

		decoder := json.NewDecoder(r.Body)
		params := User{}
		err := decoder.Decode(&params)

		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(500)
			return
		}
		user, err := apiCfg.dbQueries.CreateUser(r.Context(), params.Email)
		if err != nil {
			log.Printf("Error executing query: %s", err)
			return
		}
		newUser := User{
			ID: user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Email: user.Email,
		}
		marshalledNewUser, err := json.Marshal(newUser)
		if err != nil {
			log.Printf("error in marshalling: %s", err)
			return
		}
		w.WriteHeader(http.StatusCreated)
		w.Write(marshalledNewUser)
	}	
}

func getAllChirps(apiCfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Chirp struct {
			Id        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Body      string    `json:"body"`
			UserId    uuid.UUID `json:"user_id"`
		}

		chirps, err := apiCfg.dbQueries.GetAllChirps(r.Context())
		if err != nil {
			log.Printf("Error while executing sql query: %s", err)
			return
		}
		chirpsResponse := []Chirp{}
		for _, chirp := range chirps {
			chirpResponse := Chirp{
				Id: chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body: chirp.Body,
				UserId: chirp.UserID,
			}
			chirpsResponse = append(chirpsResponse, chirpResponse)
		}
		marshalledChirps, err := json.Marshal(chirpsResponse)
		if err != nil {
			log.Printf("Error marshalling chirps: %s", err)
			return 
		}	
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(marshalledChirps)
	}
}

func getChirpById(apiCfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type Chirp struct {
			Id        uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Body      string    `json:"body"`
			UserId    uuid.UUID `json:"user_id"`
		}

		chirpId, err := uuid.Parse(r.PathValue("chirpID")) 
		if err != nil {
			log.Printf("Unable to parse chirpId: %s", err)
			return
		}
		chirp, err := apiCfg.dbQueries.GetChirpById(r.Context(), chirpId)
		if err != nil {
			log.Printf("failed to execute sql query: %s", err)
		}
		chirpResp := Chirp{
			Id: chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body: chirp.Body,
			UserId: chirp.UserID,
		}
		marshalledResp, err := json.Marshal(chirpResp)
		if err != nil {
			log.Printf("Failed to unmarshal: %s", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write(marshalledResp)
	}
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("unable to connect to the database: ", err)
	}
	dbQueries := databases.New(db)
	platform := os.Getenv("PLATFORM")
	mux := http.NewServeMux()
	apiCfg := &apiConfig{
		dbQueries: dbQueries,
		platform: platform,
	}
	rootHandler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(rootHandler))
	mux.Handle("/assets/", apiCfg.middlewareMetricsInc(http.FileServer(http.Dir("./assets/"))))

	mux.HandleFunc("GET /api/healthz", healthRoute)
	mux.HandleFunc("GET /admin/metrics", displayServerHits(apiCfg))
	mux.HandleFunc("POST /admin/reset", resetDB(apiCfg))
	mux.HandleFunc("POST /api/users", createUser(apiCfg))
	mux.HandleFunc("POST /api/chirps", createChirp(apiCfg))
	mux.HandleFunc("GET /api/chirps", getAllChirps(apiCfg))
	mux.HandleFunc("GET /api/chirps/{chirpID}", getChirpById(apiCfg))

	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	err = server.ListenAndServe() 
    if err != nil {
        fmt.Println("Error starting server:", err)
    }
}