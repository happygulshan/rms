package handlers

import (
	"database/sql"
	"net/http"
	"rms/models"
	"time"

	"log"

	"fmt"
	"rms/jwt_utils"

	"rms/middleware"

	"rms/utils"

	"rms/dbhelper"

	"golang.org/x/crypto/bcrypt"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type Handler struct {
	DB *sql.DB
}

// func GetMaxPriority(userRoles []string) int {

// }
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {

	//validating refresh token
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		http.Error(w, "refresh token not found", http.StatusUnauthorized)
		return
	}

	refreshToken := cookie.Value
	fmt.Println("refresh token:", refreshToken)

	// Validate the refresh token
	userID, role, err := jwt_utils.ValidateJWT(refreshToken)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	access_token, err := jwt_utils.GenerateAccessJWT(userID, role)
	if err != nil {
		http.Error(w, "failed to generate login access token", http.StatusInternalServerError)
		return
	}

	// Return token to client
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": access_token,
		"msg":          "Access token refreshed",
	})
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {

	var user models.Users

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "wrong json data", http.StatusBadRequest)
		return
	}

	if err := utils.ValidateUserInput(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	taken, err := dbhelper.IsEmailTaken(h.DB, user.Email)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	} else if taken {
		http.Error(w, "email already registered", http.StatusBadRequest)
		return
	}

	hash, err := utils.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "error in hashing password", http.StatusInternalServerError)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "failed to start transaction", http.StatusInternalServerError)
		return
	}

	defer dbhelper.TxFinalizer(tx, &err)

	err = dbhelper.CreateUserWithRole(tx, &user, hash, "user", "")

	if err != nil {
		http.Error(w, "failed to register user", http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(map[string]string{
		"msg": "user registered successfully. Please login again to use",
	})

	if err != nil {
		http.Error(w, "error in encoding data", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) ProtectedCreateUser(w http.ResponseWriter, r *http.Request) {

	var user models.Users

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "wrong json data", http.StatusBadRequest)
		return
	}

	//basic input data simple validation
	if err := utils.ValidateProtectedUserInput(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r)

	// getting role of logged in user
	userRole := middleware.GetUserRole(r)

	// logged in user priority
	userPriority := utils.RolePriorityMap[userRole]

	// priority for the role creation
	rolePriority, exists := utils.RolePriorityMap[user.Role]
	if !exists {
		http.Error(w, "invalid role "+user.Role, http.StatusBadRequest)
		return
	}

	// if role creation is greater or equal then not allowed
	if rolePriority >= userPriority {
		fmt.Println(rolePriority, userPriority)
		http.Error(w, "you dont have priviledge to create this role", http.StatusUnauthorized)
		return
	}

	taken, err := dbhelper.IsEmailTaken(h.DB, user.Email)
	if err != nil {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	} else if taken {
		http.Error(w, "email already registered", http.StatusBadRequest)
		return
	}

	hash, err := utils.HashPassword(user.Password)
	if err != nil {
		http.Error(w, "error in hashing password", http.StatusInternalServerError)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "failed to start transaction", http.StatusInternalServerError)
		return
	}

	defer dbhelper.TxFinalizer(tx, &err)

	err = dbhelper.CreateUserWithRole(tx, &user, hash, user.Role, userID)

	if err != nil {
		http.Error(w, "failed to register user", http.StatusInternalServerError)
		return
	}

	msg := "one " + user.Role + " registered successfully"

	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"msg":     msg,
		"details": user,
	})
	if err != nil {
		http.Error(w, "error in encoding data", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {

	var user models.Users
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	if user.Role == "" {
		user.Role = "user"
	}

	givenPass := user.Password
	err := dbhelper.GetUserDetails(h.DB, &user)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "invalid email or role", http.StatusUnauthorized)
		return
	}

	// Compare the provided password with the hashed one
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(givenPass)); err != nil {
		http.Error(w, "invalid password", http.StatusUnauthorized)
		return
	}

	// userRoles := services.GetUserRoles(h.DB, user.Id)
	accessToken, err := jwt_utils.GenerateAccessJWT(user.Id, user.Role)
	log.Println(err)
	if err != nil {
		http.Error(w, "failed to generate login access token", http.StatusInternalServerError)
		return
	}

	// Generate refresh token
	refreshToken, err := jwt_utils.GenerateRefreshJWT(user.Id, user.Role)
	if err != nil {
		http.Error(w, "failed to generate login refresh token", http.StatusInternalServerError)
		return
	}

	// Set refresh token in a secure HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	// Return token to client
	err = json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
		"msg":          "Login Successful",
	})

	if err != nil {
		http.Error(w, "error in encoding data", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the refresh_token cookie by setting it expired
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production (when using HTTPS)
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,              // Delete immediately
		Expires:  time.Unix(0, 0), // Extra: for browser compatibility
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Logged out successfully"))
}
