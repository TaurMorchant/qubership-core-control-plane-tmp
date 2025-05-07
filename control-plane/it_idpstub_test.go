package main

import (
	"encoding/json"
	"net/http"
)

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	RespondWithJson(w, code, map[string]string{"error": msg})
}

func RespondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	response := []byte("")
	if payload != nil {
		response, _ = json.MarshalIndent(payload, "", "  ")
		w.Header().Set("Content-Type", "application/json")
	}
	w.WriteHeader(code)
	_, _ = w.Write(response)
}

func ResponseOk(w http.ResponseWriter, payload interface{}) {
	RespondWithJson(w, http.StatusOK, payload)
}
