resource "aws_security_group" "rds" {
  name_prefix = "rds-"
  vpc_id      = var.vpc_id

  ingress {
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["10.0.0.0/16"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.environment}-rds-sg"
  }
}

resource "aws_db_instance" "postgres" {
  identifier              = "${var.environment}-postgres"
  allocated_storage       = var.allocated_storage
  storage_type            = "gp2"
  engine                  = "postgres"
  engine_version          = var.engine_version
  instance_class          = var.instance_class
  db_name                 = var.db_name
  username                = var.db_username
  password                = var.db_password
  db_subnet_group_name    = var.db_subnet_group_name
  vpc_security_group_ids  = [aws_security_group.rds.id]
  
  # Backup and maintenance (Free Tier compatible)
  backup_retention_period = 1
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"
  
  # High availability (disabled for Free Tier)
  multi_az               = false
  
  # Encryption
  storage_encrypted      = true
  
  # Monitoring (disabled to avoid extra costs on Free Tier)
  # enabled_cloudwatch_logs_exports = ["postgresql"]
  
  # Skip final snapshot for Free Tier
  skip_final_snapshot    = true
  
  tags = {
    Name = "${var.environment}-postgres"
  }
}
