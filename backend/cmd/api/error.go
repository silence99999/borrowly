package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) WriteError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(map[string]string{
		"error": err.Error(),
	})
}
