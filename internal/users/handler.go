package users

import (
	"encoding/json"
	"erp-constructora/internal/middlewares"
	"erp-constructora/internal/services"
	"erp-constructora/internal/utils"
	"fmt"
	"log"
	"net/http"
)

type Handler struct {
	service     *Service
	mailService *services.MailService
}

func NewHandler(service *Service, mailService *services.MailService) *Handler {
	return &Handler{service: service, mailService: mailService}
}

func (h *Handler) RegisterCompanyAndAdmin(w http.ResponseWriter, r *http.Request) {
	// Validar que sea un método POST
	if r.Method != http.MethodPost {
		utils.WriteMethodNotAllowed(w)
		return
	}

	// Decodificar el JSON entrante
	var dto RegisterDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteBadRequest(w, "JSON inválido: "+err.Error())
		return
	}

	// Validaciones básicas de campos obligatorios (Reemplazo de tu antiguo Zod)
	if dto.CompanyName == "" || dto.CompanyNIT == "" || dto.AdminEmail == "" || dto.Password == "" {
		utils.WriteBadRequest(w, "Todos los campos son obligatorios")
		return
	}

	// Llamar al servicio mandando el contexto de la petición HTTP
	company, admin, err := h.service.RegisterCompanyAndAdmin(r.Context(), dto)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	// Enviar email de bienvenida
	subject := "Bienvenido a Construx"
	htmlBody := fmt.Sprintf(`
		<div style="text-align:center;padding:10px 0;">
			<div style="display:inline-block;background-color:#F99B2E;color:#000000;padding:12px 30px;border-radius:6px;font-weight:800;font-size:14px;letter-spacing:1px;margin-bottom:25px;">BIENVENIDO</div>
			<h1 style="color:#000000;font-size:26px;font-weight:700;margin:0 0 10px 0;">%s</h1>
			<p style="color:#333333;font-size:16px;line-height:1.6;margin:0 0 10px 0;">La empresa <strong style="color:#F99B2E;">%s</strong> ha sido registrada exitosamente.</p>
			<p style="color:#666666;font-size:14px;line-height:1.6;">Tu cuenta está lista. Inicia sesión y comienza a gestionar tu constructora.</p>
		</div>`, admin.Name, company.Name)
	if err := h.mailService.SendEmail(admin.Email, subject, htmlBody); err != nil {
		log.Printf("Error enviando email de bienvenida a %s: %v", admin.Email, err)
	}

	// Estructurar respuesta exitosa
	response := map[string]interface{}{
		"message": "Empresa y administrador registrados exitosamente",
		"company": company,
		"admin":   admin,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	// Forzar que siempre que salgamos con un error o éxito, el cliente sepa que es JSON
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed) // 405
		json.NewEncoder(w).Encode(map[string]string{"message": "Método no permitido"})
		return
	}

	var dto LoginDto
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		w.WriteHeader(http.StatusBadRequest) // 400
		json.NewEncoder(w).Encode(map[string]string{"message": "JSON inválido: " + err.Error()})
		return
	}

	if dto.Email == "" || dto.Password == "" {
		w.WriteHeader(http.StatusBadRequest) // 400
		json.NewEncoder(w).Encode(map[string]string{"message": "Todos los campos son obligatorios"})
		return
	}

	result, err := h.service.Login(r.Context(), dto)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized) // 401
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}

	response := map[string]interface{}{
		"message": "Inicio de sesion exitoso",
		"token":   result.Token,
		"user":    result.User,
	}

	w.WriteHeader(http.StatusOK) // 200
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) GetRoles(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	roles, err := h.service.GetRoles(r.Context(), companyID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(roles)
}

// GET /users
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	usersList, err := h.service.GetUsers(r.Context(), companyID)
	if err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(usersList)
}

// POST /users
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	var dto CreateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	user, err := h.service.CreateUser(r.Context(), companyID, dto)
	if err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

// PUT /users
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	// Extraer el ID de la URL (ejemplo usando r.PathValue si usas Go 1.22+ o el enrutador que utilices)
	userID := r.PathValue("id")
	if userID == "" {
		// Ejemplo con chi o gorilla/mux si fuera el caso: userID = chi.URLParam(r, "id")
		utils.WriteBadRequest(w, "ID de usuario no proporcionado en la URL")
		return
	}

	var dto UpdateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	// Garantizar que el ID procesado sea el enviado en la URL
	dto.ID = userID

	if err := h.service.UpdateUser(r.Context(), companyID, dto); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuario actualizado exitosamente"})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		utils.WriteMethodNotAllowed(w)
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w)
		return
	}

	// Extraer el ID de la URL
	userID := r.PathValue("id")
	if userID == "" {
		utils.WriteBadRequest(w, "ID de usuario no proporcionado en la URL")
		return
	}

	if err := h.service.DeleteUser(r.Context(), companyID, userID); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuario eliminado exitosamente"})
}
