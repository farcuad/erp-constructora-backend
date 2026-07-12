package contractors

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

func (h *Handler) CreateContractor(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var c Contractor
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		http.Error(w, "JSON corrupto", http.StatusBadRequest)
		return
	}
	c.CompanyID = companyID

	if err := h.service.CreateContractor(r.Context(), &c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var cc ContractorContract
	if err := json.NewDecoder(r.Body).Decode(&cc); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	cc.CompanyID = companyID

	if err := h.service.CreateContract(r.Context(), &cc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cc)
}

func (h *Handler) PostPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var p ContractorPayment
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	p.UserID = userID

	if err := h.service.AddPayment(r.Context(), &p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *Handler) GetAllContractPayments(w http.ResponseWriter, r *http.Request) {
	suppliers, err := h.service.GetAllContractPayment(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppliers)
}

func (h *Handler) GetALlContracts(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	suppliers, err := h.service.GetALlContract(r.Context(), companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppliers)
}

func (h *Handler) GetContracts(w http.ResponseWriter, r *http.Request) {
	// Aplicando tu estilo de ruta limpia /{project_id}
	projectID := r.PathValue("project_id")
	if projectID == "" {
		http.Error(w, "Falta el parámetro project_id en la ruta", http.StatusBadRequest)
		return
	}

	list, err := h.service.ListContractsByProject(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) UpdateContractor(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Falta el id del contratista", http.StatusBadRequest)
		return
	}

	var req UpdateContractorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	c := Contractor{
		ID:             id,
		CompanyID:      companyID,
		Name:           req.Name,
		NIT:            req.NIT,
		Representative: req.Representative,
		Phone:          req.Phone,
		Email:          req.Email,
		IsActive:       req.IsActive != nil && *req.IsActive,
	}

	if err := h.service.UpdateContractor(r.Context(), &c); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *Handler) DeleteContractor(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Falta el id del contratista", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteContractor(r.Context(), companyID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateContractorContract(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Falta el id del contrato", http.StatusBadRequest)
		return
	}

	var req UpdateContractorContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	cc := ContractorContract{
		ID:          id,
		CompanyID:   companyID,
		Title:       req.Title,
		TotalAmount: req.TotalAmount,
		Balance:     req.Balance,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Status:      req.Status,
	}

	if err := h.service.UpdateContract(r.Context(), &cc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cc)
}

func (h *Handler) DeleteContractorContract(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Falta el id del contrato", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteContract(r.Context(), companyID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
