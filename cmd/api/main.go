package main

import (
	"compress/gzip"
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"

	"form-to-1milion/internal/database"
	"form-to-1milion/internal/producer"
	"form-to-1milion/internal/user"
	"form-to-1milion/internal/utils/env"
)

// GzipResponseWriter wraps http.ResponseWriter para adicionar compressão gzip
type GzipResponseWriter struct {
	io.Writer
	http.ResponseWriter
}

func (w *GzipResponseWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

// Middleware para compressão GZIP
type GzipMiddleware struct {
	handler http.Handler
}

func (m *GzipMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		m.handler.ServeHTTP(w, r)
		return
	}

	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Vary", "Accept-Encoding")
	gz := gzip.NewWriter(w)
	defer gz.Close()

	m.handler.ServeHTTP(&GzipResponseWriter{Writer: gz, ResponseWriter: w}, r)
}

func main() {
	log.Println("🚀 Iniciando API...")

	// Conectar ao banco de dados
	db := database.Connect()
	defer db.Close()
	database.RunMigrations(db)

	repo := user.NewRepository(db)
	service := user.NewService(repo)

	sqsQueueURL := env.Get("SQS_QUEUE_URL", "http://localhost:4566/000000000000/user-create-queue")
	awsRegion := env.Get("AWS_REGION", "us-east-2")

	// Configurar AWS SDK com retry automático
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		log.Fatalf("❌ Falha ao carregar AWS config: %v", err)
	}

	var sqsClient *sqs.Client

	if env.Get("ENVIRONMENT", "prod") == "local" {
		sqsClient = sqs.NewFromConfig(cfg, func(opts *sqs.Options) {
			opts.BaseEndpoint = aws.String("http://local.stack:4566")

			opts.HTTPClient = &http.Client{
				Timeout: 30 * time.Second,
				Transport: &http.Transport{
					MaxIdleConns:        200,
					MaxIdleConnsPerHost: 100,
					IdleConnTimeout:     90 * time.Second,
				},
			}
		})
	} else {
		sqsClient = sqs.NewFromConfig(cfg, func(opts *sqs.Options) {
			opts.BaseEndpoint = aws.String("http://localstack:4566")

			opts.HTTPClient = &http.Client{
				Timeout: 30 * time.Second,
				Transport: &http.Transport{
					MaxIdleConns:        200,
					MaxIdleConnsPerHost: 100,
					IdleConnTimeout:     90 * time.Second,
				},
			}
		})
	}

	sqsProducer, err := producer.NewSQSProducer(sqsClient, sqsQueueURL)
	if err != nil {
		log.Println(sqsQueueURL)
		log.Fatalf("❌ Falha ao criar SQS producer: %v", err)
	}
	defer func() {
		if err := sqsProducer.Close(); err != nil {
			log.Printf("❌ Erro ao fechar SQS producer: %v", err)
		}
	}()

	// Criar handler com task queue assíncrona
	handler := user.NewHandler(service, sqsProducer)

	// Registrar rotas
	mux := http.NewServeMux()

	// Health check endpoint para load balancer
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("/user", handler.Create)
	mux.HandleFunc("/users", handler.List)
	mux.HandleFunc("/usersCount", handler.Count)

	// Aplicar middleware de compressão GZIP
	gzipHandler := &GzipMiddleware{handler: mux}

	// Configurar servidor HTTP com timeouts e buffers otimizados
	server := &http.Server{
		Addr:           ":8080",
		Handler:        gzipHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    120 * time.Second,
		MaxHeaderBytes: 1 << 16, // 64KB

	}

	// Canal para erros do servidor
	serverErrors := make(chan error, 1)

	// Iniciar servidor em goroutine
	go func() {
		log.Println("✅ Servidor rodando em http://localhost:8080")
		serverErrors <- server.ListenAndServe()
	}()

	// Esperar por sinais de shutdown (SIGINT, SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		log.Printf("📍 Sinal recebido: %v. Iniciando shutdown gracioso...", sig)
	case err := <-serverErrors:
		if err != http.ErrServerClosed {
			log.Fatalf("❌ Erro do servidor: %v", err)
		}
	}

	// Shutdown com timeout de 30 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown do servidor HTTP
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("❌ Erro durante shutdown do servidor: %v", err)
	}
	log.Println("👋 Servidor finalizado com sucesso")
}
