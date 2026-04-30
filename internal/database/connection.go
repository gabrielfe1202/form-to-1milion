package database

import (
	"database/sql"
	"fmt"
	"form-to-1milion/internal/utils/env"
	"log"
	"net/url"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect() *sql.DB {
	pgHost := env.Get("DB_HOST", "localhost")
	pgPort := env.Get("DB_PORT", "5432")
	pgUser := env.Get("DB_USER", "user")
	pgPassword := env.Get("DB_PASSWORD", "password")
	pgDB := env.Get("DB_NAME", "form_to_1milion_db")
	pgSSL := env.Get("DB_SSLMODE", "disable")

	// URL encode the password to handle special characters
	encodedPassword := url.QueryEscape(pgPassword)

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		pgUser, encodedPassword, pgHost, pgPort, pgDB, pgSSL,
	)

	var db *sql.DB
	var err error

	maxRetries := 10
	baseDelay := 2 * time.Second

	for i := 1; i <= maxRetries; i++ {
		log.Printf("Tentando conectar ao banco (tentativa %d/%d)...", i, maxRetries)

		db, err = sql.Open("pgx", dsn)
		if err != nil {
			log.Printf("Erro ao abrir conexão: %v", err)
		} else {
			err = db.Ping()
			if err == nil {
				log.Println("Conectado ao banco com sucesso ✅")

				// Pool de conexões OTIMIZADO para alta concorrência
				db.SetMaxOpenConns(100) // Até 100 conexões simultâneas
				db.SetMaxIdleConns(20)  // Manter 20 conexões idle
				db.SetConnMaxIdleTime(5 * time.Minute)
				db.SetConnMaxLifetime(30 * time.Minute)

				return db
			}

			log.Printf("Banco ainda não está pronto: %v", err)
		}

		sleep := time.Duration(i) * baseDelay
		log.Printf("Aguardando %s antes da próxima tentativa...", sleep)
		time.Sleep(sleep)
	}

	log.Fatal("Não foi possível conectar ao banco após várias tentativas")
	return nil
}
func RunMigrations(db *sql.DB) error {
	// Criar tabela de usuários
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			document TEXT,
			phone TEXT
		);
	`)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela users: %w", err)
	}

	// Criar índices para melhorar performance de queries
	indices := []string{
		`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,
		`CREATE INDEX IF NOT EXISTS idx_users_document ON users(document);`,
	}

	for _, indexSQL := range indices {
		if _, err := db.Exec(indexSQL); err != nil {
			log.Printf("⚠️  Aviso ao criar índice: %v", err)
			// Não retornar erro aqui pois o índice pode já existir
		}
	}

	log.Println("✅ Migrações executadas com sucesso")
	return nil
}
