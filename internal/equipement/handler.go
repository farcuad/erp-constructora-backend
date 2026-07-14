package equipement

import (
	"encoding/json"
	"net/http"

	"erp-constructora/internal/middlewares"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateEquipment(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
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
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
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
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
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

func (h *Handler) GetAssignment(w http.ResponseWriter, r *http.Request) {

	equipmentID := r.PathValue("equipment_id")
	if equipmentID == "" {
		http.Error(w, "Falta equipment_id", http.StatusBadRequest)
		return
	}

	list, err := h.service.GetEquipementassignments(r.Context(), equipmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
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

func (h *Handler) GetMaintenanceById(w http.ResponseWriter, r *http.Request) {

	equipmentID := r.PathValue("equipment_id")
	if equipmentID == "" {
		http.Error(w, "Falta equipment_id en los parametros", http.StatusBadRequest)
		return
	}

	list, err := h.service.GetMaintenance(r.Context(), equipmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) UpdateEquipment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "El parámetro id es obligatorio", http.StatusBadRequest)
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req UpdateEquipmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	e, err := h.service.UpdateEquipment(r.Context(), id, companyID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func (h *Handler) DeleteEquipment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "El parámetro id es obligatorio", http.StatusBadRequest)
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	if err := h.service.DeleteEquipment(r.Context(), id, companyID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "recurso eliminado"})
}

func (h *Handler) GetAllEquipmentTypes(w http.ResponseWriter, r *http.Request) {

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Acceso no autorizado", http.StatusUnauthorized)
		return
	}

	list, err := h.service.GetEquipmentTypes(r.Context(), companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) CreateEquipmentType(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Acceso no autorizado", http.StatusUnauthorized)
		return
	}

	var eq EquipmentType
	if err := json.NewDecoder(r.Body).Decode(&eq); err != nil {
		http.Error(w, "JSON corrupto o inválido", http.StatusBadRequest)
		return
	}
	eq.CompanyID = companyID

	if err := h.service.CreateEquipmentTypes(r.Context(), &eq); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(eq)
}

func (h *Handler) UpdateEquipmentType(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "El parámetro id es obligatorio", http.StatusBadRequest)
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req UpdateEquipmentTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	et, err := h.service.UpdateEquipmentType(r.Context(), id, companyID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(et)
}

func (h *Handler) DeleteEquipmentType(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "El parámetro id es obligatorio", http.StatusBadRequest)
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	if err := h.service.DeleteEquipmentType(r.Context(), id, companyID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "recurso eliminado"})
}
