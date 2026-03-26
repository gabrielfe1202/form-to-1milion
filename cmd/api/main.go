package main

import (
	"fmt"
	"log"
	"net/http"

	"form-to-1milion/internal/database"
	"form-to-1milion/internal/producer"
	"form-to-1milion/internal/user"
)

func main() {
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

	handler := user.NewHandler(service, rabbitMQProducer)

	http.HandleFunc("/user", handler.Create)
	http.HandleFunc("/users", handler.List)
	http.HandleFunc("/usersCount", handler.Count)

	fmt.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
