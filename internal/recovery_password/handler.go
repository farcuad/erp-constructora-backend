package recoverypassword

import (
	"encoding/json"
	"erp-constructora/internal/utils"
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
		utils.WriteMethodNotAllowed(w)
		return
	}

	var dto RequestResetDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	if dto.Email == "" {
		utils.WriteBadRequest(w, "El correo es obligatorio")
		return
	}

	if err := h.service.RequestPasswordReset(dto.Email); err != nil {
		utils.WriteInternalError(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Si el correo está registrado, recibirás un código de verificación."})
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.WriteMethodNotAllowed(w)
		return
	}

	var dto ResetPasswordDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		utils.WriteBadRequest(w, "JSON inválido")
		return
	}

	if dto.Email == "" || dto.Token == "" || dto.NewPassword == "" {
		utils.WriteBadRequest(w, "Todos los campos son obligatorios")
		return
	}

	if err := h.service.ResetPassword(dto.Email, dto.Token, dto.NewPassword); err != nil {
		utils.WriteBadRequest(w, utils.GetPGErrorMessage(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Contraseña actualizada exitosamente."})
}
