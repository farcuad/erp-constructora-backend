package equipement

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

func (h *Handler) CreateEquipment(w http.ResponseWriter, r *http.Request) {
	companyID, ok := users.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Acceso no autorizado", http.StatusUnauthorized)
		return
	}

	var eq Equipment
	if err := json.NewDecoder(r.Body).Decode(&eq); err != nil {
		http.Error(w, "JSON corrupto o inválido", http.StatusBadRequest)
		return
	}
	eq.CompanyID = companyID

	if err := h.service.RegisterEquipment(r.Context(), &eq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(eq)
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	companyID, ok := users.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Acceso no autorizado", http.StatusUnauthorized)
		return
	}

	list, err := h.service.ListEquipment(r.Context(), companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) Assign(w http.ResponseWriter, r *http.Request) {
	userID, ok := users.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Acceso no autorizado", http.StatusUnauthorized)
		return
	}

	var assign EquipmentAssignment
	if err := json.NewDecoder(r.Body).Decode(&assign); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	assign.AssignedBy = userID

	if err := h.service.AssignEquipment(r.Context(), &assign); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(assign)
}

func (h *Handler) Maintenance(w http.ResponseWriter, r *http.Request) {
	var m MaintenanceRecord
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.service.RegisterMaintenance(r.Context(), &m); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}
