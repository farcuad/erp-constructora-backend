package rag

import (
	"encoding/json"
	"net/http"

	"erp-constructora/internal/middlewares"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) HandleChat(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Formato JSON inválido", http.StatusBadRequest)
		return
	}

	if req.Question == "" {
		http.Error(w, "La pregunta no puede estar vacía", http.StatusBadRequest)
		return
	}

	response, err := h.service.Ask(r.Context(), companyID, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
