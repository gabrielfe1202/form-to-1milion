package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQSConsumer struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSConsumer(sqsClient *sqs.Client, queueURL string) (*SQSConsumer, error) {
	if sqsClient == nil {
		return nil, fmt.Errorf("sqs client não pode ser nil")
	}

	if queueURL == "" {
		return nil, fmt.Errorf("queue url não pode ser vazia")
	}

	// Aumentar timeout para 15 segundos para permitir carregamento de credenciais AWS
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := sqsClient.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if err != nil {
		return nil, fmt.Errorf("erro ao validar fila SQS: %w", err)
	}

	fmt.Println("SQS consumer conectado com sucesso ✅")
	return &SQSConsumer{
		client:   sqsClient,
		queueURL: queueURL,
	}, nil
}

func (c *SQSConsumer) Close() error {
	// SQS client doesn't require explicit close
	return nil
}

func (c *SQSConsumer) Consume(ctx context.Context, handler func(Task) error) error {
	if c == nil || c.client == nil {
		return fmt.Errorf("sqs consumer não inicializado")
	}

	// Configure poll settings
	waiter := 20 // Long polling timeout in seconds
	maxMessages := int32(10)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Receive messages with long polling
		output, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(c.queueURL),
			MaxNumberOfMessages: maxMessages,
			WaitTimeSeconds:     int32(waiter),
		})
		if err != nil {
			fmt.Printf("erro ao receber mensagens: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Process each message
		if len(output.Messages) > 0 {
			fmt.Printf("📨 Recebidas %d mensagens\n", len(output.Messages))
		}

		for _, msg := range output.Messages {
			var task Task
			if err := json.Unmarshal([]byte(*msg.Body), &task); err != nil {
				fmt.Printf("erro ao desserializar mensagem: %v\n", err)
				_ = c.deleteMessage(*msg.ReceiptHandle)
				continue
			}

			if err := handler(task); err != nil {
				fmt.Printf("erro ao processar mensagem: %v\n", err)
				// Requeue the message by not deleting it (SQS will make it available again)
				continue
			}

			// Delete message on success
			if err := c.deleteMessage(*msg.ReceiptHandle); err != nil {
				fmt.Printf("erro ao deletar mensagem: %v\n", err)
			} else {
				fmt.Printf("Mensagem processada e deletada\n")
			}
		}
	}
}

func (c *SQSConsumer) deleteMessage(receiptHandle string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: aws.String(receiptHandle),
	})
	if err != nil {
		return fmt.Errorf("erro ao deletar mensagem da fila: %w", err)
	}

	return nil
}
