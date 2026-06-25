package purchase

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	companyID, _ := r.Context().Value("company_id").(string)
	userID, _ := r.Context().Value("user_id").(string)

	if companyID == "" || userID == "" {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var req CreatePurchaseOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Payload inválido", http.StatusBadRequest)
		return
	}

	order, err := h.service.GeneratePurchaseOrder(r.Context(), companyID, userID, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(order)
}
