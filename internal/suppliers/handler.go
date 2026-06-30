package suppliers

import (
	"encoding/json"
	"net/http"

	"erp-constructora/internal/users"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// --- CONTROLADORES DE PROVEEDORES ---

func (h *Handler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	companyID, ok := users.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		http.Error(w, "No autorizado: ID de empresa ausente", http.StatusUnauthorized)
		return
	}

	var s Supplier
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Inyectar el ID de la empresa autenticada
	s.CompanyID = companyID

	if err := h.service.CreateSupplier(r.Context(), &s); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

func (h *Handler) GetAllSuppliers(w http.ResponseWriter, r *http.Request) {
	companyID, ok := users.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No se encontró la constructora en el contexto de autenticación", http.StatusUnauthorized)
		return
	}

	suppliers, err := h.service.ListSuppliers(r.Context(), companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppliers)
}
