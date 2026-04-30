package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
)

type SQSProducer struct {
	client   *sqs.Client
	queueURL string
}

func NewSQSProducer(sqsClient *sqs.Client, queueURL string) (*SQSProducer, error) {
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

	fmt.Println("SQS conectado com sucesso ✅")
	return &SQSProducer{
		client:   sqsClient,
		queueURL: queueURL,
	}, nil
}

func (p *SQSProducer) Close() error {
	// SQS client doesn't require explicit close
	return nil
}

func (p *SQSProducer) EnqueueTask(ctx context.Context, task Task) error {
	if p == nil || p.client == nil {
		return fmt.Errorf("sqs producer não inicializado")
	}

	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	if task.Created.IsZero() {
		task.Created = time.Now().UTC()
	}

	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("erro ao serializar tarefa: %w", err)
	}

	// Aplicar timeout mais agressivo se o contexto recebido não tiver limite
	// Máximo 2 segundos por operação SQS
	deadline, ok := ctx.Deadline()
	if !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
	} else {
		// Garantir que o timeout não seja muito longo
		remaining := time.Until(deadline)
		if remaining > 2*time.Second {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, 2*time.Second)
			defer cancel()
		}
	}

	_, err = p.client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(p.queueURL),
		MessageBody: aws.String(string(body)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"TaskID": {
				StringValue: aws.String(task.ID),
				DataType:    aws.String("String"),
			},
			"TaskType": {
				StringValue: aws.String(task.Type),
				DataType:    aws.String("String"),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("erro ao enviar mensagem para SQS: %w", err)
	}

	return nil
}
