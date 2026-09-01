package recoverypassword

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service *AuthService
}

func NewHandler(service *AuthService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var dto RequestResetDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if dto.Email == "" {
		http.Error(w, "El correo es obligatorio", http.StatusBadRequest)
		return
	}

	if err := h.service.RequestPasswordReset(dto.Email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Si el correo está registrado, recibirás un código de verificación."})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var dto ResetPasswordDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	if dto.Email == "" || dto.Token == "" || dto.NewPassword == "" {
		http.Error(w, "Todos los campos son obligatorios", http.StatusBadRequest)
		return
	}

	if err := h.service.ResetPassword(dto.Email, dto.Token, dto.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Contraseña actualizada exitosamente."})
}