package personnel

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

func (h *Handler) CreatePosition(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var p Position
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}
	p.CompanyID = companyID

	if err := h.service.CreatePosition(r.Context(), &p); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var e Employee
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}
	e.CompanyID = companyID

	if err := h.service.RegisterEmployee(r.Context(), &e); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(e)
}

func (h *Handler) GetEmployees(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	list, err := h.service.ListEmployees(r.Context(), companyID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	var c Contract
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	if err := h.service.AddContract(r.Context(), &c); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func (h *Handler) GetAllPositions(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	list, err := h.service.GetPositions(r.Context(), companyID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) UpdatePosition(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id del cargo")
		return
	}

	var req UpdatePositionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	p := Position{
		ID:         id,
		CompanyID:  companyID,
		Name:       req.Name,
		BaseSalary: req.BaseSalary,
	}

	if err := h.service.UpdatePosition(r.Context(), &p); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func (h *Handler) DeletePosition(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id del cargo")
		return
	}

	if err := h.service.DeletePosition(r.Context(), companyID, id); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id del empleado")
		return
	}

	var req UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	e := Employee{
		ID:         id,
		CompanyID:  companyID,
		PositionID: req.PositionID,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		DNI:        req.DNI,
		Phone:      req.Phone,
		Email:      req.Email,
		ImgURL:     req.ImgURL,
		Status:     req.Status,
	}

	if err := h.service.UpdateEmployee(r.Context(), &e); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(e)
}

func (h *Handler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id del empleado")
		return
	}

	if err := h.service.DeleteEmployee(r.Context(), companyID, id); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetALlContracts(w http.ResponseWriter, r *http.Request) {

	projectID := r.PathValue("project_id")
	if projectID == "" {
		utils.WriteBadRequest(w, "Falta el project_id")
		return
	}

	list, err := h.service.GetAllContract(r.Context(), projectID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) UpdateContract(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id del contrato")
		return
	}

	var req UpdateContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	c := Contract{
		ID:           id,
		ContractType: req.ContractType,
		Salary:       req.Salary,
		StartDate:    req.StartDate,
		EndDate:      req.EndDate,
		Status:       req.Status,
	}

	if err := h.service.UpdateContract(r.Context(), &c); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *Handler) DeleteContract(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id del contrato")
		return
	}

	if err := h.service.DeleteContract(r.Context(), id); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
