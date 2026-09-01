package utils

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/lib/pq"
)

type ErrorResponse struct {
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Message: message})
}

func WriteMethodNotAllowed(w http.ResponseWriter) {
	WriteError(w, http.StatusMethodNotAllowed, "Método no permitido")
}

func WriteUnauthorized(w http.ResponseWriter) {
	WriteError(w, http.StatusUnauthorized, "No autorizado")
}

func WriteBadRequest(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusBadRequest, message)
}

func WriteNotFound(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusNotFound, message)
}

func WriteInternalError(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusInternalServerError, message)
}

func WriteConflict(w http.ResponseWriter, message string) {
	WriteError(w, http.StatusConflict, message)
}

func WriteCreated(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(data)
}

func WriteOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(data)
}

func GetPGErrorMessage(err error) string {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code {
		case "23505":
			return "Registro duplicado"
		case "23506":
			return "No se puede eliminar porque existen registros relacionados"
		case "23503":
			return "Clave foránea violada"
		case "23514":
			return "Valor viola una restricción de check"
		case "23502":
			return "Campo requerido no puede ser nulo"
		}
		if pqErr.Message != "" {
			return pqErr.Message
		}
	}
	if err != nil {
		msg := err.Error()
		if len(msg) > 200 {
			msg = msg[:200]
		}
		return msg
	}
	return "Error desconocido"
}

func GetUserFriendlyMessage(action string, name string) string {
	switch action {
	case "delete":
		return fmt.Sprintf("No se puede eliminar %s porque tiene registros relacionados", name)
	case "create":
		return fmt.Sprintf("Ya existe un %s con esos datos", name)
	case "update":
		return fmt.Sprintf("No se puede actualizar %s", name)
	}
	return "Error en la operación"
}

func IsForeignKeyViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23506" || pqErr.Code == "23503"
	}
	if err != nil {
		msg := err.Error()
		return strings.Contains(msg, "23506") || strings.Contains(msg, "foreign key") || strings.Contains(msg, "FOREIGN KEY") || strings.Contains(msg, "violates foreign key") || strings.Contains(msg, "insert or update on table") && strings.Contains(msg, "violates foreign key constraint")
	}
	return false
}

func IsUniqueViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		return pqErr.Code == "23505"
	}
	return false
}

func ExecWithPGErrorCheck(fn func() error) error {
	err := fn()
	if err != nil {
		if IsForeignKeyViolation(err) {
			return fmt.Errorf("foreign_key_violation: %s", GetPGErrorMessage(err))
		}
		if IsUniqueViolation(err) {
			return fmt.Errorf("unique_violation: %s", GetPGErrorMessage(err))
		}
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("not_found: %s", GetPGErrorMessage(err))
		}
	}
	return nil
}
