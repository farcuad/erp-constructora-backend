package contractors

import (
	"context"
	"encoding/json"
	"erp-constructora/internal/middlewares"
	"erp-constructora/internal/notifications"
	"erp-constructora/internal/utils"
	"fmt"
	"log"
	"net/http"
)

func strPtr(s string) *string {
	return &s
}

type Handler struct {
	service  *Service
	notifier notifications.Notifier
}

func NewHandler(service *Service, notifier notifications.Notifier) *Handler {
	return &Handler{service: service, notifier: notifier}
}

// notify emite una notificación a toda la empresa (el actor queda excluido por NotifyFromContext)
func (h *Handler) notify(ctx context.Context, req notifications.CreateNotificationRequest) {
	if h.notifier == nil {
		return
	}
	if err := h.notifier.NotifyFromContext(ctx, req); err != nil {
		log.Printf("[NOTIFY ERROR] contractors: %v", err)
	}
}

func (h *Handler) CreateContractor(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var c Contractor
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		utils.WriteBadRequest(w, "JSON corrupto")
		return
	}
	c.CompanyID = companyID

	if err := h.service.CreateContractor(r.Context(), &c); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "CONTRACTOR",
		EntityID:   &c.ID,
		Type:       "CONTRACTOR_CREATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Nuevo contratista",
		Message:    fmt.Sprintf("Se registró el contratista: %s.", c.Name),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func (h *Handler) CreateContract(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var cc ContractorContract
	if err := json.NewDecoder(r.Body).Decode(&cc); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}
	cc.CompanyID = companyID

	if err := h.service.CreateContract(r.Context(), &cc); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		ProjectID:  &cc.ProjectID,
		EntityType: "CONTRACTOR_CONTRACT",
		EntityID:   &cc.ID,
		Type:       "CONTRACTOR_CONTRACT_CREATED",
		Priority:   notifications.PriorityHigh,
		Title:      "Nuevo contrato de contratista",
		Message:    fmt.Sprintf("Se creó un contrato de contratista por $%.2f.", cc.TotalAmount),
		LinkToUI:   strPtr("/dashboard/projects/" + cc.ProjectID + "/contractors"),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cc)
}

func (h *Handler) PostPayment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var p ContractorPayment
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}
	p.UserID = userID

	if err := h.service.AddPayment(r.Context(), &p); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "CONTRACTOR_PAYMENT",
		EntityID:   &p.ID,
		Type:       "CONTRACTOR_PAYMENT_CREATED",
		Priority:   notifications.PriorityHigh,
		Title:      "Nuevo pago a contratista",
		Message:    fmt.Sprintf("Se registró un pago a contratista de $%.2f.", p.Amount),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *Handler) GetAllContractPayments(w http.ResponseWriter, r *http.Request) {
	suppliers, err := h.service.GetAllContractPayment(r.Context())
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppliers)
}

func (h *Handler) GetALlContracts(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	suppliers, err := h.service.GetALlContract(r.Context(), companyID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppliers)
}

func (h *Handler) GetContracts(w http.ResponseWriter, r *http.Request) {
	// Aplicando tu estilo de ruta limpia /{project_id}
	projectID := r.PathValue("project_id")
	if projectID == "" {
		utils.WriteBadRequest(w, "Falta el parámetro project_id en la ruta")
		return
	}

	list, err := h.service.ListContractsByProject(r.Context(), projectID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) UpdateContractor(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id del contratista")
		return
	}

	var req UpdateContractorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
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
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "CONTRACTOR",
		EntityID:   &id,
		Type:       "CONTRACTOR_UPDATED",
		Priority:   notifications.PriorityLow,
		Title:      "Contratista actualizado",
		Message:    fmt.Sprintf("Se actualizó el contratista: %s.", c.Name),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(c)
}

func (h *Handler) DeleteContractor(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id del contratista")
		return
	}

	if err := h.service.DeleteContractor(r.Context(), companyID, id); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "CONTRACTOR",
		EntityID:   &id,
		Type:       "CONTRACTOR_DELETED",
		Priority:   notifications.PriorityLow,
		Title:      "Contratista eliminado",
		Message:    "Se eliminó un contratista.",
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) UpdateContractorContract(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id del contrato")
		return
	}

	var req UpdateContractorContractRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
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
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		ProjectID:  &cc.ProjectID,
		EntityType: "CONTRACTOR_CONTRACT",
		EntityID:   &id,
		Type:       "CONTRACTOR_CONTRACT_UPDATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Contrato de contratista actualizado",
		Message:    "Se actualizó un contrato de contratista.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cc)
}

func (h *Handler) DeleteContractorContract(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id del contrato")
		return
	}

	if err := h.service.DeleteContract(r.Context(), companyID, id); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "CONTRACTOR_CONTRACT",
		EntityID:   &id,
		Type:       "CONTRACTOR_CONTRACT_DELETED",
		Priority:   notifications.PriorityLow,
		Title:      "Contrato de contratista eliminado",
		Message:    "Se eliminó un contrato de contratista.",
	})

	w.WriteHeader(http.StatusNoContent)
}
