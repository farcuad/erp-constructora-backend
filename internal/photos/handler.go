package photos

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

// UploadPhotoMetadata recibe la URL ya generada por Supabase Storage junto con las relaciones
func (h *Handler) UploadPhotoMetadata(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	userID, okUser := middlewares.GetUserIDFromContext(r.Context())
	if !ok || !okUser {
		utils.WriteUnauthorized(w)
		return
	}

	var photo ProjectPhoto
	if err := json.NewDecoder(r.Body).Decode(&photo); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	if photo.ProjectID == "" {
		utils.WriteBadRequest(w, "project_id es requerido")
		return
	}
	if photo.PhotoURL == "" {
		utils.WriteBadRequest(w, "photo_url es requerido")
		return
	}

	photo.CompanyID = companyID
	photo.UserID = userID

	if err := h.service.RegisterPhoto(r.Context(), &photo); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(photo)
}

func (h *Handler) GetGallery(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	projectID := r.PathValue("project_id")
	if projectID == "" {
		utils.WriteBadRequest(w, "El parámetro project_id es requerido")
		return
	}

	gallery, err := h.service.GetProjectGallery(r.Context(), companyID, projectID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gallery)
}

func (h *Handler) UpdatePhoto(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el parámetro id")
		return
	}

	var req UpdatePhotoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	if err := h.service.UpdatePhoto(r.Context(), companyID, id, req); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Foto actualizada"})
}

func (h *Handler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el parámetro id")
		return
	}

	if err := h.service.DeletePhoto(r.Context(), companyID, id); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Foto eliminada"})
}
