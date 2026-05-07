output "queue_url" {
  description = "SQS queue URL"
  value       = aws_sqs_queue.main.url
}

output "queue_arn" {
  description = "SQS queue ARN"
  value       = aws_sqs_queue.main.arn
}
