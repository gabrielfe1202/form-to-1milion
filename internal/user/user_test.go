package user

import (
	"testing"
)

func TestUser_Normalize(t *testing.T) {
	u := &User{
		Name:     "  Gabriel Ferreira  ",
		Email:    "  TEST@EMAIL.COM  ",
		Phone:    "(11) 99999-8888",
		Document: "123.456.789-00",
	}

	u.Normalize()

	if u.Name != "Gabriel Ferreira" {
		t.Errorf("expected trimmed name, got %s", u.Name)
	}

	if u.Email != "test@email.com" {
		t.Errorf("expected normalized email, got %s", u.Email)
	}

	if u.Phone != "11999998888" {
		t.Errorf("expected only digits phone, got %s", u.Phone)
	}

	if u.Document != "12345678900" {
		t.Errorf("expected only digits document, got %s", u.Document)
	}
}

func TestUser_Validate_Success(t *testing.T) {
	u := &User{
		Name:     "Gabriel Ferreira",
		Email:    "gabriel@email.com",
		Document: "12345678900",
		Phone:    "11999998888",
	}

	err := u.Validate()

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestUser_Validate_NameRequired(t *testing.T) {
	u := &User{
		Name:  "",
		Email: "test@email.com",
	}

	err := u.Validate()

	if err == nil || err.Error() != "name is required" {
		t.Errorf("expected 'name is required', got %v", err)
	}
}

func TestUser_Validate_NameSize(t *testing.T) {
	u := &User{
		Name:  "ab",
		Email: "test@email.com",
	}

	err := u.Validate()

	if err == nil || err.Error() != "name must be between 3 and 100 characters" {
		t.Errorf("expected name size error, got %v", err)
	}
}

func TestUser_Validate_EmailRequired(t *testing.T) {
	u := &User{
		Name:  "Gabriel",
		Email: "",
	}

	err := u.Validate()

	if err == nil || err.Error() != "Email is required" {
		t.Errorf("expected 'email is required', got %v", err)
	}
}

func TestUser_Validate_InvalidEmail(t *testing.T) {
	u := &User{
		Name:  "Gabriel",
		Email: "invalid-email",
	}

	err := u.Validate()

	if err == nil || err.Error() != "invalid email format" {
		t.Errorf("expected invalid email error, got %v", err)
	}
}

func TestUser_Validate_InvalidDocument(t *testing.T) {
	u := &User{
		Name:     "Gabriel",
		Email:    "gabriel@email.com",
		Document: "123abc",
	}

	err := u.Validate()

	if err == nil || err.Error() != "document must contain only numbers" {
		t.Errorf("expected document error, got %v", err)
	}
}

func TestUser_Validate_InvalidPhone(t *testing.T) {
	u := &User{
		Name:     "Gabriel",
		Email:    "gabriel@email.com",
		Document: "68526842595",
		Phone:    "9999-abc",
	}

	err := u.Validate()

	if err == nil || err.Error() != "phone must contain only numbers" {
		t.Errorf("expected phone error, got %v", err)
	}
}
