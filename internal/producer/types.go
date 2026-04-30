package producer

import (
	"context"
	"time"
)

// Task representa uma tarefa a ser enfileirada
type Task struct {
	ID      string      `json:"id"`
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
	Created time.Time   `json:"created"`
}

// TaskProducer é a interface para produzir tarefas
type TaskProducer interface {
	EnqueueTask(ctx context.Context, task Task) error
	Close() error
}
