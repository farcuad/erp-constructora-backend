package inventory

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

func (h *Handler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	companyID, ok := users.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var wh Warehouse
	if err := json.NewDecoder(r.Body).Decode(&wh); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	wh.CompanyID = companyID

	if err := h.service.CreateWarehouse(r.Context(), &wh); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wh)
}

func (h *Handler) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	companyID, ok := users.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var m Material
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	m.CompanyID = companyID

	if err := h.service.CreateMaterial(r.Context(), &m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func (h *Handler) PostMovement(w http.ResponseWriter, r *http.Request) {
	userID, ok := users.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var m StockMovement
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	m.UserID = userID

	if err := h.service.RegisterMovement(r.Context(), &m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func (h *Handler) GetStock(w http.ResponseWriter, r *http.Request) {
	warehouseID := r.PathValue("warehouse_id")
	if warehouseID == "" {
		http.Error(w, "Falta warehouse_id", http.StatusBadRequest)
		return
	}

	stock, err := h.service.GetCurrentStock(r.Context(), warehouseID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stock)
}

func (h *Handler) UpdateMaterial(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "El parámetro id es obligatorio", http.StatusBadRequest)
		return
	}

	companyID, ok := users.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req UpdateMaterialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	m, err := h.service.UpdateMaterial(r.Context(), id, companyID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func (h *Handler) DeleteMaterial(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "El parámetro id es obligatorio", http.StatusBadRequest)
		return
	}

	companyID, ok := users.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	if err := h.service.DeleteMaterial(r.Context(), id, companyID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "recurso eliminado"})
}

func (h *Handler) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "El parámetro id es obligatorio", http.StatusBadRequest)
		return
	}

	companyID, ok := users.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req UpdateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	wh, err := h.service.UpdateWarehouse(r.Context(), id, companyID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(wh)
}

func (h *Handler) DeleteWarehouse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "El parámetro id es obligatorio", http.StatusBadRequest)
		return
	}

	companyID, ok := users.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	if err := h.service.DeleteWarehouse(r.Context(), id, companyID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "recurso eliminado"})
}
