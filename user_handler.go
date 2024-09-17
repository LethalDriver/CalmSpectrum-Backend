package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type UserHandler struct {
	s *UserService
}

func NewUserHandler(s *UserService) *UserHandler {
	return &UserHandler{
		s: s,
	}
}

func (h *UserHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var regReq RegistrationRequest
    err = json.Unmarshal(bodyBytes, &regReq)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

	err = h.s.RegisterUser(regReq)
	if err != nil {
		if err.Error() == "user already exists" {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		http.Error(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
	}
}