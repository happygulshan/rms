package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"rms/models"
	"strings"
	"time"

	"log"

	"fmt"
	"rms/jwt_utils"

	"rms/middleware"

	"rms/utils"

	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	DB *sql.DB
}

func (h *Handler) GetUserRoles(id string) []string {
	//Fetch role names for the user
	query := `
		SELECT r.name 
		FROM user_roles ur
		JOIN roles r ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`

	rows, err := h.DB.Query(query, id)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var userRoles []string
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil
		}
		userRoles = append(userRoles, role)
	}
	return userRoles

}

// Simple email validation but need complex regex for production
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
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
	userID, err := jwt_utils.ValidateJWT(refreshToken)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	userRoles := h.GetUserRoles(userID)
	access_token, err := jwt_utils.GenerateAccessJWT(userID, userRoles)
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

	//basic simple validation
	if user.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if user.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	if user.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	// Validate email format (simple regex)
	if !isValidEmail(user.Email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}

	// Validate password strength
	if len(user.Password) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	err := h.DB.QueryRow("SELECT id FROM users WHERE email = $1", user.Email).Scan(&user.Id)

	if err == nil {
		http.Error(w, "email already registered", http.StatusBadRequest)
		return
	} else if err != sql.ErrNoRows {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error in hashing password", http.StatusInternalServerError)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = tx.QueryRow("INSERT INTO users(name, email, password) VALUES($1, $2, $3) RETURNING id",
		user.Name, user.Email, hashPassword).Scan(&user.Id)

	if err != nil {
		http.Error(w, "failed to register user", http.StatusInternalServerError)
		return
	}

	var roleID string
	err = tx.QueryRow("SELECT id FROM roles WHERE name = 'user'").Scan(&roleID)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, "something wrong with server", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO user_roles(user_id, role_id) VALUES($1, $2)", user.Id, roleID)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, "wrong in assigning roles(internal error)", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"msg": "user registered successfully. Please login again to use",
	})

}

func (h *Handler) ProtectedCreateUser(w http.ResponseWriter, r *http.Request) {

	var user models.ProtectedUsers

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "wrong json data", http.StatusBadRequest)
		return
	}

	//basic input data simple validation
	if user.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if user.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}
	if user.Password == "" {
		http.Error(w, "password is required", http.StatusBadRequest)
		return
	}

	// Validate email format (simple regex)
	if !isValidEmail(user.Email) {
		http.Error(w, "invalid email format", http.StatusBadRequest)
		return
	}

	// Validate password strength
	if len(user.Password) < 8 {
		http.Error(w, "password must be at least 8 characters", http.StatusBadRequest)
		return
	}

	userID := middleware.GetUserID(r)

	// getting role of logged in user
	userRoles := middleware.GetUserRoles(r)

	// logged in user priority
	userPriority := utils.GetUserPriority(userRoles)

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

	err := h.DB.QueryRow("SELECT id FROM users WHERE email = $1", user.Email).Scan(&user.Id)

	if err == nil {
		http.Error(w, "email already registered", http.StatusBadRequest)
		return
	} else if err != sql.ErrNoRows {
		http.Error(w, "database error", http.StatusInternalServerError)
		return
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "error in hashing password", http.StatusInternalServerError)
		return
	}

	tx, err := h.DB.Begin()
	if err != nil {
		http.Error(w, "failed to start transaction", http.StatusInternalServerError)
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	err = tx.QueryRow("INSERT INTO users(name, email, password, created_by) VALUES($1, $2, $3, $4) RETURNING id",
		user.Name, user.Email, hashPassword, userID).Scan(&user.Id)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, "failed to register user", http.StatusInternalServerError)
		return
	}

	var roleID string
	err = tx.QueryRow("SELECT id FROM roles WHERE name = $1", user.Role).Scan(&roleID)

	if err != nil {
		log.Println(err.Error())
		http.Error(w, "something wrong with server", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO user_roles(user_id, role_id) VALUES($1, $2)", user.Id, roleID)
	if err != nil {
		log.Println(err.Error())
		http.Error(w, "wrong in assigning roles(internal error)", http.StatusInternalServerError)
		return
	}

	msg := "one " + user.Role + " registered successfully"

	json.NewEncoder(w).Encode(map[string]interface{}{
		"msg":     msg,
		"details": user,
	})

}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	var user models.Users
	var hashedPassword string

	// Only fetch hashed password
	err := h.DB.QueryRow("SELECT id, name, email, password FROM users WHERE email=$1", req.Email).
		Scan(&user.Id, &user.Name, &user.Email, &hashedPassword)

	if err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	// Compare the provided password with the hashed one
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		http.Error(w, "invalid email or password", http.StatusUnauthorized)
		return
	}

	userRoles := h.GetUserRoles(user.Id)
	accessToken, err := jwt_utils.GenerateAccessJWT(user.Id, userRoles)
	log.Println(err)
	if err != nil {
		http.Error(w, "failed to generate login access token", http.StatusInternalServerError)
		return
	}

	// Generate refresh token
	refreshToken, err := jwt_utils.GenerateRefreshJWT(user.Id)
	if err != nil {
		http.Error(w, "failed to generate login refresh token", http.StatusInternalServerError)
		return
	}

	// Validate the refresh token
	_, err = jwt_utils.ValidateJWT(refreshToken)
	if err != nil {
		http.Error(w, "invalid refresh token", http.StatusUnauthorized)
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
	json.NewEncoder(w).Encode(map[string]string{
		"access_token": accessToken,
		"msg":          "Login Successful",
	})
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
