package inventory

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

func (h *Handler) CreateWarehouse(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var wh Warehouse
	if err := json.NewDecoder(r.Body).Decode(&wh); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}
	wh.CompanyID = companyID

	if err := h.service.CreateWarehouse(r.Context(), &wh); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wh)
}

func (h *Handler) CreateMaterial(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var m Material
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}
	m.CompanyID = companyID

	if err := h.service.CreateMaterial(r.Context(), &m); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func (h *Handler) GetAllMaterials(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	material, err := h.service.GetMaterials(r.Context(), companyID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(material)
}

func (h *Handler) PostMovement(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var m StockMovement
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}
	m.UserID = userID

	if err := h.service.RegisterMovement(r.Context(), &m); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func (h *Handler) GetStock(w http.ResponseWriter, r *http.Request) {
	warehouseID := r.PathValue("warehouse_id")
	if warehouseID == "" {
		utils.WriteBadRequest(w, "Falta warehouse_id")
		return
	}

	stock, err := h.service.GetCurrentStock(r.Context(), warehouseID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stock)
}

func (h *Handler) UpdateMaterial(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "El parámetro id es obligatorio")
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var req UpdateMaterialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido: "+err.Error())
		return
	}

	m, err := h.service.UpdateMaterial(r.Context(), id, companyID, &req)
	if err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}

func (h *Handler) DeleteMaterial(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "El parámetro id es obligatorio")
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	if err := h.service.DeleteMaterial(r.Context(), id, companyID); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "recurso eliminado"})
}

func (h *Handler) GetAllWarehouses(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	suppliers, err := h.service.GetAllWarehouse(r.Context(), companyID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suppliers)
}

func (h *Handler) UpdateWarehouse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "El parámetro id es obligatorio")
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var req UpdateWarehouseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido: "+err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")

	wh, err := h.service.UpdateWarehouse(r.Context(), id, companyID, &req)
	if err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	json.NewEncoder(w).Encode(wh)
}

func (h *Handler) DeleteWarehouse(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "El parámetro id es obligatorio")
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	if err := h.service.DeleteWarehouse(r.Context(), id, companyID); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "recurso eliminado"})
}
