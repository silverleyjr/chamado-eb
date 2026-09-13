package api

import (
	"chamadoApi/internal/tools"
	"encoding/json"
	"net/http"
)

// DATA RESPONSES
type LoginResponse struct {
	Code          int
	Authorization string
}
type GetChamadoResponse struct {
	Number      int
	Name        string
	Section     string
	Description string
	Response    string
	Status      string
	CreatedAt   string
	Code        int
}
type GetChamadoBySectionResponse struct {
	Chamados []tools.ChamadoData
	Code     int
}
type GetUserResponse struct {
	Name        string
	Team        string
	Role        string
	TimeCreated string
	Code        int
}

// SIMPLE RESPONSES
type CodeResponse struct {
	Code int
}

// ERROR RESPONSES
type Error struct {
	Code    int
	Message string
}

func writeError(w http.ResponseWriter, message string, code int) {
	resp := Error{
		Code:    code,
		Message: message,
	}
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(resp)
}

var (
	RequestErrorHandler = func(w http.ResponseWriter, err error) {
		writeError(w, err.Error(), http.StatusBadRequest)
	}
	InternalErrorHandler = func(w http.ResponseWriter) {
		writeError(w, "An Unexpected Error Occured.", http.StatusInternalServerError)
	}
	UnauthorizedErrorHandler = func(w http.ResponseWriter, err error) {
		writeError(w, err.Error(), http.StatusForbidden)
	}
)
