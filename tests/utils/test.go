package utils_test

import (
	"context"
	"form-to-1milion/internal/producer"
)

type FakeProducer struct {
	Tasks []producer.Task
	Err   error
}

func (f *FakeProducer) EnqueueTask(ctx context.Context, task producer.Task) error {
	if f.Err != nil {
		return f.Err
	}
	f.Tasks = append(f.Tasks, task)
	return nil
}

func (f *FakeProducer) Close() error {
	return nil
}
