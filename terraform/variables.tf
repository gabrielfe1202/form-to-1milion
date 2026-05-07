variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-2"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "prod"
}

# VPC Configuration
variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones"
  type        = list(string)
  default     = ["us-east-2a"]
}

# RDS Configuration
variable "db_allocated_storage" {
  description = "RDS allocated storage in GB"
  type        = number
  default     = 20
}

variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t2.micro"
}

variable "db_engine_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "15"
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "form_to_1milion_db"
  sensitive   = false
}

variable "db_username" {
  description = "Database master username"
  type        = string
  default     = "postgres"
  sensitive   = true
}

variable "db_password" {
  description = "Database master password"
  type        = string
  sensitive   = true
}

# SQS Configuration
variable "queue_name" {
  description = "SQS queue name"
  type        = string
  default     = "user-create-queue"
}

variable "message_retention_seconds" {
  description = "Message retention in seconds (default: 4 days)"
  type        = number
  default     = 345600
}

variable "visibility_timeout_seconds" {
  description = "Visibility timeout in seconds"
  type        = number
  default     = 300
}

# ECS Configuration
variable "cluster_name" {
  description = "ECS cluster name"
  type        = string
  default     = "form-to-1milion"
}

variable "api_desired_count" {
  description = "Desired number of API tasks"
  type        = number
  default     = 1
}

variable "worker_desired_count" {
  description = "Desired number of Worker tasks"
  type        = number
  default     = 1
}

variable "api_cpu" {
  description = "CPU units for API task (256-4096)"
  type        = number
  default     = 512
}

variable "api_memory" {
  description = "Memory for API task in MB (512-30720)"
  type        = number
  default     = 1024
}

variable "worker_cpu" {
  description = "CPU units for Worker task (256-4096)"
  type        = number
  default     = 256
}

variable "worker_memory" {
  description = "Memory for Worker task in MB (512-30720)"
  type        = number
  default     = 512
}
