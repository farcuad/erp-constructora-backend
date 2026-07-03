package contractors

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

func (h *Handler) CreateContractor(w http.ResponseWriter, r *http.Request) {
	companyID, ok := users.GetCompanyIDFromContext(r.Context())
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
	companyID, ok := users.GetCompanyIDFromContext(r.Context())
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
	userID, ok := users.GetUserIDFromContext(r.Context())
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
