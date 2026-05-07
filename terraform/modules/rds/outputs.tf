output "db_endpoint" {
  description = "RDS endpoint (host only)"
  value       = split(":", aws_db_instance.postgres.endpoint)[0]
}

output "db_port" {
  description = "RDS port"
  value       = aws_db_instance.postgres.port
}

output "security_group_id" {
  description = "RDS security group ID"
  value       = aws_security_group.rds.id
}
