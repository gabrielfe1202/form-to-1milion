package user

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"form-to-1milion/internal/observability"
	"form-to-1milion/internal/producer"

	"go.opentelemetry.io/otel/attribute"
)

type Handler struct {
	service  *Service
	producer producer.TaskProducer
	metrics  *observability.Metrics
	emails   map[string]bool
	mu       sync.RWMutex
}

func NewHandler(service *Service, producer producer.TaskProducer, metrics *observability.Metrics) *Handler {
	return &Handler{
		service:  service,
		producer: producer,
		metrics:  metrics,
		emails:   make(map[string]bool),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := observability.TraceUserCreation(r.Context(), "")

	if r.Method != http.MethodPost {
		observability.SpanRecordError(ctx, nil, "invalid_http_method")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limita o tamanho do corpo para evitar DoS e alocações gigantescas.
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20)) // 1MiB
	decoder.DisallowUnknownFields()

	var u User
	if err := decoder.Decode(&u); err != nil {
		observability.SpanRecordError(ctx, err, "failed_to_decode_json")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	u.Normalize()

	if err := u.Validate(); err != nil {
		observability.SpanRecordError(ctx, err, "validation_failed")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.mu.RLock()
	_, exists := h.emails[u.Email]
	h.mu.RUnlock()

	if exists {
		observability.SpanAddEvent(ctx, "email_duplicate",
			attribute.String("email", u.Email))
		http.Error(w, "Email already exists", http.StatusConflict)
		return
	}

	q := time.Now().UTC().UnixNano()
	task := producer.Task{
		//ID:      strconv.FormatInt(q, 10),
		ID:      u.Email,
		Type:    "create_user",
		Payload: u,
		Created: time.Unix(0, q).UTC(),
	}

	// Use context da requisição para cancelamento imediato em shutdown/timeouts.
	if err := h.producer.EnqueueTask(r.Context(), task); err != nil {
		observability.SpanRecordError(ctx, err, "rabbitmq_publish_failed")
		if h.metrics != nil {
			h.metrics.RabbitMQPublishErrors.Inc()
		}
		http.Error(w, "Publish error", http.StatusInternalServerError)
		return
	}

	h.mu.Lock()
	h.emails[u.Email] = true
	h.mu.Unlock()

	// Registrar métricas de sucesso
	if h.metrics != nil {
		h.metrics.UserCreatedTotal.Inc()
		h.metrics.RabbitMQPublishTotal.Inc()
	}

	observability.SpanAddEvent(ctx, "user_created",
		attribute.String("email", u.Email))

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := h.service.ListUsers()
	if err != nil {
		observability.SpanRecordError(ctx, err, "database_query_failed")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Registrar métricas de sucesso
	if h.metrics != nil {
		h.metrics.UserListTotal.Inc()
	}

	observability.SpanAddEvent(ctx, "users_retrieved",
		attribute.Int("count", len(users)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func (h *Handler) Count(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	countUsers, err := h.service.CountUsers()
	if err != nil {
		observability.SpanRecordError(ctx, err, "count_query_failed")
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Registrar métricas de sucesso
	if h.metrics != nil {
		h.metrics.UserCountTotal.Inc()
	}

	observability.SpanAddEvent(ctx, "users_counted",
		attribute.Int("total", countUsers))

	response := struct {
		Total int `json:"total"`
	}{
		Total: countUsers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
