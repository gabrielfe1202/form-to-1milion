package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"form-to-1milion/internal/consumer"
	"form-to-1milion/internal/database"
	"form-to-1milion/internal/user"
)

func main() {
	// Criar consumer RabbitMQ
	rabbitMQConsumer, err := consumer.NewRabbitMQConsumer("amqp://user:password@rabbitmq:5672/", "users")
	if err != nil {
		log.Fatalf("falha ao criar RabbitMQ consumer: %v", err)
	}
	defer func() {
		if err := rabbitMQConsumer.Close(); err != nil {
			log.Printf("erro ao fechar RabbitMQ consumer: %v", err)
		}
	}()

	// Criar contexto com cancelamento via sinais
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

	fmt.Println("Consumer iniciado, aguardando mensagens...")
	if err := rabbitMQConsumer.Consume(ctx, handler); err != nil {
		log.Printf("erro ao consumir mensagens: %v", err)
	}

	fmt.Println("Consumer finalizado")
}
