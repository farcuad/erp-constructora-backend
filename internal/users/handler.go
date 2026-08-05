package users

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

func (h *Handler) RegisterCompanyAndAdmin(w http.ResponseWriter, r *http.Request) {
	// Validar que sea un método POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Decodificar el JSON entrante
	var dto RegisterDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Validaciones básicas de campos obligatorios (Reemplazo de tu antiguo Zod)
	if dto.CompanyName == "" || dto.CompanyNIT == "" || dto.AdminEmail == "" || dto.Password == "" {
		http.Error(w, "Todos los campos son obligatorios", http.StatusBadRequest)
		return
	}

	// Llamar al servicio mandando el contexto de la petición HTTP
	company, admin, err := h.service.RegisterCompanyAndAdmin(r.Context(), dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	roles, err := h.service.GetRoles(r.Context(), companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	usersList, err := h.service.GetUsers(r.Context(), companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	var dto CreateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	user, err := h.service.CreateUser(r.Context(), companyID, dto)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
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
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Extraer el ID de la URL (ejemplo usando r.PathValue si usas Go 1.22+ o el enrutador que utilices)
	userID := r.PathValue("id")
	if userID == "" {
		// Ejemplo con chi o gorilla/mux si fuera el caso: userID = chi.URLParam(r, "id")
		http.Error(w, "ID de usuario no proporcionado en la URL", http.StatusBadRequest)
		return
	}

	var dto UpdateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	// Garantizar que el ID procesado sea el enviado en la URL
	dto.ID = userID

	if err := h.service.UpdateUser(r.Context(), companyID, dto); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuario actualizado exitosamente"})
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	companyID, ok := middlewares.GetCompanyIDFromContext(r.Context())
	if !ok {
		http.Error(w, "No autorizado", http.StatusUnauthorized)
		return
	}

	// Extraer el ID de la URL
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "ID de usuario no proporcionado en la URL", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteUser(r.Context(), companyID, userID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuario eliminado exitosamente"})
}
