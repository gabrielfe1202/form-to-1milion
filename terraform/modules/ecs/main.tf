# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = var.cluster_name

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = {
    Name = "${var.environment}-cluster"
  }
}

resource "aws_ecs_cluster_capacity_providers" "main" {
  cluster_name = aws_ecs_cluster.main.name

  capacity_providers = ["FARGATE", "FARGATE_SPOT"]

  default_capacity_provider_strategy {
    base              = 1
    weight            = 100
    capacity_provider = "FARGATE"
  }
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "api" {
  name              = "/ecs/${var.environment}/api"
  retention_in_days = 30

  tags = {
    Name = "${var.environment}-api-logs"
  }
}

resource "aws_cloudwatch_log_group" "worker" {
  name              = "/ecs/${var.environment}/worker"
  retention_in_days = 30

  tags = {
    Name = "${var.environment}-worker-logs"
  }
}

# IAM Role for ECS Task Execution
resource "aws_iam_role" "ecs_task_execution_role" {
  name_prefix = "ecsTaskExecutionRole-"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "ecs_task_execution_role_policy" {
  role       = aws_iam_role.ecs_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# IAM Role for ECS Task (application permissions)
resource "aws_iam_role" "ecs_task_role" {
  name_prefix = "ecsTaskRole-"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

# IAM Policy for SQS access
resource "aws_iam_role_policy" "ecs_sqs_policy" {
  name_prefix = "ecs-sqs-policy-"
  role        = aws_iam_role.ecs_task_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage",
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:ChangeMessageVisibility"
        ]
        Resource = [var.sqs_queue_arn]
      }
    ]
  })
}

# IAM Policy for SSM Parameter Store (DB Password)
resource "aws_iam_role_policy" "ecs_ssm_policy" {
  name_prefix = "ecs-ssm-policy-"
  role        = aws_iam_role.ecs_task_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "ssm:GetParameter",
          "ssm:GetParameters"
        ]
        Resource = "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter/${var.environment}/db/password"
      }
    ]
  })
}

# API Task Definition
resource "aws_ecs_task_definition" "api" {
  family                   = "${var.environment}-api"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.api_cpu
  memory                   = var.api_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn

  container_definitions = jsonencode(
    [
      {
        name      = var.api_container_name
        image     = var.api_image
        essential = true
        portMappings = [
          {
            containerPort = var.api_port
            hostPort      = var.api_port
            protocol      = "tcp"
          }
        ]
        environment = [
          {
            name  = "DB_HOST"
            value = var.db_host
          },
          {
            name  = "DB_PORT"
            value = tostring(var.db_port)
          },
          {
            name  = "DB_NAME"
            value = var.db_name
          },
          {
            name  = "DB_USER"
            value = var.db_username
          },
          {
            name  = "DB_SSLMODE"
            value = "require"
          },
          {
            name  = "SQS_QUEUE_URL"
            value = var.sqs_queue_url
          },
          {
            name  = "AWS_REGION"
            value = var.aws_region
          }
        ]
        secrets = [
          {
            name      = "DB_PASSWORD"
            valueFrom = "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter/${var.environment}/db/password"
          }
        ]
        logConfiguration = {
          logDriver = "awslogs"
          options = {
            "awslogs-group"         = aws_cloudwatch_log_group.api.name
            "awslogs-region"        = var.aws_region
            "awslogs-stream-prefix" = "ecs"
          }
        }
      }
    ]
  )

  tags = {
    Name = "${var.environment}-api-task"
  }
}

# Worker Task Definition
resource "aws_ecs_task_definition" "worker" {
  family                   = "${var.environment}-worker"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.worker_cpu
  memory                   = var.worker_memory
  execution_role_arn       = aws_iam_role.ecs_task_execution_role.arn
  task_role_arn            = aws_iam_role.ecs_task_role.arn

  container_definitions = jsonencode(
    [
      {
        name      = var.worker_container_name
        image     = var.worker_image
        essential = true
        environment = [
          {
            name  = "DB_HOST"
            value = var.db_host
          },
          {
            name  = "DB_PORT"
            value = tostring(var.db_port)
          },
          {
            name  = "DB_NAME"
            value = var.db_name
          },
          {
            name  = "DB_USER"
            value = var.db_username
          },
          {
            name  = "DB_SSLMODE"
            value = "require"
          },
          {
            name  = "SQS_QUEUE_URL"
            value = var.sqs_queue_url
          },
          {
            name  = "AWS_REGION"
            value = var.aws_region
          }
        ]
        secrets = [
          {
            name      = "DB_PASSWORD"
            valueFrom = "arn:aws:ssm:${var.aws_region}:${data.aws_caller_identity.current.account_id}:parameter/${var.environment}/db/password"
          }
        ]
        logConfiguration = {
          logDriver = "awslogs"
          options = {
            "awslogs-group"         = aws_cloudwatch_log_group.worker.name
            "awslogs-region"        = var.aws_region
            "awslogs-stream-prefix" = "ecs"
          }
        }
      }
    ]
  )

  tags = {
    Name = "${var.environment}-worker-task"
  }
}

# Application Load Balancer
resource "aws_lb" "main" {
  name_prefix        = "api"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [var.alb_security_group]
  subnets            = var.public_subnets

  tags = {
    Name = "${var.environment}-alb"
  }
}

# Target Group for API
resource "aws_lb_target_group" "api" {
  name_prefix = "api"
  port        = var.api_port
  protocol    = "HTTP"
  vpc_id      = var.vpc_id
  target_type = "ip"

  health_check {
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 3
    interval            = 30
    path                = "/"
    matcher             = "200-299"
  }

  tags = {
    Name = "${var.environment}-api-tg"
  }
}

# ALB Listener
resource "aws_lb_listener" "api" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.api.arn
  }
}

# ECS Service for API
resource "aws_ecs_service" "api" {
  name            = "${var.environment}-api-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = var.api_desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnets
    security_groups  = [var.ecs_security_group]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.api.arn
    container_name   = var.api_container_name
    container_port   = var.api_port
  }

  depends_on = [
    aws_lb_listener.api
  ]

  tags = {
    Name = "${var.environment}-api-service"
  }
}

# ECS Service for Worker
resource "aws_ecs_service" "worker" {
  name            = "${var.environment}-worker-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.worker.arn
  desired_count   = var.worker_desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnets
    security_groups  = [var.ecs_security_group]
    assign_public_ip = false
  }

  tags = {
    Name = "${var.environment}-worker-service"
  }
}

# Data source for current AWS account ID
data "aws_caller_identity" "current" {}
