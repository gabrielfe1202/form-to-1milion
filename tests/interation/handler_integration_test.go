package interation_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"form-to-1milion/internal/database"
	"form-to-1milion/internal/user"
	utils_test "form-to-1milion/tests/utils"

	"github.com/joho/godotenv"
)

func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env.test")
	if err != nil {
		fmt.Println("Aviso: .env não carregado")
	}

	code := m.Run()
	os.Exit(code)
}

func setupTest(t *testing.T) *user.Handler {
	db := database.Connect()

	database.RunMigrations(db)

	repo := user.NewRepository(db)
	service := user.NewService(repo)

	fakeProducer := &utils_test.FakeProducer{}

	handler := user.NewHandler(service, fakeProducer)

	return handler
}

func TestCreateUser_Success(t *testing.T) {
	handler := setupTest(t)

	body := map[string]string{
		"name":     "Gabriel",
		"email":    "gabriel@test.com",
		"document": "58469521548",
		"phone":    "558426335214",
	}

	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	handler := setupTest(t)

	body := map[string]string{
		"name":     "Gabriel",
		"email":    "dup@test.com",
		"document": "58469521548",
		"phone":    "558426335214",
	}

	jsonBody, _ := json.Marshal(body)

	req1 := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonBody))
	w1 := httptest.NewRecorder()
	handler.Create(w1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonBody))
	w2 := httptest.NewRecorder()
	handler.Create(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w2.Code)
	}
}

func TestCreateUser_InvalidJSON(t *testing.T) {
	handler := setupTest(t)

	req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBuffer([]byte(`{invalid json}`)))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateUser_MethodNotAllowed(t *testing.T) {
	handler := setupTest(t)

	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestListUsers(t *testing.T) {
	handler := setupTest(t)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCountUsers(t *testing.T) {
	handler := setupTest(t)

	req := httptest.NewRequest(http.MethodGet, "/usersCount", nil)
	w := httptest.NewRecorder()

	handler.Count(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
