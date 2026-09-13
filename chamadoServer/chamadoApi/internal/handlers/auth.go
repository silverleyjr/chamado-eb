package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"chamadoApi/api"
	"chamadoApi/internal/tools"

	"github.com/golang-jwt/jwt/v5"
	log "github.com/sirupsen/logrus"
)

func secretKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "dev-secret-change-me"
	}
	return []byte(secret)
}

func Login(w http.ResponseWriter, r *http.Request) {
	noAcessControl(w)
	name := decodedHeader(r, "Name")
	if name == "" {
		api.RequestErrorHandler(w, errors.New("Name is required"))
		return
	}
	password := r.Header.Get("Password")

	userData, err := tools.GetUserData(tools.Database, name)
	if err != nil {
		api.RequestErrorHandler(w, errors.New("Invalid name"))
		return
	}
	if userData.Password != password {
		api.RequestErrorHandler(w, errors.New("Invalid password"))
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"userName": userData.Name,
			"userRole": userData.Role,
			"team":     userData.Team,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey())
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	response := api.LoginResponse{
		Authorization: tokenString,
		Code:          http.StatusOK,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}
