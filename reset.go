package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		err := fmt.Errorf("reset attempted in non-dev environment")
		respondWithError(w, http.StatusForbidden, err.Error(), err)
		return
	}
	cfg.fileserverHits.Store(0)
	cfg.db.Reset(r.Context())
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0 and database reset to initial state"))
}
