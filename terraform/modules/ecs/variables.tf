variable "environment" {
  type = string
}

variable "cluster_name" {
  type = string
}

variable "api_container_name" {
  type = string
}

variable "api_image" {
  type = string
}

variable "api_port" {
  type = number
}

variable "worker_container_name" {
  type = string
}

variable "worker_image" {
  type = string
}

variable "api_desired_count" {
  type = number
}

variable "worker_desired_count" {
  type = number
}

variable "api_cpu" {
  type = number
}

variable "api_memory" {
  type = number
}

variable "worker_cpu" {
  type = number
}

variable "worker_memory" {
  type = number
}

variable "vpc_id" {
  type = string
}

variable "private_subnets" {
  type = list(string)
}

variable "public_subnets" {
  type = list(string)
}

variable "alb_security_group" {
  type = string
}

variable "ecs_security_group" {
  type = string
}

variable "rds_security_group" {
  type = string
}

variable "db_host" {
  type = string
}

variable "db_port" {
  type = number
}

variable "db_name" {
  type = string
}

variable "db_username" {
  type      = string
  sensitive = true
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "sqs_queue_url" {
  type = string
}

variable "sqs_queue_arn" {
  type = string
}

variable "aws_region" {
  type = string
}
