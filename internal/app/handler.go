package app

import (
	"encoding/json"
	"erp-constructora/internal/utils"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// PublishRelease se dispara al compilar la app: guarda la nueva versión.
func (h *Handler) PublishRelease(w http.ResponseWriter, r *http.Request) {
	var req CreateReleaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "Formato JSON inválido")
		return
	}

	rel, err := h.service.PublishRelease(r.Context(), &req)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(rel)
}

// GetLatestRelease devuelve la subida más reciente de la app.
// Lo consume la web (para descargar el .apk) y la app móvil (para comparar versiones).
func (h *Handler) GetLatestRelease(w http.ResponseWriter, r *http.Request) {
	rel, err := h.service.GetLatestRelease(r.Context())
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}
	if rel == nil {
		utils.WriteNotFound(w, "No hay ninguna versión publicada de la app")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rel)
}
