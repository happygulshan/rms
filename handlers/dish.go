package handlers

import (
	"fmt"
	"log"
	"net/http"
	"rms/dbhelper"
	"rms/middleware"
	"rms/models"

	"github.com/gorilla/mux"

	"rms/utils"
	"strconv"
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
		createdBy, err := dbhelper.RestaurantCreatedBy(h.DB, resID)
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

	
	dish.RestaurantID = resID

	dish.CreatedBy = userID
	err = dbhelper.CreateDish(h.DB, &dish)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, "failed to create dish", http.StatusInternalServerError)
		return
	}

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

	dishes, err := dbhelper.GetAllDishes(h.DB, limit, offset, restaurantID)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "failed to fetch dishes", http.StatusInternalServerError)
		return
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
