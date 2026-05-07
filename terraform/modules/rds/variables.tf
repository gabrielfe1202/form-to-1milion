variable "environment" {
  type = string
}

variable "vpc_id" {
  type = string
}

variable "db_subnet_group_name" {
  type = string
}

variable "allocated_storage" {
  type = number
}

variable "instance_class" {
  type = string
}

variable "engine_version" {
  type = string
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
