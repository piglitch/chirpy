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

	auth "main.go/internal"
	"main.go/internal/databases"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	dbQueries *databases.Queries
	platform string
	jwtSecret string
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
		log.Print("r.Header: ", r.Header)
		cleanedTokenString, err := auth.GetBearerToken(r.Header)
		if err != nil {
			log.Printf("Error cleaning the bearer token: %s", err)
			w.WriteHeader(401)
			return
		}
		
		log.Print(cleanedTokenString, 89)
		userId, err := auth.ValidateJWT(cleanedTokenString, apiCfg.jwtSecret)
		if err != nil {
			log.Printf("Invalid token: %s. Line: %d", err, 92)
			w.WriteHeader(401)
			return
		}
		decoder := json.NewDecoder(r.Body)
		params := Chirp{}
		err = decoder.Decode(&params)
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
			UserID: userId,
		}

		newChirp, err := apiCfg.dbQueries.CreateChirp(r.Context(), validation)
		if err != nil {
			log.Printf("Error executing query: %s", err)
			w.WriteHeader(500)
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
			Hashed_Pass string `json:"password"`
		}
		type UserResp struct{
			ID uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Email string `json:"email"`
		}
		type errMsg struct{
			Body string `json:"body"`
		}
		decoder := json.NewDecoder(r.Body)
		params := User{}
		err := decoder.Decode(&params)
		if len(params.Hashed_Pass) == 0 {
			marshalledErr, err := json.Marshal(errMsg{Body: "Must include a password"})
			if err != nil {
				log.Printf("Error marshalling: %s", err)
				return
			}	
			w.WriteHeader(http.StatusBadRequest)
			w.Write(marshalledErr)
			return
		}
		if err != nil {
			log.Printf("Error decoding parameters: %s", err)
			w.WriteHeader(500)
			return
		}
		hashed_pass, err := auth.HashPassword(params.Hashed_Pass) 
		if err != nil {
			log.Printf("Error while hashing password: %s", err)
			return
		}
		userWpass := databases.CreateUserParams{
			Email: params.Email,
			HashedPassword: hashed_pass,
		}
		user, err := apiCfg.dbQueries.CreateUser(r.Context(), userWpass)
		if err != nil {
			log.Printf("Error executing query: %s", err)
			return
		}
		newUser := UserResp{
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

func userLogin(apiCfg *apiConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type User struct{
			Email string `json:"email"`
			Hashed_Pass string `json:"password"`
			Expiry int `json:"expiry_in_seconds"`
		}
		decoder := json.NewDecoder(r.Body)
		params := User{}
		err := decoder.Decode(&params)
		if err != nil {
			log.Printf("Error decodinng params: %s", err)
			return
		}
		dbUser, err := apiCfg.dbQueries.GetUserByEmail(r.Context(), params.Email)
		if err != nil {
			log.Printf("failed executing query: %s", err)
			return 
		}
		tryPassword := params.Hashed_Pass
		log.Print(287, params)
		dbPassword, err := apiCfg.dbQueries.GetPassword(r.Context(), params.Email)
		if err != nil {
			log.Printf("Error executing query: %s on Line 289", err)
			return
		}
		log.Print(300, dbPassword)
		err = auth.CheckPasswordHash(tryPassword, dbPassword)
		log.Print("err: ", err)
		if err != nil {
			w.WriteHeader(401)
			w.Write([]byte("Incorrect email or password"))
			return
		}
		def_expiry := 3600 
		if params.Expiry != 0 {
			def_expiry = params.Expiry 
		}
		tokenString, err := auth.MakeJWT(dbUser.ID, apiCfg.jwtSecret, time.Duration(def_expiry)*time.Second)
		if err != nil {
			log.Printf("Error genearting jwt: %s", err)
			return
		}
		log.Print("token string: ", tokenString)
		_, err = auth.ValidateJWT(tokenString, apiCfg.jwtSecret)
		if err != nil {
			log.Printf("Error validating token: %s. Line: %d", err, 347)
			return
		}
		type UserResp struct{
			ID uuid.UUID `json:"id"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
			Email string `json:"email"`
			Hashed_Pass string `json:"password"`
			Token string `json:"token"`
		}
		respUser := UserResp{
			ID: dbUser.ID,
			CreatedAt: dbUser.CreatedAt,
			UpdatedAt: dbUser.UpdatedAt,
			Email: params.Email,
			Token: tokenString,
		}
		marshalledResp, err := json.Marshal(respUser)
		if err != nil {
			log.Printf("Error marshalling response: %s", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(marshalledResp)
	}
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	jWTSecret := os.Getenv("JWT_SECRET") 
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
		jwtSecret: jWTSecret,
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
	mux.HandleFunc("POST /api/login", userLogin(apiCfg))

	server := &http.Server{
		Addr: ":8080",
		Handler: mux,
	}
	err = server.ListenAndServe() 
    if err != nil {
        fmt.Println("Error starting server:", err)
    }
}