package user

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"form-to-1milion/internal/producer"
)

type Handler struct {
	service  *Service
	producer producer.TaskProducer
	emails   map[string]bool
	mu       sync.RWMutex
}

func NewHandler(service *Service, prod producer.TaskProducer) *Handler {
	return &Handler{
		service:  service,
		producer: prod,
		emails:   make(map[string]bool),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limita o tamanho do corpo para evitar DoS
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20)) // 1MiB
	decoder.DisallowUnknownFields()

	var u User
	if err := decoder.Decode(&u); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	u.Normalize()

	if err := u.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.mu.RLock()
	_, exists := h.emails[u.Email]
	h.mu.RUnlock()

	if exists {
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	task := producer.Task{
		ID:      u.Email,
		Type:    "create_user",
		Payload: u,
		Created: time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := h.producer.EnqueueTask(ctx, task); err != nil {
		http.Error(w, "Failed to process request", http.StatusInternalServerError)
		log.Printf("❌ Erro ao enfileirar: %v", err)
		return
	}

	h.mu.Lock()
	h.emails[u.Email] = true
	h.mu.Unlock()

	w.WriteHeader(http.StatusAccepted)
	log.Printf("✅ Request aceita: %s", u.Email)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *Handler) Count(w http.ResponseWriter, r *http.Request) {
	countUsers, err := h.service.CountUsers()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	response := struct {
		Total int `json:"total"`
	}{
		Total: countUsers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
