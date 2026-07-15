package users

import (
	"encoding/json"
	"net/http"
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

	token, err := h.service.Login(r.Context(), dto)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized) // 401
		json.NewEncoder(w).Encode(map[string]string{"message": err.Error()})
		return
	}

	response := map[string]interface{}{
		"message": "Inicio de sesion exitoso",
		"token":   token,
	}

	w.WriteHeader(http.StatusOK) // 200
	json.NewEncoder(w).Encode(response)
}
