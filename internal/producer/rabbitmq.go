package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type RabbitMQProducer struct {
	conn      *amqp091.Connection
	channel   *amqp091.Channel
	queueName string
}

func NewRabbitMQProducer(amqpURL, queueName string) (*RabbitMQProducer, error) {
	var conn *amqp091.Connection
	var ch *amqp091.Channel
	var err error

	maxRetries := 10
	baseDelay := 2 * time.Second

	for i := 1; i <= maxRetries; i++ {
		fmt.Printf("Conectando ao RabbitMQ (tentativa %d/%d)...", i, maxRetries)

		conn, err = amqp091.Dial(amqpURL)
		if err != nil {
			fmt.Printf("Erro ao conectar: %v", err)
		} else {
			ch, err = conn.Channel()
			if err != nil {
				fmt.Printf("Erro ao abrir canal: %v", err)
				_ = conn.Close()
			} else {
				_, err = ch.QueueDeclare(
					queueName,
					true,
					false,
					false,
					false,
					nil,
				)
				if err != nil {
					fmt.Printf("Erro ao declarar fila: %v", err)
					_ = ch.Close()
					_ = conn.Close()
				} else {
					fmt.Println("RabbitMQ conectado com sucesso ✅")
					return &RabbitMQProducer{
						conn:      conn,
						channel:   ch,
						queueName: queueName,
					}, nil
				}
			}
		}

		sleep := time.Duration(i) * baseDelay
		fmt.Printf("Aguardando %s antes da próxima tentativa...", sleep)
		time.Sleep(sleep)
	}

	return nil, fmt.Errorf("não foi possível conectar ao RabbitMQ após %d tentativas", maxRetries)
}

func (p *RabbitMQProducer) Close() error {
	var err error
	if p.channel != nil {
		if chErr := p.channel.Close(); chErr != nil {
			err = fmt.Errorf("erro ao fechar canal RabbitMQ: %w", chErr)
		}
	}
	if p.conn != nil {
		if connErr := p.conn.Close(); connErr != nil {
			if err != nil {
				err = fmt.Errorf("%v; erro ao fechar conexão RabbitMQ: %w", err, connErr)
			} else {
				err = fmt.Errorf("erro ao fechar conexão RabbitMQ: %w", connErr)
			}
		}
	}
	return err
}

func (p *RabbitMQProducer) EnqueueTask(ctx context.Context, task Task) error {
	if p == nil || p.channel == nil {
		return fmt.Errorf("rabbitmq producer não inicializado")
	}

	if task.Created.IsZero() {
		task.Created = time.Now().UTC()
	}

	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("erro ao serializar tarefa: %w", err)
	}

	return p.channel.PublishWithContext(ctx,
		"",          // exchange
		p.queueName, // routing key
		false,       // mandatory
		false,       // immediate
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   task.Created,
		},
	)
}

func (p *RabbitMQProducer) Consume(ctx context.Context, handler func(Task) error) error {
	if p == nil || p.channel == nil {
		return fmt.Errorf("rabbitmq producer não inicializado")
	}

	msgs, err := p.channel.Consume(
		p.queueName,
		"",    // consumer
		false, // autoAck
		false, // exclusive
		false, // no-local (não suportado pelo RabbitMQ)
		false, // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("erro ao iniciar consumo: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("canal de mensagens fechado")
			}

			var task Task
			if err := json.Unmarshal(msg.Body, &task); err != nil {
				_ = msg.Nack(false, false)
				continue
			}

			if err := handler(task); err != nil {
				_ = msg.Nack(false, true)
				continue
			}

			_ = msg.Ack(false)
		}
	}
}
