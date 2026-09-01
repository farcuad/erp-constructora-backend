package documents

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
		log.Printf("[NOTIFY ERROR] documents: %v", err)
	}
}

func (h *Handler) CreateType(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var t DocumentType
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}
	t.CompanyID = companyID

	if err := h.service.CreateDocumentType(r.Context(), &t); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "DOCUMENT_TYPE",
		EntityID:   &t.ID,
		Type:       "DOCUMENT_TYPE_CREATED",
		Priority:   notifications.PriorityLow,
		Title:      "Nuevo tipo de documento",
		Message:    fmt.Sprintf("Se creó el tipo de documento: %s.", t.Name),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(t)
}

// CreateDocument Payload mixto estructurado de forma limpia
func (h *Handler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	userID, okUser := middlewares.GetUserIDFromContext(r.Context())
	if !ok || !okUser {
		utils.WriteUnauthorized(w)
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
		utils.WriteBadRequest(w, "JSON inválido")
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
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		ProjectID:  &doc.ProjectID,
		EntityType: "DOCUMENT",
		EntityID:   &doc.ID,
		Type:       "DOCUMENT_CREATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Nuevo documento",
		Message:    fmt.Sprintf("Se añadió el documento: %s.", doc.Title),
		LinkToUI:   strPtr("/dashboard/projects/" + doc.ProjectID + "/documents"),
	})

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
		utils.WriteUnauthorized(w)
		return
	}

	var ver DocumentVersion
	if err := json.NewDecoder(r.Body).Decode(&ver); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}
	ver.CompanyID = companyID
	ver.UserID = userID

	if err := h.service.UploadNewVersion(r.Context(), &ver); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "DOCUMENT_VERSION",
		EntityID:   &ver.ID,
		Type:       "DOCUMENT_VERSION_UPDATED",
		Priority:   notifications.PriorityMedium,
		Title:      "Nueva versión de documento",
		Message:    "Se subió una nueva versión de un documento.",
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ver)
}

func (h *Handler) UpdateDocumentType(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateDocumentTypeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	if err := h.service.UpdateDocumentType(r.Context(), companyID, id, req); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "DOCUMENT_TYPE",
		EntityID:   &id,
		Type:       "DOCUMENT_TYPE_UPDATED",
		Priority:   notifications.PriorityLow,
		Title:      "Tipo de documento actualizado",
		Message:    "Se actualizó un tipo de documento.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Tipo de documento actualizado"})
}

func (h *Handler) DeleteDocumentType(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.DeleteDocumentType(r.Context(), companyID, id); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "DOCUMENT_TYPE",
		EntityID:   &id,
		Type:       "DOCUMENT_TYPE_DELETED",
		Priority:   notifications.PriorityLow,
		Title:      "Tipo de documento eliminado",
		Message:    "Se eliminó un tipo de documento.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Tipo de documento eliminado"})
}

func (h *Handler) UpdateDocument(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	if err := h.service.UpdateDocument(r.Context(), companyID, id, req); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "DOCUMENT",
		EntityID:   &id,
		Type:       "DOCUMENT_UPDATED",
		Priority:   notifications.PriorityLow,
		Title:      "Documento actualizado",
		Message:    "Se actualizó un documento.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Documento actualizado"})
}

func (h *Handler) GetTypes(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	types, err := h.service.GetDocumentTypes(r.Context(), companyID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types)
}

func (h *Handler) GetDocuments(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	projectID := r.PathValue("project_id")
	if projectID == "" {
		utils.WriteBadRequest(w, "Falta project_id en la ruta")
		return
	}

	docs, err := h.service.GetProjectDocuments(r.Context(), companyID, projectID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(docs)
}

func (h *Handler) GetDocumentByID(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	id := r.PathValue("id")
	if id == "" {
		utils.WriteBadRequest(w, "Falta el id del documento")
		return
	}

	doc, err := h.service.GetDocumentByID(r.Context(), companyID, id)
	if err != nil {
		utils.WriteNotFound(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func (h *Handler) GetVersions(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	documentID := r.PathValue("document_id")
	if documentID == "" {
		utils.WriteBadRequest(w, "Falta el id del documento")
		return
	}

	versions, err := h.service.GetDocumentVersions(r.Context(), companyID, documentID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

func (h *Handler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
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

	if err := h.service.DeleteDocument(r.Context(), companyID, id); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	h.notify(r.Context(), notifications.CreateNotificationRequest{
		EntityType: "DOCUMENT",
		EntityID:   &id,
		Type:       "DOCUMENT_DELETED",
		Priority:   notifications.PriorityLow,
		Title:      "Documento eliminado",
		Message:    "Se eliminó un documento.",
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Documento eliminado"})
}
