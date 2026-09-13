package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"chamadoApi/api"
	"chamadoApi/internal/tools"

	log "github.com/sirupsen/logrus"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	noAcessControl(w)
	name := decodedHeader(r, "Name")
	if name == "" {
		api.RequestErrorHandler(w, errors.New("Name is required"))
		return
	}
	userData, err := tools.GetUserData(tools.Database, name)
	if err != nil {
		api.RequestErrorHandler(w, errors.New("Invalid name"))
		return
	}
	response := api.GetUserResponse{
		Name:        userData.Name,
		Team:        userData.Team,
		Role:        userData.Role,
		TimeCreated: userData.TimeCreated,
		Code:        http.StatusOK,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}

func PostUser(w http.ResponseWriter, r *http.Request) {
	noAcessControl(w)
	role, _, ok := getRoleAndTeam(r)
	if !ok || !isAdmin(role) {
		api.UnauthorizedErrorHandler(w, errors.New("Only admins can create users"))
		return
	}

	name := decodedHeader(r, "Name")
	password := r.Header.Get("Password")
	team := decodedHeader(r, "Team")
	userRole := r.Header.Get("Role")
	if userRole == "" {
		userRole = "technician"
	}

	if name == "" || password == "" || team == "" {
		api.RequestErrorHandler(w, errors.New("Invalid user"))
		return
	}

	newUser := tools.UserData{
		Name:     name,
		Password: password,
		Team:     team,
		Role:     userRole,
	}

	if _, err := tools.CreateUser(tools.Database, newUser); err != nil {
		api.InternalErrorHandler(w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response := api.CodeResponse{Code: http.StatusOK}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	noAcessControl(w)
	role, _, ok := getRoleAndTeam(r)
	if !ok || !isAdmin(role) {
		api.UnauthorizedErrorHandler(w, errors.New("Only admins can delete users"))
		return
	}

	name := decodedHeader(r, "Name")
	if name == "" {
		api.RequestErrorHandler(w, errors.New("Name is required"))
		return
	}
	if err := tools.DeleteUser(tools.Database, name); err != nil {
		api.RequestErrorHandler(w, errors.New("User not found"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	response := api.CodeResponse{Code: http.StatusOK}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}
