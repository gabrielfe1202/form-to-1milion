package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("form-to-1milion")

// TraceOperation cria um span para uma operação e registra seu tempo de execução
func TraceOperation(ctx context.Context, operationName string, fn func(context.Context) error) error {
	ctx, span := tracer.Start(ctx, operationName)
	defer span.End()

	start := time.Now()
	err := fn(ctx)
	duration := time.Since(start)

	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		span.RecordError(err)
	}

	span.SetAttributes(
		attribute.String("operation", operationName),
		attribute.String("duration_ms", duration.String()),
	)

	return err
}

// TraceUserCreation registra a criação de um usuário
func TraceUserCreation(ctx context.Context, email string) context.Context {
	_, span := tracer.Start(ctx, "create_user")
	span.SetAttributes(
		attribute.String("user.email", email),
		attribute.String("operation", "user.create"),
	)
	return ctx
}

// TraceUserList registra uma listagem de usuários
func TraceUserList(ctx context.Context, count int) {
	_, span := tracer.Start(ctx, "list_users")
	span.SetAttributes(
		attribute.Int("user.count", count),
		attribute.String("operation", "user.list"),
	)
}

// TraceUserCount registra uma contagem de usuários
func TraceUserCount(ctx context.Context, count int) {
	_, span := tracer.Start(ctx, "count_users")
	span.SetAttributes(
		attribute.Int("user.count", count),
		attribute.String("operation", "user.count"),
	)
}

// TraceDatabaseQuery rastreia uma query de banco de dados
func TraceDatabaseQuery(ctx context.Context, queryType string, query string) context.Context {
	_, span := tracer.Start(ctx, "database_query")
	span.SetAttributes(
		attribute.String("db.operation", queryType),
		attribute.String("db.statement", query),
	)
	return ctx
}

// TraceRabbitMQPublish rastreia uma publicação no RabbitMQ
func TraceRabbitMQPublish(ctx context.Context, queue string, taskID string, taskType string) context.Context {
	_, span := tracer.Start(ctx, "rabbitmq_publish")
	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.destination", queue),
		attribute.String("messaging.message_id", taskID),
		attribute.String("messaging.message_type", taskType),
	)
	return ctx
}

// TraceRabbitMQConsume rastreia um consumo do RabbitMQ
func TraceRabbitMQConsume(ctx context.Context, queue string, taskID string, taskType string) context.Context {
	_, span := tracer.Start(ctx, "rabbitmq_consume")
	span.SetAttributes(
		attribute.String("messaging.system", "rabbitmq"),
		attribute.String("messaging.source", queue),
		attribute.String("messaging.message_id", taskID),
		attribute.String("messaging.message_type", taskType),
	)
	return ctx
}

// SpanRecordError registra um erro em um span ativo
func SpanRecordError(ctx context.Context, err error, message string) {
	span := trace.SpanFromContext(ctx)
	if span != nil && span.IsRecording() {
		span.RecordError(err)
		span.SetStatus(codes.Error, message)
	}
}

// SpanAddEvent registra um evento em um span
func SpanAddEvent(ctx context.Context, eventName string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span != nil && span.IsRecording() {
		span.AddEvent(eventName, trace.WithAttributes(attrs...))
	}
}
