package handlers

import (
	"log"
	"net/http"
	"rms/middleware"
	"rms/models"
)

func (h *Handler) CreateAddress(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	var add models.Addresses
	if err := json.NewDecoder(r.Body).Decode(&add); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	err := h.DB.QueryRow("INSERT INTO addresses (user_id, lat, lng) VALUES($1::uuid, $2, $3) RETURNING id",
		userID, add.Lat, add.Lng).Scan(&add.Id)

	// default label is home
	add.Label = models.Home

	if err != nil {
		log.Println(err.Error())
		http.Error(w, "failed to create address", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(add)
	if err != nil {
		http.Error(w, "error in encoding data", http.StatusInternalServerError)
		return
	}
}
