package purchase

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

// --- CONTROLADORES DE ÓRDENES DE COMPRA ---

func (h *Handler) CreatePurchaseOrder(w http.ResponseWriter, r *http.Request) {
	// 1. Extraer tanto el company_id como el user_id usando tus helpers nativos
	companyID, okCompany := users.GetCompanyIDFromContext(r.Context())
	userID, okUser := users.GetUserIDFromContext(r.Context())

	if !okCompany || !okUser {
		http.Error(w, "Credenciales de usuario o empresa no válidas en el contexto", http.StatusUnauthorized)
		return
	}

	var po PurchaseOrder
	if err := json.NewDecoder(r.Body).Decode(&po); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Inyectar de forma segura los datos extraídos del JWT de la sesión
	po.CompanyID = companyID
	po.UserID = userID

	if err := h.service.CreatePurchaseOrder(r.Context(), &po); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(po)
}

func (h *Handler) GetOrdersByProject(w http.ResponseWriter, r *http.Request) {
	// En Go 1.22+, extraemos el parámetro dinámico {project_id} definido en router.go
	projectID := r.PathValue("project_id")
	if projectID == "" {
		http.Error(w, "El parámetro project_id es obligatorio", http.StatusBadRequest)
		return
	}

	orders, err := h.service.ListOrdersByProject(r.Context(), projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}
