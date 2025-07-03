package handlers

import (
	"net/http"
	"rms/dbhelper"
	"rms/middleware"
	"rms/models"
	"rms/utils"

	// "github.com/gorilla/mux"
	"strconv"
	// "strings"
)

func (h *Handler) GetAllRestaurants(w http.ResponseWriter, r *http.Request) {

	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	userID := middleware.GetUserID(r)

	restaurants, err := dbhelper.GetAllRestaurants(h.DB, limit, offset)

	if err != nil {
		http.Error(w, "failed to fetch restaurants", http.StatusInternalServerError)
		return

	}

	// Handle case where no rows matched
	if len(restaurants) == 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "No restaurants found"})
		return
	}

	// Send back the tasks
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"restaurants": restaurants,
		"page":        page,
		"limit":       limit,
		"user_id":     userID,
	})

	if err != nil {
		http.Error(w, "error in encoding data", http.StatusInternalServerError)
		return
	}
}

func RolesForCreation() []string {
	return []string{"superadmin", "admin", "subadmin"}
}

func (h *Handler) CreateRestaurant(w http.ResponseWriter, r *http.Request) {

	userRoles := middleware.GetUserRoles(r)

	// logged in user priority
	userPriority := utils.GetUserPriority(userRoles)

	if userPriority < 2 {
		http.Error(w, "forbidden: not authorized to create restaurant", http.StatusForbidden)
		return
	}

	//authorized user creating rest.:
	var restaurant models.Restaurant
	if err := json.NewDecoder(r.Body).Decode(&restaurant); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	restaurant.CreatedBy = middleware.GetUserID(r)

	err := dbhelper.CreateRestaurant(h.DB, &restaurant)

	if err != nil {
		http.Error(w, "failed to create restaurant", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(restaurant)
	if err != nil {
		http.Error(w, "error in encoding data", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) CalculateDistance(w http.ResponseWriter, r *http.Request) {
	resID := r.URL.Query().Get("res_id")
	addID := r.URL.Query().Get("add_id")

	userID := middleware.GetUserID(r)
	
	// 
	userIDAdd, err := dbhelper.GetUserID(h.DB, addID)
	
	if err != nil {
		http.Error(w, "invalid add id", http.StatusBadRequest)
		return
	}
	if userID != userIDAdd {
		http.Error(w, "no authorization for this user address", http.StatusUnauthorized)
		return
	}

	resLat, resLng, err := dbhelper.GetLatAndLngRest(h.DB, resID)

	if err != nil {
		http.Error(w, "invalid res id", http.StatusBadRequest)
		return
	}

	addLat, addLng, err := dbhelper.GetLatAndLngUser(h.DB, addID)

	if err != nil {
		http.Error(w, "invalid add id", http.StatusBadRequest)
		return
	}

	dist := utils.HaversineDistance(resLat, resLng, addLat, addLng)

	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"distance": dist,
	})
	if err != nil {
		http.Error(w, "error in encoding data", http.StatusInternalServerError)
		return
	}
}
