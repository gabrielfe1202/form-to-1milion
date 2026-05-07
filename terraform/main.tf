terraform {
  

  # Descomente para usar S3 backend
  # backend "s3" {
  #   bucket         = "form-to-1milion-terraform"
  #   key            = "prod/terraform.tfstate"
  #   region         = "us-east-2"
  #   encrypt        = true
  #   dynamodb_table = "terraform-locks"
  # }
}

provider "aws" {
  region = var.aws_region

  default_tags {
    tags = {
      Environment = var.environment
      Project     = "form-to-1milion"
      CreatedBy   = "Terraform"
    }
  }
}

# VPC
module "vpc" {
  source = "./modules/vpc"

  environment       = var.environment
  vpc_cidr          = var.vpc_cidr
  availability_zones = var.availability_zones
}

# ECR
module "ecr" {
  source = "./modules/ecr"

  environment = var.environment
}

# RDS
module "rds" {
  source = "./modules/rds"

  environment            = var.environment
  allocated_storage      = var.db_allocated_storage
  instance_class         = var.db_instance_class
  engine_version         = var.db_engine_version
  db_name                = var.db_name
  db_username            = var.db_username
  db_password            = var.db_password
  vpc_id                 = module.vpc.vpc_id
  db_subnet_group_name   = module.vpc.db_subnet_group_name
}

# SQS
module "sqs" {
  source = "./modules/sqs"

  environment  = var.environment
  queue_name   = var.queue_name
  message_retention_seconds = var.message_retention_seconds
  visibility_timeout_seconds = var.visibility_timeout_seconds
}

# ECS
module "ecs" {
  source = "./modules/ecs"

  environment                = var.environment
  cluster_name              = var.cluster_name
  api_container_name        = "form-to-1milion-api"
  api_image                 = "${module.ecr.api_repository_url}:latest"
  api_port                  = 8080
  worker_container_name     = "form-to-1milion-worker"
  worker_image              = "${module.ecr.worker_repository_url}:latest"
  api_desired_count         = var.api_desired_count
  worker_desired_count      = var.worker_desired_count
  api_cpu                   = var.api_cpu
  api_memory                = var.api_memory
  worker_cpu                = var.worker_cpu
  worker_memory             = var.worker_memory
  vpc_id                    = module.vpc.vpc_id
  private_subnets          = module.vpc.private_subnet_ids
  public_subnets           = module.vpc.public_subnet_ids
  alb_security_group       = module.vpc.alb_security_group_id
  ecs_security_group       = module.vpc.ecs_security_group_id
  rds_security_group       = module.rds.security_group_id
  
  # Database
  db_host                  = module.rds.db_endpoint
  db_port                  = module.rds.db_port
  db_name                  = var.db_name
  db_username              = var.db_username
  db_password              = var.db_password
  
  # SQS
  sqs_queue_url            = module.sqs.queue_url
  sqs_queue_arn            = module.sqs.queue_arn
  aws_region               = var.aws_region
}


data "aws_caller_identity" "current" {}

