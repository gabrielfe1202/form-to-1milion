package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"form-to-1milion/internal/consumer"
	"form-to-1milion/internal/database"
	"form-to-1milion/internal/user"
	"form-to-1milion/internal/utils/env"
)

func main() {
	sqsQueueURL := env.Get("SQS_QUEUE_URL", "http://localhost:4566/000000000000/user-create-queue")
	awsRegion := env.Get("AWS_REGION", "us-east-2")

	// Para AWS: usa IAM role do ECS automaticamente
	// Para LocalStack: define BaseEndpoint
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		log.Fatalf("falha ao carregar AWS config: %v", err)
	}

	var sqsClient *sqs.Client

	// Se está rodando em ambiente local (LocalStack), configura endpoint customizado
	if env.Get("ENVIRONMENT", "prod") == "local" {
		sqsClient = sqs.NewFromConfig(cfg, func(opts *sqs.Options) {
			opts.BaseEndpoint = aws.String("http://localhost:4566")
		})
	} else {
		// Para AWS, usa as credenciais da IAM role do ECS
		sqsClient = sqs.NewFromConfig(cfg)
	}

	sqsConsumer, err := consumer.NewSQSConsumer(sqsClient, sqsQueueURL)
	if err != nil {
		log.Fatalf("falha ao criar SQS consumer: %v", err)
	}
	defer func() {
		if err := sqsConsumer.Close(); err != nil {
			log.Printf("erro ao fechar SQS consumer: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db := database.Connect()
	database.RunMigrations(db)

	userRepo := user.NewRepository(db)
	userService := user.NewService(userRepo)

	// Capturar sinais de interrupção
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutdown signal received, stopping consumer...")
		cancel()
	}()

	// Handler para processar mensagens
	handler := func(task consumer.Task) error {

		fmt.Printf("Processando tarefa: %s\n", task.ID)

		var u user.User
		payloadBytes, err := json.Marshal(task.Payload)
		if err != nil {
			return fmt.Errorf("erro ao serializar payload temporariamente: %w", err)
		}

		if err := json.Unmarshal(payloadBytes, &u); err != nil {
			return fmt.Errorf("erro ao desserializar payload para user.User: %w", err)
		}

		if err := userService.CreateUser(u); err != nil {
			return fmt.Errorf("erro ao criar usuário: %w", err)
		}

		return nil
	}

	fmt.Println("============================================================")
	fmt.Printf("Consumer iniciado com sucesso\n")
	fmt.Printf("Queue URL: %s\n", sqsQueueURL)
	fmt.Println("Aguardando mensagens (polling a cada 20 segundos)...")
	fmt.Println("============================================================")

	if err := sqsConsumer.Consume(ctx, handler); err != nil {
		log.Printf("erro ao consumir mensagens: %v", err)
	}

	fmt.Println("\nConsumer finalizado")
}
