package user

import (
	"errors"
	"form-to-1milion/internal/utils/validation"
	"strings"
)

type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Document string `json:"document"`
	Phone    string `json:"phone"`
}

func (u *User) Normalize() {
	u.Name = strings.TrimSpace(u.Name)
	u.Email = strings.TrimSpace(u.Email)
	u.Email = strings.ToLower(u.Email)
	u.Phone = validation.OnlyDigits(u.Phone)
	u.Document = validation.OnlyDigits(u.Document)
}

func (u *User) Validate() error {
	if u.Name == "" {
		return errors.New("name is required")
	}
	if len(u.Name) < 3 || len(u.Name) > 100 {
		return errors.New("name must be between 3 and 100 characters")
	}

	if u.Email == "" {
		return errors.New("Email is required")
	}
	if !validation.IsValidEmail(u.Email) {
		return errors.New("invalid email format")
	}

	if u.Document == "" {
		return errors.New("Document is required")
	}

	if !validation.IsOnlyNumbers(u.Document) {
		return errors.New("document must contain only numbers")
	}

	if u.Phone == "" || !validation.IsOnlyNumbers(u.Phone) {
		return errors.New("phone must contain only numbers")
	}

	return nil
}
