package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"form-to-1milion/internal/consumer"
	"form-to-1milion/internal/database"
	"form-to-1milion/internal/producer"
	"form-to-1milion/internal/user"
	"form-to-1milion/internal/utils/env"
)

// setupSQSClient cria um cliente SQS conectado ao LocalStack
func setupSQSClient() (*sqs.Client, error) {
	sqsEndpoint := env.Get("AWS_ENDPOINT_URL", "http://localhost:4566")
	awsRegion := env.Get("AWS_REGION", "us-east-2")

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("falha ao carregar AWS config: %w", err)
	}

	client := sqs.NewFromConfig(cfg, func(opts *sqs.Options) {
		opts.BaseEndpoint = aws.String(sqsEndpoint)
	})

	return client, nil
}

// purgeQueue limpa todas as mensagens da fila antes do teste
func purgeQueue(t *testing.T, client *sqs.Client, queueURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.PurgeQueue(ctx, &sqs.PurgeQueueInput{
		QueueUrl: aws.String(queueURL),
	})
	if err != nil {
		t.Logf("⚠️ Aviso ao limpar fila: %v", err)
	}
}

// consumeMessages consome mensagens da fila e processa com o handler
func consumeMessages(ctx context.Context, t *testing.T, sqsConsumer *consumer.SQSConsumer, service *user.Service, maxMessages int) {
	processedCount := 0

	handler := func(task consumer.Task) error {
		processedCount++

		if task.Type == "create_user" {
			var u user.User
			payloadBytes, err := json.Marshal(task.Payload)
			if err != nil {
				return fmt.Errorf("erro ao serializar payload: %w", err)
			}

			if err := json.Unmarshal(payloadBytes, &u); err != nil {
				return fmt.Errorf("erro ao desserializar user: %w", err)
			}

			if err := service.CreateUser(u); err != nil {
				return fmt.Errorf("erro ao criar usuário: %w", err)
			}
		}

		return nil
	}

	// Consumir mensagens com timeout
	consumeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	go func() {
		err := sqsConsumer.Consume(consumeCtx, handler)
		if err != nil && err != context.DeadlineExceeded {
			t.Logf("❌ Erro ao consumir mensagens: %v", err)
		}
	}()

	// Aguardar processamento das mensagens
	time.Sleep(2 * time.Second)
}

// setupTestE2E retorna os componentes necessários para testes e2e com SQS real
func setupTestE2E(t *testing.T) (*user.Handler, *user.Service, producer.TaskProducer, *consumer.SQSConsumer) {
	// Conectar ao banco de dados
	db := database.Connect()
	database.RunMigrations(db)

	repo := user.NewRepository(db)
	service := user.NewService(repo)

	// Conectar ao SQS
	sqsClient, err := setupSQSClient()
	if err != nil {
		t.Fatalf("❌ Falha ao conectar ao SQS: %v", err)
	}

	queueURL := env.Get("SQS_QUEUE_URL", "http://localhost:4566/000000000000/form-to-1milion-queue")

	// Criar producer SQS real
	sqsProducer, err := producer.NewSQSProducer(sqsClient, queueURL)
	if err != nil {
		t.Fatalf("❌ Falha ao criar SQS producer: %v", err)
	}

	// Criar consumer SQS real
	sqsConsumer, err := consumer.NewSQSConsumer(sqsClient, queueURL)
	if err != nil {
		t.Fatalf("❌ Falha ao criar SQS consumer: %v", err)
	}

	// Limpar a fila antes do teste
	purgeQueue(t, sqsClient, queueURL)

	// Aguardar um pouco para garantir que a fila foi purgada
	time.Sleep(500 * time.Millisecond)

	handler := user.NewHandler(service, sqsProducer)

	return handler, service, sqsProducer, sqsConsumer
}

// E2E_CompleteUserJourney testa o fluxo completo de um usuário
// desde a criação até a listagem e contagem usando SQS real
func TestE2E_CompleteUserJourney(t *testing.T) {
	handler, service, _, sqsConsumer := setupTestE2E(t)

	t.Run("Create user and verify in list", func(t *testing.T) {
		// 1. Criar um novo usuário
		createBody := map[string]string{
			"name":     "John Doe",
			"email":    "johndoe@test.com",
			"document": "12345678900",
			"phone":    "5511987654321",
		}

		jsonBody, _ := json.Marshal(createBody)
		createReq := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonBody))
		createReq.Header.Set("Content-Type", "application/json")
		createW := httptest.NewRecorder()

		handler.Create(createW, createReq)

		if createW.Code != http.StatusAccepted {
			t.Fatalf("❌ Failed to create user: expected 202, got %d", createW.Code)
		}

		// 2. Consumir e processar mensagens da fila SQS real
		ctx := context.Background()
		consumeMessages(ctx, t, sqsConsumer, service, 1)

		// 3. Listar usuários
		listReq := httptest.NewRequest(http.MethodGet, "/users", nil)
		listW := httptest.NewRecorder()

		handler.List(listW, listReq)

		if listW.Code != http.StatusOK {
			t.Fatalf("❌ Failed to list users: expected 200, got %d", listW.Code)
		}

		// 4. Verificar se a resposta contém usuários
		var users []user.User
		err := json.Unmarshal(listW.Body.Bytes(), &users)
		if err != nil {
			t.Fatalf("❌ Failed to unmarshal response: %v", err)
		}

		if len(users) == 0 {
			t.Fatalf("❌ Expected users in list, got empty list")
		}

		fmt.Printf("✅ User found in list: %s\n", users[0].Email)
	})
}

// E2E_MultipleUsersCreation testa a criação de múltiplos usuários com SQS real
func TestE2E_MultipleUsersCreation(t *testing.T) {
	handler, service, _, sqsConsumer := setupTestE2E(t)

	usersData := []map[string]string{
		{
			"name":     "Alice Silva",
			"email":    "alice@test.com",
			"document": "12345678901",
			"phone":    "5511912345678",
		},
		{
			"name":     "Bob Santos",
			"email":    "bob@test.com",
			"document": "12345678902",
			"phone":    "5511912345679",
		},
		{
			"name":     "Carol Costa",
			"email":    "carol@test.com",
			"document": "12345678903",
			"phone":    "5511912345680",
		},
	}

	// 1. Criar múltiplos usuários
	for _, userData := range usersData {
		jsonBody, _ := json.Marshal(userData)
		req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.Create(w, req)

		if w.Code != http.StatusAccepted {
			t.Fatalf("❌ Failed to create user %s: expected 202, got %d", userData["email"], w.Code)
		}
	}

	// 2. Consumir e processar mensagens da fila SQS real
	ctx := context.Background()
	consumeMessages(ctx, t, sqsConsumer, service, 3)

	// 3. Contar usuários
	countReq := httptest.NewRequest(http.MethodGet, "/usersCount", nil)
	countW := httptest.NewRecorder()

	handler.Count(countW, countReq)

	if countW.Code != http.StatusOK {
		t.Fatalf("❌ Failed to count users: expected 200, got %d", countW.Code)
	}

	// 4. Verificar a contagem
	var countResponse struct {
		Total int `json:"total"`
	}
	err := json.Unmarshal(countW.Body.Bytes(), &countResponse)
	if err != nil {
		t.Fatalf("❌ Failed to unmarshal count response: %v", err)
	}

	if countResponse.Total < 3 {
		t.Fatalf("❌ Expected at least 3 users, got %d", countResponse.Total)
	}

	fmt.Printf("✅ Successfully created and counted %d users\n", countResponse.Total)
}

// E2E_UserCreationWithListAndCount testa o fluxo completo
// criação -> listagem -> contagem com SQS real
func TestE2E_UserCreationWithListAndCount(t *testing.T) {
	handler, service, _, sqsConsumer := setupTestE2E(t)

	userBody := map[string]string{
		"name":     "Maria Santos",
		"email":    "maria@test.com",
		"document": "98765432100",
		"phone":    "5511998765432",
	}

	// 1. Criar usuário
	jsonBody, _ := json.Marshal(userBody)
	createReq := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonBody))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()

	handler.Create(createW, createReq)

	if createW.Code != http.StatusAccepted {
		t.Fatalf("❌ Create failed: expected 202, got %d", createW.Code)
	}

	// 2. Consumir e processar mensagens da fila SQS real
	ctx := context.Background()
	consumeMessages(ctx, t, sqsConsumer, service, 1)

	// 3. Listar todos os usuários
	listReq := httptest.NewRequest(http.MethodGet, "/users", nil)
	listW := httptest.NewRecorder()

	handler.List(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("❌ List failed: expected 200, got %d", listW.Code)
	}

	var users []user.User
	err := json.Unmarshal(listW.Body.Bytes(), &users)
	if err != nil {
		t.Fatalf("❌ Failed to unmarshal users: %v", err)
	}

	// 4. Contar usuários
	countReq := httptest.NewRequest(http.MethodGet, "/usersCount", nil)
	countW := httptest.NewRecorder()

	handler.Count(countW, countReq)

	if countW.Code != http.StatusOK {
		t.Fatalf("❌ Count failed: expected 200, got %d", countW.Code)
	}

	var countResponse struct {
		Total int `json:"total"`
	}
	err = json.Unmarshal(countW.Body.Bytes(), &countResponse)
	if err != nil {
		t.Fatalf("❌ Failed to unmarshal count: %v", err)
	}

	// 5. Verificar consistência entre listagem e contagem
	if len(users) != countResponse.Total {
		t.Fatalf("❌ Inconsistency: listed %d users but count returned %d", len(users), countResponse.Total)
	}

	fmt.Printf("✅ E2E Test Passed: Created 1 user, listed %d users, count returned %d\n", len(users), countResponse.Total)
}

// E2E_ErrorHandling testa tratamento de erros no fluxo completo com SQS real
func TestE2E_ErrorHandling(t *testing.T) {
	handler, service, _, sqsConsumer := setupTestE2E(t)

	t.Run("Create invalid user then list should still work", func(t *testing.T) {
		// 1. Tentar criar usuário inválido
		invalidBody := map[string]string{
			"name":     "Invalid",
			"email":    "not-an-email",
			"document": "123",
			"phone":    "123",
		}

		jsonBody, _ := json.Marshal(invalidBody)
		invalidReq := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonBody))
		invalidReq.Header.Set("Content-Type", "application/json")
		invalidW := httptest.NewRecorder()

		handler.Create(invalidW, invalidReq)

		if invalidW.Code != http.StatusBadRequest {
			t.Fatalf("❌ Expected 400 for invalid user, got %d", invalidW.Code)
		}

		// 2. Listar usuários deve continuar funcionando
		listReq := httptest.NewRequest(http.MethodGet, "/users", nil)
		listW := httptest.NewRecorder()

		handler.List(listW, listReq)

		if listW.Code != http.StatusOK {
			t.Fatalf("❌ List should work after invalid create: expected 200, got %d", listW.Code)
		}

		fmt.Println("✅ Invalid user correctly rejected and list still works")
	})

	t.Run("Duplicate email rejection with SQS", func(t *testing.T) {
		duplicateEmail := map[string]string{
			"name":     "Duplicate Test",
			"email":    fmt.Sprintf("duplicate-%d@test.com", time.Now().UnixNano()),
			"document": "11111111111",
			"phone":    "5511999999999",
		}

		// Criar primeiro usuário
		jsonBody1, _ := json.Marshal(duplicateEmail)
		req1 := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonBody1))
		req1.Header.Set("Content-Type", "application/json")
		w1 := httptest.NewRecorder()
		handler.Create(w1, req1)

		if w1.Code != http.StatusAccepted {
			t.Fatalf("❌ First create should succeed: expected 202, got %d", w1.Code)
		}

		// Consumir mensagens
		ctx := context.Background()
		consumeMessages(ctx, t, sqsConsumer, service, 1)

		// Tentar criar segundo usuário com mesmo email
		jsonBody2, _ := json.Marshal(duplicateEmail)
		req2 := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBuffer(jsonBody2))
		req2.Header.Set("Content-Type", "application/json")
		w2 := httptest.NewRecorder()
		handler.Create(w2, req2)

		if w2.Code != http.StatusConflict {
			t.Fatalf("❌ Duplicate email should be rejected: expected 409, got %d", w2.Code)
		}

		// Verificar que usuários ainda podem ser listados
		listReq := httptest.NewRequest(http.MethodGet, "/users", nil)
		listW := httptest.NewRecorder()
		handler.List(listW, listReq)

		if listW.Code != http.StatusOK {
			t.Fatalf("❌ List should work after duplicate error: expected 200, got %d", listW.Code)
		}

		fmt.Println("✅ Duplicate email correctly rejected and list still works")
	})
}
