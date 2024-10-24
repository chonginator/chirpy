package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	const port = "8080"
	const filepathRoot = "."
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
	}

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(filepathRoot))
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.middlewareMetricsInc(fileServer)))
	mux.HandleFunc("GET /api/healthz", handlerReadiness)
	mux.HandleFunc("POST /api/validate_chirp", handlerValidateChirp)
	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	srv := &http.Server{
		Addr: ":" + port,
		Handler: mux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(srv.ListenAndServe())
}


func handlerValidateChirp(w http.ResponseWriter, r *http.Request) {
	const maxChirpLength = 140

	type parameters struct {
		Body string `json:"body"`
	}
	params := parameters{}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		writeErrorResponse(w)
		return
	}

	if params.Body == "" {
		w.WriteHeader(400)
		writeErrorResponse(w)
		return
	}

	if len(params.Body) > maxChirpLength {
		log.Printf("Chirp exceeds %d characters", maxChirpLength)
		w.WriteHeader(400)
		writeErrorResponse(w)
		return
	}

	type successRespBody = struct {
		Valid bool `json:"valid"`
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	respBody, err := json.Marshal(successRespBody{ Valid: true })
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}
	w.Write(respBody)
}

func writeErrorResponse(w http.ResponseWriter) {
	type errRespBody struct {
		Error string 
	}

	w.Header().Set("Content-Type", "application/json")
	respBody, err := json.Marshal(errRespBody{ Error: "Something went wrong" })
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		return
	}
	w.Write(respBody)
}