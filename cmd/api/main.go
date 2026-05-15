package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"form-to-1milion/internal/database"
	"form-to-1milion/internal/observability"
	"form-to-1milion/internal/producer"
	"form-to-1milion/internal/user"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize OpenTelemetry tracer
	tp, err := observability.InitTracer(ctx, "form-to-1milion-api")
	if err != nil {
		log.Printf("Warning: Failed to initialize tracer: %v", err)
	} else {
		defer func() {
			if err := observability.ShutdownTracer(ctx, tp); err != nil {
				log.Printf("Error shutting down tracer: %v", err)
			}
		}()
	}

	// Initialize metrics
	metrics := observability.NewMetrics()

	db := database.Connect()
	database.RunMigrations(db)

	repo := user.NewRepository(db)
	service := user.NewService(repo)

	rabbitMQProducer, err := producer.NewRabbitMQProducer("amqp://user:password@rabbitmq:5672/", "users")
	if err != nil {
		log.Fatalf("falha ao criar RabbitMQ producer: %v", err)
	}
	defer func() {
		if err := rabbitMQProducer.Close(); err != nil {
			log.Printf("erro ao fechar RabbitMQ producer: %v", err)
		}
	}()

	handler := user.NewHandler(service, rabbitMQProducer, metrics)

	// Register HTTP endpoints with observability
	mux := http.NewServeMux()
	mux.HandleFunc("/user", handler.Create)
	mux.HandleFunc("/users", handler.List)
	mux.HandleFunc("/usersCount", handler.Count)

	// Register Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Wrap the mux with observability middleware
	wrappedMux := observability.WrapHTTPHandler(mux, metrics)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: wrappedMux,
	}

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*60) // 5 minutes timeout
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
	}()

	fmt.Println("Server running at http://localhost:8080")
	fmt.Println("Metrics available at http://localhost:8080/metrics")
	log.Fatal(srv.ListenAndServe())
}
