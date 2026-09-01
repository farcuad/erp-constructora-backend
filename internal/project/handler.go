package project

import (
	"encoding/json"
	"erp-constructora/internal/middlewares"
	"erp-constructora/internal/utils"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteMethodNotAllowed(w)
		return
	}

	// Extraer el company_id de forma segura gracias al middleware
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var dto CreateProjectDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	if dto.Name == "" {
		utils.WriteBadRequest(w, "El nombre del proyecto es obligatorio")
		return
	}

	project, err := h.service.CreateProject(r.Context(), companyID, dto)
	if err != nil {
		if err == middlewares.ErrProjectLimitExceeded {
			utils.WriteError(w, http.StatusPaymentRequired, err.Error())
			return
		}
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(project)
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		utils.WriteMethodNotAllowed(w)
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	projects, err := h.service.ListProjects(r.Context(), companyID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		utils.WriteMethodNotAllowed(w)
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "ID del proyecto es requerido")
		return
	}

	var dto UpdateProjectDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	project, err := h.service.UpdateProject(r.Context(), companyID, id, dto)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(project)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteMethodNotAllowed(w)
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "ID del proyecto es requerido")
		return
	}

	err := h.service.DeleteProject(r.Context(), companyID, id)
	if err != nil {
		if strings.Contains(err.Error(), "no se puede eliminar") {
			utils.WriteConflict(w, err.Error())
			return
		}
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "recurso eliminado"})
}
