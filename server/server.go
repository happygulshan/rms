package server

import (
	"database/sql"
	"net/http"

	"rms/handlers"

	"rms/middleware"

	"github.com/gorilla/mux"
)

func InitRoutes(db *sql.DB) *mux.Router {

	r := mux.NewRouter()
	h := handlers.Handler{DB: db}
	authMiddleware := middleware.AuthMiddleware(h.DB)
	r.HandleFunc("/signup", h.CreateUser)
	r.HandleFunc("/login", h.Login)
	r.Handle("/logout", authMiddleware(http.HandlerFunc(h.Logout)))

	// restaurant work
	r.Handle("/restaurants", authMiddleware(http.HandlerFunc(h.GetAllRestaurants))).Methods("GET")
	r.Handle("/restaurant", authMiddleware(http.HandlerFunc(h.CreateRestaurant))).Methods("POST")

	// dishes work
	r.Handle("/dishes/{reID}", authMiddleware(http.HandlerFunc(h.GetAllDishes))).Methods("GET")
	r.Handle("/dish/{reID}", authMiddleware(http.HandlerFunc(h.CreateDish))).Methods("POST")

	r.Handle("/protected/user", authMiddleware(http.HandlerFunc(h.ProtectedCreateUser))).Methods("POST")
	r.Handle("/distance", authMiddleware(http.HandlerFunc(h.CalculateDistance))).Methods("GET")
	r.Handle("/address", authMiddleware(http.HandlerFunc(h.CreateAddress))).Methods("POST")
	r.HandleFunc("/refresh", h.RefreshToken).Methods("GET")

	return r
}
