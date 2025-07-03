package jwt_utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// func ValidateJWT(tokenStr string) (string, error) {

// 	jwtKey := []byte(os.Getenv("jwt_secret_key"))
// 	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
// 		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
// 			return nil, fmt.Errorf("unexpected signing method")
// 		}
// 		return jwtKey, nil
// 	})

// 	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {

// 		if userIDRaw, ok := claims["user_id"]; ok {
// 			if userIDStr, ok := userIDRaw.(string); ok {
// 				return userIDStr, nil
// 			}

// 			return "", fmt.Errorf("user_id is not a string")
// 		}
// 	}
// 	return "", err
// }

func ValidateJWT(tokenString string) (string, string, error) {
	jwtKey := []byte(os.Getenv("jwt_secret_key"))

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Ensure token is signed with expected algorithm
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtKey, nil
	})

	if err != nil {
		return "", "", err
	}

	// Validate claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check expiration
		exp, ok := claims["exp"].(float64)
		if !ok || time.Now().Unix() > int64(exp) {
			return "", "", errors.New("token expired")
		}

		// Extract user_id
		userID, ok := claims["user_id"].(string)
		if !ok {
			return "", "", errors.New("user_id missing in token")
		}

		// Extract roles
		role, ok := claims["role"].(string)
		if !ok {
			return "", "", errors.New("role missing or invalid in token")
		}

		return userID, role, nil
	}

	return "", "", errors.New("invalid token")
}

func GenerateRefreshJWT(userID string, role string) (string, error) {
	jwtKey := []byte(os.Getenv("jwt_secret_key"))
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(), // expires in 24h
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func GenerateAccessJWT(userID string, role string) (string, error) {
	jwtKey := []byte(os.Getenv("jwt_secret_key"))
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(15 * time.Minute).Unix(), // expires in 15mins
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
