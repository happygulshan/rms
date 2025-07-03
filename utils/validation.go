package utils

import (
	"errors"
	"rms/models"
	"strings"
)

// Simple email validation but need complex regex for production
func isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

func ValidateUserInput(user *models.Users) error {
	if user.Name == "" {
		return errors.New("name is required")
	}
	if user.Email == "" || !isValidEmail(user.Email) {
		return errors.New("invalid email")
	}
	if len(user.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func ValidateProtectedUserInput(user *models.Users) error {
	if user.Name == "" {
		return errors.New("name is required")
	}
	if user.Email == "" || !isValidEmail(user.Email) {
		return errors.New("invalid email")
	}
	if len(user.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	if user.Role == "" {
		return errors.New("role cant be mepty")
	}
	return nil
}
