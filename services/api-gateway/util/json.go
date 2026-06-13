package util

import (
	"books-and-trust/shared/contracts"
	"encoding/json"
	"net/http"
	"time"
)

// Send Json Success Response
func WriteJSON(w http.ResponseWriter, data any, status int, message string) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := contracts.ResponseSuccess{
		Status:    "success",
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	return json.NewEncoder(w).Encode(resp)
}

// Send Json Error Response
func WriteError(w http.ResponseWriter, httpStatus int, errCode string, message string, details any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)

	resp := contracts.ResponseError{
		Status:    "error",
		Code:      errCode,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return json.NewEncoder(w).Encode(resp)
}

// Read Json
func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxByte := 1_048_578

	r.Body = http.MaxBytesReader(w, r.Body, int64(maxByte))

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}
