package photos

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

// UploadPhotoMetadata recibe la URL ya generada por Supabase Storage junto con las relaciones
func (h *Handler) UploadPhotoMetadata(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	userID, okUser := middlewares.GetUserIDFromContext(r.Context())
	if !ok || !okUser {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var photo ProjectPhoto
	if err := json.NewDecoder(r.Body).Decode(&photo); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	photo.CompanyID = companyID
	photo.UserID = userID

	if err := h.service.RegisterPhoto(r.Context(), &photo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(photo)
}

func (h *Handler) GetGallery(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	projectID := r.PathValue("project_id")
	if projectID == "" {
		http.Error(w, "El parámetro project_id es requerido", http.StatusBadRequest)
		return
	}

	gallery, err := h.service.GetProjectGallery(r.Context(), companyID, projectID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(gallery)
}

func (h *Handler) UpdatePhoto(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Falta el parámetro id", http.StatusBadRequest)
		return
	}

	var req UpdatePhotoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdatePhoto(r.Context(), companyID, id, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Foto actualizada"})
}

func (h *Handler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "Falta el parámetro id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeletePhoto(r.Context(), companyID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Foto eliminada"})
}
