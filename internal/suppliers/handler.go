package suppliers

import (
	"encoding/json"
	"erp-constructora/internal/middlewares"
	"erp-constructora/internal/utils"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// --- CONTROLADORES DE PROVEEDORES ---

func (h *Handler) CreateSupplier(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok || companyID == "" {
		utils.WriteUnauthorized(w)
		return
	}

	var s Supplier
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		utils.WriteBadRequest(w, "JSON inválido: "+err.Error())
		return
	}

	// Inyectar el ID de la empresa autenticada
	s.CompanyID = companyID

	if err := h.service.CreateSupplier(r.Context(), &s); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(s)
}

func (h *Handler) GetAllSuppliers(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	suppliers, err := h.service.ListSuppliers(r.Context(), companyID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppliers)
}

func (h *Handler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "El parámetro id es obligatorio")
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var req UpdateSupplierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido: "+err.Error())
		return
	}

	sup, err := h.service.UpdateSupplier(r.Context(), id, companyID, &req)
	if err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sup)
}

func (h *Handler) DeleteSupplier(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "El parámetro id es obligatorio")
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	if err := h.service.DeleteSupplier(r.Context(), id, companyID); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "recurso eliminado"})
}
