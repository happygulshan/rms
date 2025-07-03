package handlers

import (
	"log"
	"net/http"
	"rms/dbhelper"
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
	add.UserID = userID

	err := dbhelper.CreateAddress(h.DB, &add)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, "failed to create address", http.StatusInternalServerError)
		return
	}

	// default label is home
	add.Label = models.Home

	err = json.NewEncoder(w).Encode(add)
	if err != nil {
		http.Error(w, "error in encoding data", http.StatusInternalServerError)
		return
	}
}
