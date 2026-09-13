package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"chamadoApi/api"
	"chamadoApi/internal/tools"

	log "github.com/sirupsen/logrus"
)

func GetChamado(w http.ResponseWriter, r *http.Request) {
	noAcessControl(w)
	role, team, ok := getRoleAndTeam(r)
	if !ok {
		api.InternalErrorHandler(w)
		return
	}

	mode := r.Header.Get("Mode")
	var response any

	if mode == "all" {
		status := r.Header.Get("Status")
		chamados, err := tools.GetAllChamados(tools.Database, status)
		if err != nil {
			api.InternalErrorHandler(w)
			return
		}
		response = api.GetChamadoBySectionResponse{
			Chamados: chamados,
			Code:     http.StatusOK,
		}
	} else if mode == "bySection" {
		section := decodedHeader(r, "Section")
		if section == "" {
			section = team
		}
		if !isAdmin(role) && section != team {
			api.UnauthorizedErrorHandler(w, errors.New("Cannot view chamados from another section"))
			return
		}
		status := r.Header.Get("Status")
		chamados, err := tools.GetChamadoBySection(tools.Database, section, status)
		if err != nil {
			api.RequestErrorHandler(w, errors.New("Invalid section"))
			return
		}
		response = api.GetChamadoBySectionResponse{
			Chamados: chamados,
			Code:     http.StatusOK,
		}
	} else {
		number, err := strconv.Atoi(r.Header.Get("Number"))
		if err != nil {
			api.RequestErrorHandler(w, errors.New("Chamado number invalid"))
			return
		}
		chamado, err := tools.GetChamadoByNumber(tools.Database, number)
		if err != nil {
			api.RequestErrorHandler(w, errors.New("Chamado not found"))
			return
		}
		if !isAdmin(role) && chamado.Section != team {
			api.UnauthorizedErrorHandler(w, errors.New("Cannot view a chamado from another section"))
			return
		}
		response = api.GetChamadoResponse{
			Number:      chamado.Number,
			Name:        chamado.Name,
			Section:     chamado.Section,
			Description: chamado.Description,
			Response:    chamado.Response,
			Status:      chamado.Status,
			CreatedAt:   chamado.CreatedAt,
			Code:        http.StatusOK,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}

func GetChamadoPublic(w http.ResponseWriter, r *http.Request) {
	noAcessControl(w)

	section := decodedHeader(r, "Section")
	if section == "" {
		api.RequestErrorHandler(w, errors.New("Section is required"))
		return
	}
	status := r.Header.Get("Status")

	chamados, err := tools.GetChamadoBySection(tools.Database, section, status)
	if err != nil {
		api.RequestErrorHandler(w, errors.New("Invalid section"))
		return
	}
	response := api.GetChamadoBySectionResponse{
		Chamados: chamados,
		Code:     http.StatusOK,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}

func PostChamado(w http.ResponseWriter, r *http.Request) {
	noAcessControl(w)

	name := decodedHeader(r, "Name")
	section := decodedHeader(r, "Section")
	description := decodedHeader(r, "Description")

	if name == "" || section == "" || description == "" {
		api.RequestErrorHandler(w, errors.New("Invalid chamado"))
		return
	}

	created, err := tools.CreateChamado(tools.Database, tools.ChamadoData{
		Name:        name,
		Section:     section,
		Description: description,
	})
	if err != nil {
		api.InternalErrorHandler(w)
		return
	}

	response := api.GetChamadoResponse{
		Number:      created.Number,
		Name:        created.Name,
		Section:     created.Section,
		Description: created.Description,
		Response:    created.Response,
		Status:      created.Status,
		CreatedAt:   created.CreatedAt,
		Code:        http.StatusOK,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}

func PutChamado(w http.ResponseWriter, r *http.Request) {
	noAcessControl(w)
	role, team, ok := getRoleAndTeam(r)
	if !ok {
		api.InternalErrorHandler(w)
		return
	}

	number, err := strconv.Atoi(r.Header.Get("Number"))
	if err != nil {
		api.RequestErrorHandler(w, errors.New("Chamado number invalid"))
		return
	}
	status := r.Header.Get("Status")
	responseText := decodedHeader(r, "Response")
	if status == "" {
		api.RequestErrorHandler(w, errors.New("Status is required"))
		return
	}

	existing, err := tools.GetChamadoByNumber(tools.Database, number)
	if err != nil {
		api.RequestErrorHandler(w, errors.New("Chamado not found"))
		return
	}
	if !isAdmin(role) && existing.Section != team {
		api.UnauthorizedErrorHandler(w, errors.New("Cannot update a chamado from another section"))
		return
	}
	if responseText == "" {
		responseText = existing.Response
	}

	updated, err := tools.UpdateChamado(tools.Database, number, status, responseText)
	if err != nil {
		api.InternalErrorHandler(w)
		return
	}

	response := api.GetChamadoResponse{
		Number:      updated.Number,
		Name:        updated.Name,
		Section:     updated.Section,
		Description: updated.Description,
		Response:    updated.Response,
		Status:      updated.Status,
		CreatedAt:   updated.CreatedAt,
		Code:        http.StatusOK,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}

func DeleteChamado(w http.ResponseWriter, r *http.Request) {
	noAcessControl(w)
	role, team, ok := getRoleAndTeam(r)
	if !ok {
		api.InternalErrorHandler(w)
		return
	}

	number, err := strconv.Atoi(r.Header.Get("Number"))
	if err != nil {
		api.RequestErrorHandler(w, errors.New("Chamado number invalid"))
		return
	}

	if !isAdmin(role) {
		existing, err := tools.GetChamadoByNumber(tools.Database, number)
		if err != nil {
			api.RequestErrorHandler(w, errors.New("Chamado not found"))
			return
		}
		if existing.Section != team {
			api.UnauthorizedErrorHandler(w, errors.New("Cannot delete a chamado from another section"))
			return
		}
	}

	if err := tools.DeleteChamado(tools.Database, number); err != nil {
		api.RequestErrorHandler(w, errors.New("Chamado not found"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := api.CodeResponse{Code: http.StatusOK}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
	}
}
