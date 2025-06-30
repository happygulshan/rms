package handlers

import (
	"fmt"
	"log"
	"net/http"
	"rms/middleware"
	"rms/models"

	"github.com/gorilla/mux"

	"strconv"
	"rms/utils"
)

func (h *Handler) CreateDish(w http.ResponseWriter, r *http.Request) {

	userID := middleware.GetUserID(r)

	//Get user roles
	userRoles := middleware.GetUserRoles(r)

	// logged in user priority
	userPriority := utils.GetUserPriority(userRoles)
	fmt.Println(userPriority)
	//Check if user has permission
	if userPriority < 2 {
		http.Error(w, "forbidden: not authorized to create dish", http.StatusForbidden)
		return
	}

	vars := mux.Vars(r)
	resID := vars["reID"]

	var err error
	// If subadmin, verify ownership of restaurant
	if utils.HasOnlySubadminPrivileges(userRoles) {
		var createdBy string
		err = h.DB.QueryRow("SELECT created_by FROM restaurants WHERE id = $1::uuid", resID).Scan(&createdBy)
		if err != nil {
			http.Error(w, "invalid restaurant ID", http.StatusBadRequest)
			return
		}
		if createdBy != userID {
			http.Error(w, "unauthorized: you do not own this restaurant", http.StatusUnauthorized)
			return
		}
	}

	var dish models.Dish
	if err := json.NewDecoder(r.Body).Decode(&dish); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	err = h.DB.QueryRow("INSERT INTO dishes (restaurant_id, Name, description, price, created_by) VALUES($1::uuid, $2, $3, $4, $5::uuid) RETURNING id, created_at",
		resID, dish.Name, dish.Description, dish.Price, userID).Scan(&dish.ID, &dish.CreatedAt)

	if err != nil {
		log.Println(err.Error(), resID)
		http.Error(w, "failed to create dish", http.StatusInternalServerError)
		return
	}
	dish.RestaurantID = resID
	dish.CreatedBy = userID

	err = json.NewEncoder(w).Encode(dish)
	if err != nil {
		http.Error(w, "error in encoding data", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) GetAllDishes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	restaurantID := vars["reID"]

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
	rows, err := h.DB.Query("SELECT id, restaurant_id, name, description, price, created_at, created_by FROM dishes WHERE restaurant_id = $1::uuid ORDER BY created_at DESC LIMIT $2 OFFSET $3", restaurantID, limit, offset)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "failed to fetch dishes", http.StatusInternalServerError)
		return
	}

	defer rows.Close()

	var dishes []models.Dish

	for rows.Next() {
		var dish models.Dish
		if err := rows.Scan(&dish.ID, &dish.RestaurantID, &dish.Name, &dish.Description, &dish.Price, &dish.CreatedAt, &dish.CreatedBy); err != nil {
			log.Println(err.Error())
			http.Error(w, "failed to scan dish", http.StatusInternalServerError)
			return
		}
		dishes = append(dishes, dish)
	}

	// Handle case where no rows matched
	if len(dishes) == 0 {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "No dishes for this restaurant found"})
		return
	}

	// Send back the tasks
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"restaurants": dishes,
		"page":        page,
		"limit":       limit,
		"user_id":     userID,
	})

	if err != nil {
		http.Error(w, "error in encoding data", http.StatusInternalServerError)
		return
	}
	
}
