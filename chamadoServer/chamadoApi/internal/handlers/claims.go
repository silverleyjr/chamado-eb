package handlers

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
)

func getClaims(r *http.Request) (jwt.MapClaims, bool) {
	claims, ok := r.Context().Value("claims").(jwt.MapClaims)
	return claims, ok
}

func getRoleAndTeam(r *http.Request) (role string, team string, ok bool) {
	claims, present := getClaims(r)
	if !present {
		return "", "", false
	}
	role, roleOk := claims["userRole"].(string)
	team, teamOk := claims["team"].(string)
	return role, team, roleOk && teamOk
}

func isAdmin(role string) bool {
	return role == "admin"
}
