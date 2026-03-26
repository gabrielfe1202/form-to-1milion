package database

import (
	"database/sql"
	"fmt"
	"form-to-1milion/internal/utils/env"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Connect() *sql.DB {
	pgHost := env.Get("DB_HOST")
	pgPort := env.Get("DB_PORT")
	pgUser := env.Get("DB_USER")
	pgPassword := env.Get("DB_PASSWORD")
	pgDB := env.Get("DB_NAME")
	pgSSL := env.Get("DB_SSLMODE")

	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		pgUser, pgPassword, pgHost, pgPort, pgDB, pgSSL,
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

				// Pool de conexões
				db.SetMaxOpenConns(50)
				db.SetMaxIdleConns(10)
				// db.SetConnMaxIdleTime(5 * time.Minute)
				// db.SetConnMaxLifetime(30 * time.Minute)

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
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			document TEXT,
			phone TEXT
		);
	`)
	return err
}
