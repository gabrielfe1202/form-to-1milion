# VPC Outputs
output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "vpc_cidr" {
  description = "VPC CIDR block"
  value       = module.vpc.vpc_cidr
}

# RDS Outputs
output "db_endpoint" {
  description = "RDS database endpoint"
  value       = module.rds.db_endpoint
}

output "db_port" {
  description = "RDS database port"
  value       = module.rds.db_port
}

output "db_name" {
  description = "Database name"
  value       = var.db_name
}

# SQS Outputs
output "sqs_queue_url" {
  description = "SQS queue URL"
  value       = module.sqs.queue_url
}

output "sqs_queue_arn" {
  description = "SQS queue ARN"
  value       = module.sqs.queue_arn
}

# ECR Outputs
output "api_repository_url" {
  description = "ECR API repository URL"
  value       = module.ecr.api_repository_url
}

output "worker_repository_url" {
  description = "ECR Worker repository URL"
  value       = module.ecr.worker_repository_url
}

# ECS Outputs
output "ecs_cluster_name" {
  description = "ECS cluster name"
  value       = module.ecs.cluster_name
}

output "api_load_balancer_dns" {
  description = "API load balancer DNS name"
  value       = module.ecs.load_balancer_dns
}

output "api_service_name" {
  description = "API ECS service name"
  value       = module.ecs.api_service_name
}

output "worker_service_name" {
  description = "Worker ECS service name"
  value       = module.ecs.worker_service_name
}
