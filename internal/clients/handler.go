package clients

import (
	"encoding/json"
	"erp-constructora/internal/middlewares"
	"erp-constructora/internal/utils"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		utils.WriteUnauthorized(w)
		return
	}

	var req CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Payload inválido")
		return
	}

	client, err := h.service.CreateClient(r.Context(), companyID, &req)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(client)
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		utils.WriteUnauthorized(w)
		return
	}

	clients, err := h.service.GetClients(r.Context(), companyID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clients)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "ID del cliente es requerido")
		return
	}

	var req UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Payload inválido")
		return
	}

	client, err := h.service.UpdateClient(r.Context(), companyID, id, &req)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(client)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "ID del cliente es requerido")
		return
	}

	err := h.service.DeleteClient(r.Context(), companyID, id)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "recurso eliminado"})
}
