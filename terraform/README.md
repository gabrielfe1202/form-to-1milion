# Terraform Infrastructure for form-to-1milion

This Terraform configuration deploys the form-to-1milion application to AWS with:

- **VPC**: Full networking setup with public/private subnets across multiple AZs
- **RDS PostgreSQL**: Multi-AZ PostgreSQL 15 database
- **SQS**: AWS SQS queue for task processing
- **ECR**: Docker image repositories for API and Worker
- **ECS Fargate**: Container orchestration for API and Worker services
- **ALB**: Application Load Balancer for API routing
- **CloudWatch**: Logs and monitoring

## Prerequisites

1. **AWS Account** with appropriate permissions
2. **Terraform** >= 1.0
3. **AWS CLI** configured with credentials
4. **Docker** for building and pushing images
5. **Go** 1.24+ (for building the application)

## Structure

```
terraform/
├── main.tf                 # Main configuration
├── variables.tf            # Variable definitions
├── outputs.tf              # Output values
├── terraform.tfvars.example # Example values
├── modules/
│   ├── vpc/               # VPC module
│   ├── rds/               # RDS PostgreSQL module
│   ├── sqs/               # SQS queue module
│   ├── ecr/               # ECR repositories module
│   └── ecs/               # ECS cluster, services, and tasks
└── scripts/
    └── build-and-push.sh  # Script to build and push Docker images
```

## Setup Instructions

### 1. Prepare Database Password

Store your database password in SSM Parameter Store:

```bash
# Replace 'prod' with your environment and 'your-password' with a secure password
aws ssm put-parameter \
  --name "/prod/db/password" \
  --value "your-secure-password-here" \
  --type "SecureString" \
  --overwrite \
  --region us-east-2
```

### 2. Create terraform.tfvars

Copy the example file and update with your values:

```bash
cp terraform.tfvars.example terraform.tfvars
# Edit terraform.tfvars with your desired values
```

**Important**: Ensure database password is secure and stored in SSM.

### 3. Build Docker Images

```bash
# From the project root directory
bash terraform/scripts/build-and-push.sh
```

This script will:
- Authenticate with ECR
- Build API Docker image
- Build Worker Docker image
- Push both images to ECR

### 4. Initialize Terraform

```bash
cd terraform
terraform init
```

### 5. Plan Infrastructure

```bash
terraform plan -out=tfplan
```

Review the plan carefully to ensure everything looks correct.

### 6. Apply Infrastructure

```bash
terraform apply tfplan
```

This will create all AWS resources. It typically takes 5-10 minutes.

### 7. Get Outputs

```bash
terraform output
```

You'll get important information like:
- `api_load_balancer_dns`: DNS name to access your API
- `sqs_queue_url`: SQS queue URL for your application
- `db_endpoint`: Database endpoint
- `ecr_repositories`: ECR image repository URLs

## Configuration Variables

### Key Variables in `terraform.tfvars`

```hcl
# AWS Region
aws_region = "us-east-2"

# Environment
environment = "prod"

# VPC CIDR
vpc_cidr = "10.0.0.0/16"

# Database
db_allocated_storage = 100         # GB
db_instance_class   = "db.t3.small"
db_engine_version   = "15"
db_name             = "form_to_1milion_db"
db_username         = "postgres"   # Set in tfvars or environment
db_password         = "***"        # Use SSM Parameter Store!

# SQS
queue_name = "minha-fila"

# ECS
api_desired_count    = 2
worker_desired_count = 2
api_cpu              = 512
api_memory           = 1024
worker_cpu           = 256
worker_memory        = 512
```

## Environment Variables in ECS Tasks

The application containers receive these environment variables:

**Database:**
- `DB_HOST`: RDS endpoint
- `DB_PORT`: Database port (5432)
- `DB_NAME`: Database name
- `DB_USER`: Database username
- `DB_PASSWORD`: From SSM Parameter Store (secure)
- `DB_SSLMODE`: require

**AWS/SQS:**
- `SQS_QUEUE_URL`: Queue URL
- `AWS_REGION`: AWS region

## Accessing Your Application

After deployment, access your API via the ALB:

```bash
# Get the ALB DNS name
terraform output api_load_balancer_dns

# Access the API
curl http://<ALB-DNS>/
```

## Scaling

To scale the services, update `terraform.tfvars`:

```hcl
api_desired_count    = 5    # Scale to 5 instances
worker_desired_count = 10   # Scale workers to 10
```

Then apply changes:

```bash
terraform apply
```

## Database Management

Connect to your RDS database:

```bash
# Get database endpoint
terraform output db_endpoint

# Connect (requires psql client)
psql -h <db-endpoint> -U postgres -d form_to_1milion_db
```

## Monitoring

View CloudWatch logs:

```bash
# API logs
aws logs tail /ecs/prod/api --follow

# Worker logs
aws logs tail /ecs/prod/worker --follow
```

View ECS services:

```bash
# Get cluster name
terraform output ecs_cluster_name

# List services
aws ecs list-services --cluster <cluster-name>
```

## Costs

Estimated monthly costs (rough):
- **RDS db.t3.small**: ~$30
- **ECS Fargate (2 API tasks + 2 worker tasks)**: ~$50-100
- **Data transfer**: Variable
- **Other services (ALB, SQS, ECR)**: ~$20

Total: ~$100-150/month

## Cleanup

To destroy all infrastructure:

```bash
terraform destroy
```

**Warning**: This will delete the database including all data (final snapshot is saved).

## Troubleshooting

### 1. ECS Tasks Not Starting

Check CloudWatch logs:

```bash
aws logs tail /ecs/prod/api --follow
aws logs tail /ecs/prod/worker --follow
```

### 2. Database Connection Issues

Verify security groups:

```bash
aws ec2 describe-security-groups --filters "Name=tag:Name,Values=*rds*"
```

### 3. SQS Queue Not Working

Check queue attributes:

```bash
aws sqs get-queue-attributes --queue-url <queue-url> --attribute-names All
```

### 4. ECR Images Not Found

Verify images in ECR:

```bash
aws ecr describe-images --repository-name prod/form-to-1milion-api
aws ecr describe-images --repository-name prod/form-to-1milion-worker
```

## Security Best Practices

1. ✅ Database password stored in SSM Parameter Store
2. ✅ RDS encryption enabled
3. ✅ SQS encryption enabled (KMS)
4. ✅ VPC with private subnets for applications
5. ✅ Security groups restrict access
6. ✅ IAM roles with least privilege
7. ✅ ECR image scanning enabled
8. ⚠️ **TODO**: Enable HTTPS on ALB
9. ⚠️ **TODO**: Add WAF rules
10. ⚠️ **TODO**: Enable VPC Flow Logs

## Next Steps

1. Enable HTTPS on ALB with ACM certificate
2. Set up CloudFront CDN
3. Configure auto-scaling policies
4. Set up backup and disaster recovery
5. Implement monitoring and alerting
6. Set up CI/CD pipeline for automatic deployments

## Support

For issues or questions, check:
- Terraform documentation: https://www.terraform.io/docs
- AWS documentation: https://docs.aws.amazon.com
- Project README: ../README.md
