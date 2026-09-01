package subscriptions

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

func (h *Handler) GetMySubscription(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	sub, err := h.service.GetMySubscription(r.Context(), companyID)
	if err != nil {
		utils.WriteNotFound(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}

func (h *Handler) GetAllSubscriptions(w http.ResponseWriter, r *http.Request) {
	subs, err := h.service.GetAllWithCompany(r.Context())
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(subs)
}

func (h *Handler) GetSubscriptionByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id de la suscripción")
		return
	}

	sub, payments, err := h.service.GetByIDWithPayments(r.Context(), id)
	if err != nil {
		utils.WriteNotFound(w, utils.GetPGErrorMessage(err))
		return
	}

	resp := map[string]interface{}{
		"subscription": sub,
		"payments":     payments,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	_, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	subID := r.PathValue("id")
	if subID == "" {
		utils.WriteBadRequest(w, "Falta el id de la suscripción")
		return
	}

	var req UpdateSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	sub, err := h.service.UpdateSubscription(r.Context(), subID, subID, &req)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sub)
}
