output "api_repository_url" {
  description = "ECR API repository URL"
  value       = aws_ecr_repository.api.repository_url
}

output "api_repository_arn" {
  description = "ECR API repository ARN"
  value       = aws_ecr_repository.api.arn
}

output "worker_repository_url" {
  description = "ECR Worker repository URL"
  value       = aws_ecr_repository.worker.repository_url
}

output "worker_repository_arn" {
  description = "ECR Worker repository ARN"
  value       = aws_ecr_repository.worker.arn
}
