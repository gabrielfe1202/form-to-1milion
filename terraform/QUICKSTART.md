# Quick Start - Deploy form-to-1milion to AWS

Complete step-by-step guide to deploy your Go application to AWS using Terraform.

## ⏱️ Time: ~30 minutes

## Prerequisites Checklist

- [ ] AWS Account with permissions
- [ ] AWS CLI installed and configured: `aws --version`
- [ ] Terraform installed: `terraform --version` (v1.0+)
- [ ] Docker installed: `docker --version`
- [ ] Go 1.24+ (for building): `go version`
- [ ] Git configured

Verify everything:

```powershell
aws --version
terraform --version
docker --version
go version
```

## Step 1: Store Database Password in AWS Secrets Manager

```powershell
# Generate a secure password or use your own
$password = "YourSecurePasswordHere123!"

# Store in SSM Parameter Store
aws ssm put-parameter `
  --name "/prod/db/password" `
  --value $password `
  --type "SecureString" `
  --overwrite `
  --region us-east-2

# Verify it was stored
aws ssm get-parameter --name "/prod/db/password" --region us-east-2
```

Alternatively, store in Secrets Manager:

```powershell
aws secretsmanager create-secret `
  --name form-to-1milion/db/password `
  --secret-string $password `
  --region us-east-2
```

## Step 2: Create terraform.tfvars

From the `terraform/` directory:

```powershell
cd terraform

# Copy the example file
Copy-Item terraform.tfvars.example terraform.tfvars

# Edit to customize (optional)
# Important variables to check:
# - aws_region: us-east-2 (or your region)
# - environment: prod
# - db_allocated_storage: 20 (adjust as needed)
# - api_desired_count: 2 (number of API instances)
# - worker_desired_count: 2 (number of worker instances)
```

Content of `terraform.tfvars`:

```hcl
aws_region = "us-east-2"
environment = "prod"

# VPC Configuration
vpc_cidr           = "10.0.0.0/16"
availability_zones = ["us-east-2a", "us-east-2b", "us-east-2c"]

# RDS Configuration
db_allocated_storage = 100
db_instance_class   = "db.t3.small"
db_engine_version   = "15"
db_name             = "form_to_1milion_db"

# SQS Configuration
queue_name = "minha-fila"

# ECS Configuration - Scaling
cluster_name         = "form-to-1milion"
api_desired_count    = 2
worker_desired_count = 2

# Task Resources
api_cpu            = 512
api_memory         = 1024
worker_cpu         = 256
worker_memory      = 512
```

## Step 3: Build and Push Docker Images to ECR

Make sure you're still in the `terraform/` directory.

### On Windows (PowerShell):

```powershell
# Make script executable (one time)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Run the build script
.\scripts\build-and-push.ps1 -Environment prod -Region us-east-2
```

### On macOS/Linux (Bash):

```bash
# Make script executable (one time)
chmod +x scripts/build-and-push.sh

# Run the build script
bash scripts/build-and-push.sh prod us-east-2
```

The script will:

1. ✅ Authenticate Docker with ECR
2. ✅ Build the API Docker image
3. ✅ Push API image to ECR
4. ✅ Build the Worker Docker image
5. ✅ Push Worker image to ECR

The script creates both `:latest` and timestamped tags for easy rollback.

Expected output:

```
API Image: 123456789012.dkr.ecr.us-east-2.amazonaws.com/prod/form-to-1milion-api:latest
Worker Image: 123456789012.dkr.ecr.us-east-2.amazonaws.com/prod/form-to-1milion-worker:latest
```

## Step 4: Initialize Terraform

From the `terraform/` directory:

```powershell
terraform init
```

This downloads the AWS provider and sets up your Terraform workspace.

## Step 5: Review Terraform Plan

```powershell
terraform plan -out=tfplan
```

This will show you:

- [ ] 3-5 subnets (public and private)
- [ ] 1 RDS PostgreSQL database
- [ ] 1 SQS queue
- [ ] 1 ECS cluster
- [ ] 2 ECS services (API and Worker)
- [ ] 1 Application Load Balancer
- [ ] Security groups, IAM roles, CloudWatch logs, etc.

**Review carefully** before proceeding!

## Step 6: Apply Infrastructure

```powershell
terraform apply tfplan
```

This will create all AWS resources. **Wait 5-15 minutes** for completion, especially for RDS.

You'll see progress:

```
aws_vpc.main: Creating...
aws_internet_gateway.main: Creating...
aws_db_instance.postgres: Creating... (this takes 5-10 minutes)
...
Apply complete! Resources: 45 added...
```

## Step 7: Get Your Application URL

```powershell
# See all outputs
terraform output

# Get just the API URL
terraform output api_load_balancer_dns
```

Copy the DNS name (e.g., `api-abc123.us-east-2.elb.amazonaws.com`).

## Step 8: Test Your Application

```powershell
# Replace with your actual DNS name
$URL = "http://api-abc123.us-east-2.elb.amazonaws.com"

# Test the API
curl $URL
curl "$URL/user"  # or whatever endpoints you have
```

If you get a 502 Bad Gateway, wait a few minutes for ECS tasks to start.

## Step 9: Monitor Your Services

### Check ECS Services Status

```powershell
# Get cluster name
$ClusterName = (terraform output -raw ecs_cluster_name)

# List services
aws ecs list-services --cluster $ClusterName

# Check service status
aws ecs describe-services --cluster $ClusterName --services form-to-1milion-api-service
```

### View CloudWatch Logs

```powershell
# API logs (real-time)
aws logs tail /ecs/prod/api --follow

# Worker logs (real-time)
aws logs tail /ecs/prod/worker --follow
```

### Check SQS Queue

```powershell
$QueueUrl = (terraform output -raw sqs_queue_url)

# Get queue attributes
aws sqs get-queue-attributes --queue-url $QueueUrl --attribute-names All

# Get message count
aws sqs get-queue-attributes --queue-url $QueueUrl --attribute-names ApproximateNumberOfMessages
```

## Step 10: Scale Services (Optional)

Edit `terraform.tfvars` to change desired counts:

```hcl
api_desired_count    = 5    # Increase from 2 to 5
worker_desired_count = 10   # Increase from 2 to 10
```

Then apply:

```powershell
terraform plan -out=tfplan
terraform apply tfplan
```

## 🎉 Deployment Complete!

Your application is now running on AWS!

### What You Have

- ✅ **VPC**: Isolated network with 3 AZs for high availability
- ✅ **RDS PostgreSQL**: Multi-AZ database with automatic backups
- ✅ **SQS Queue**: For async tasks
- ✅ **2 ECS Services**: API (front-end) + Worker (processing)
- ✅ **ALB**: Load balancer distributing traffic
- ✅ **CloudWatch**: Logs for monitoring

### Next Steps

1. **HTTPS**: Add SSL certificate to ALB (ACM)
2. **Custom Domain**: Point your domain to ALB
3. **Auto-scaling**: Set up CloudWatch alarms
4. **Monitoring**: Set up dashboards and alerts
5. **Backups**: Configure RDS backup policies
6. **CI/CD**: Automate deployments

## 📊 Estimated Costs (Monthly)

| Service | Cost |
|---------|------|
| RDS (db.t3.small) | ~$30 |
| ECS Fargate (4 tasks) | ~$60 |
| ALB | ~$16 |
| Data Transfer | ~$20 |
| SQS | ~$5 |
| ECR | ~$5 |
| **Total** | **~$136/month** |

*Costs vary by region and usage. Use AWS Calculator for accurate estimates.*

## ⚠️ Troubleshooting

### Issue: ECS Tasks Won't Start

```powershell
# Check task logs
aws logs tail /ecs/prod/api --follow

# Check task details
aws ecs describe-tasks --cluster $ClusterName --tasks <task-arn>
```

### Issue: Database Connection Failed

```powershell
# Verify RDS is running
aws rds describe-db-instances --query "DBInstances[0].[DBInstanceIdentifier,DBInstanceStatus]"

# Check security groups
aws ec2 describe-security-groups --filters "Name=tag:Name,Values=*rds*"
```

### Issue: Images Not in ECR

```powershell
# List ECR repositories
aws ecr describe-repositories

# List images in repository
aws ecr describe-images --repository-name prod/form-to-1milion-api
```

### Issue: ALB Not Responding

```powershell
# Check ALB status
aws elbv2 describe-load-balancers --query LoadBalancers[].LoadBalancerArn

# Check target group health
aws elbv2 describe-target-health --target-group-arn <tg-arn>
```

## 🧹 Cleanup

To destroy everything and save costs:

```powershell
terraform destroy
```

**Warning**: This deletes your database! A final snapshot is saved.

## 📚 Additional Resources

- [Terraform Documentation](https://www.terraform.io/docs)
- [AWS Documentation](https://docs.aws.amazon.com/)
- [ECS Best Practices](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/best_practices.html)
- [RDS Best Practices](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_BestPractices.html)

## 🆘 Need Help?

1. Check CloudWatch Logs: `aws logs tail /ecs/prod/api --follow`
2. Review Terraform state: `terraform state list`
3. Check AWS Console: https://console.aws.amazon.com
4. Review this guide: `README.md`

---

**Happy Deploying! 🚀**
