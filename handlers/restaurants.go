package handlers

import (
	"fmt"
	"log"
	"net/http"
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
	rows, err := h.DB.Query("SELECT id, name, description, lat, lng, created_at, created_by FROM restaurants ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "failed to fetch restaurants", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var restaurants []models.Restaurant

	for rows.Next() {
		var restaurant models.Restaurant
		if err := rows.Scan(&restaurant.ID, &restaurant.Name, &restaurant.Description, &restaurant.Lat, &restaurant.Lng, &restaurant.CreatedAt, &restaurant.CreatedBy); err != nil {
			log.Println(err.Error())
			http.Error(w, "failed to scan restaurant", http.StatusInternalServerError)
			return
		}
		restaurants = append(restaurants, restaurant)
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

	userID := middleware.GetUserID(r)

	userRoles := middleware.GetUserRoles(r)

	// logged in user priority
	userPriority := utils.GetUserPriority(userRoles)

	if userPriority < 2 {
		http.Error(w, "forbidden: not authorized to create restaurant", http.StatusForbidden)
		return
	}

	//authorized user for rest. creation:
	var restaurant models.Restaurant
	if err := json.NewDecoder(r.Body).Decode(&restaurant); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}
	err := h.DB.QueryRow("INSERT INTO restaurants (Name, description, lat, lng, created_by) VALUES($1, $2, $3, $4, $5) RETURNING id",
		restaurant.Name, restaurant.Description, restaurant.Lat, restaurant.Lng, userID).Scan(&restaurant.ID)

	if err != nil {
		http.Error(w, "failed to create task", http.StatusInternalServerError)
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
	var addUserId string
	err := h.DB.QueryRow("SELECT user_id FROM addresses WHERE id = $1", addID).Scan(&addUserId)

	if err != nil {
		http.Error(w, "wrong address id", http.StatusBadRequest)
		return
	}

	if userID != addUserId {
		http.Error(w, "no authorization for this user address", http.StatusUnauthorized)
		return
	}

	var resLat, resLng float64
	err = h.DB.QueryRow("SELECT lat, lng FROM restaurants WHERE id = $1", resID).Scan(&resLat, &resLng)

	if err != nil {
		http.Error(w, "invalid res id", http.StatusBadRequest)
		return
	}

	var addLat, addLng float64
	err = h.DB.QueryRow("SELECT lat, lng FROM addresses WHERE id = $1", addID).Scan(&addLat, &addLng)

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
