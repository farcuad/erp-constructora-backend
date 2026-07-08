package documents

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

func (h *Handler) CreateType(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var t DocumentType
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	t.CompanyID = companyID

	if err := h.service.CreateDocumentType(r.Context(), &t); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// CreateDocument Payload mixto estructurado de forma limpia
func (h *Handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	userID, okUser := middlewares.GetUserIDFromContext(r.Context())
	if !ok || !okUser {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Definimos una estructura anidada temporal para recibir el JSON completo
	var payload struct {
		ProjectID      string `json:"project_id"`
		DocumentTypeID string `json:"document_type_id"`
		Title          string `json:"title"`
		Description    string `json:"description"`
		FileURL        string `json:"file_url"`
		FileSize       int64  `json:"file_size"`
		FileExtension  string `json:"file_extension"`
		ChangeLog      string `json:"change_log"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	doc := Document{
		CompanyID:      companyID,
		ProjectID:      payload.ProjectID,
		DocumentTypeID: payload.DocumentTypeID,
		Title:          payload.Title,
		Description:    payload.Description,
	}

	ver := DocumentVersion{
		CompanyID:     companyID,
		FileURL:       payload.FileURL,
		FileSize:      payload.FileSize,
		FileExtension: payload.FileExtension,
		ChangeLog:     payload.ChangeLog,
		UserID:        userID,
	}

	if err := h.service.UploadInitialDocument(r.Context(), &doc, &ver); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Adjuntamos la versión creada al objeto de respuesta
	doc.Versions = append(doc.Versions, ver)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}

func (h *Handler) UpdateVersion(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	userID, okUser := middlewares.GetUserIDFromContext(r.Context())
	if !ok || !okUser {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var ver DocumentVersion
	if err := json.NewDecoder(r.Body).Decode(&ver); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}
	ver.CompanyID = companyID
	ver.UserID = userID

	if err := h.service.UploadNewVersion(r.Context(), &ver); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ver)
}

func (h *Handler) UpdateDocumentType(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateDocumentTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateDocumentType(r.Context(), companyID, id, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Tipo de documento actualizado"})
}

func (h *Handler) DeleteDocumentType(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.DeleteDocumentType(r.Context(), companyID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Tipo de documento eliminado"})
}

func (h *Handler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if err := h.service.UpdateDocument(r.Context(), companyID, id, req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Documento actualizado"})
}

func (h *Handler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.DeleteDocument(r.Context(), companyID, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Documento eliminado"})
}
