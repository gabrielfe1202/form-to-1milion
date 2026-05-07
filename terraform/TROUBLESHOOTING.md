# Troubleshooting Guide

Common issues and solutions when deploying form-to-1milion to AWS.

## Table of Contents

1. [Setup Issues](#setup-issues)
2. [Terraform Issues](#terraform-issues)
3. [ECS Issues](#ecs-issues)
4. [Database Issues](#database-issues)
5. [Application Issues](#application-issues)
6. [Networking Issues](#networking-issues)

## Setup Issues

### ❌ AWS CLI not found

**Error:**
```
'aws' is not recognized as an internal or external command
```

**Solution:**
1. Install AWS CLI: https://aws.amazon.com/cli/
2. Verify: `aws --version`
3. Configure credentials: `aws configure`

### ❌ Terraform not found

**Error:**
```
'terraform' is not recognized as an internal or external command
```

**Solution:**
1. Install Terraform: https://www.terraform.io/downloads
2. Add to PATH
3. Verify: `terraform --version`

### ❌ Docker not found

**Error:**
```
'docker' is not recognized as an internal or external command
```

**Solution:**
1. Install Docker Desktop: https://www.docker.com/products/docker-desktop
2. Start Docker service
3. Verify: `docker --version`

### ❌ AWS credentials not configured

**Error:**
```
The AWS Access Key ID and Secret Access Key provided could not be validated.
```

**Solution:**
```bash
# Configure AWS credentials
aws configure

# Or set environment variables
$env:AWS_ACCESS_KEY_ID = "your-access-key"
$env:AWS_SECRET_ACCESS_KEY = "your-secret-key"
$env:AWS_REGION = "us-east-2"
```

### ❌ Database password not stored in SSM

**Error:**
```
InvalidParameterException: An error occurred when calling the PutParameter operation
```

**Solution:**
```bash
# Create parameter
aws ssm put-parameter \
  --name "/prod/db/password" \
  --value "your-password" \
  --type "SecureString" \
  --region us-east-2

# Verify
aws ssm get-parameter --name "/prod/db/password" --region us-east-2
```

## Terraform Issues

### ❌ Terraform init fails

**Error:**
```
Error: Failed to download module
```

**Solution:**
1. Check internet connection
2. Delete `.terraform/` and try again: `rm -r .terraform/`
3. Clear cache: `terraform clearcache`
4. Run init again: `terraform init`

### ❌ Invalid provider version

**Error:**
```
Error: Required version constraint not met
```

**Solution:**
1. Update Terraform: Download latest version
2. Or adjust `required_version` in `main.tf`
3. Run `terraform init` again

### ❌ Variables not set

**Error:**
```
Error: No value for required variable "db_password"
```

**Solution:**
1. Create `terraform.tfvars` file:
   ```
   cp terraform.tfvars.example terraform.tfvars
   ```
2. Edit and set all variables
3. Or pass variables: `terraform apply -var="db_password=xxx"`

### ❌ Terraform state error

**Error:**
```
Error: Error reading state for aws_instance.example in module.vpc: resource not found
```

**Solution:**
```bash
# Refresh state
terraform refresh

# Check state
terraform state list

# Remove bad state
terraform state rm aws_instance.example
```

### ❌ Plan shows 0 changes when changes expected

**Error:**
```
No changes. Your infrastructure matches the configuration.
```

**Solution:**
1. Check `terraform.tfvars` is loaded
2. Run `terraform plan` with explicit var file:
   ```bash
   terraform plan -var-file="terraform.tfvars"
   ```
3. Check for missing variables

## ECS Issues

### ❌ ECS tasks won't start

**Symptoms:**
- Tasks show `FAILED` or `STOPPED`
- ALB shows unhealthy targets
- 502 Bad Gateway errors

**Symptoms:**
```bash
# Check cluster
aws ecs list-clusters

# Check services
aws ecs list-services --cluster form-to-1milion

# Get cluster name from terraform
$CLUSTER = (terraform output -raw ecs_cluster_name)

# List tasks
aws ecs list-tasks --cluster $CLUSTER

# Check task details
aws ecs describe-tasks --cluster $CLUSTER --tasks <task-arn>
```

**Common causes:**

1. **Image not found in ECR:**
   ```bash
   # Build and push images
   ./scripts/build-and-push.ps1
   
   # Verify image exists
   aws ecr describe-images --repository-name prod/form-to-1milion-api
   ```

2. **Invalid container configuration:**
   ```bash
   # Check logs
   aws logs tail /ecs/prod/api --follow
   ```

3. **Environment variables missing:**
   - Verify `DB_HOST`, `DB_NAME`, `SQS_QUEUE_URL` are set
   - Check `/ecs/prod/api` logs

4. **Database not reachable:**
   - Check RDS is running: `aws rds describe-db-instances`
   - Check security groups allow port 5432
   - Test connection from EC2 instance

5. **Insufficient resources:**
   - Increase `api_cpu` or `api_memory` in `terraform.tfvars`
   - Reapply: `terraform apply`

### ❌ ALB shows 502 Bad Gateway

**Cause:**
- ECS tasks are not healthy

**Solution:**
```bash
# 1. Check target group health
$TG_ARN = (aws elbv2 describe-target-groups \
  --query "TargetGroups[0].TargetGroupArn" --output text)

aws elbv2 describe-target-health --target-group-arn $TG_ARN

# 2. If targets are unhealthy, check ECS tasks
aws ecs list-tasks --cluster form-to-1milion

# 3. View task logs
aws logs tail /ecs/prod/api --follow

# 4. Wait 1-2 minutes for tasks to fully boot
# 5. Refresh browser
```

### ❌ Container keeps restarting

**Symptoms:**
- Task stays in `PROVISIONING` state
- Container exits immediately after start

**Solution:**
```bash
# 1. Check logs for errors
aws logs tail /ecs/prod/api --follow

# 2. Common errors:
# - "database connection refused" → Check RDS security group
# - "Cannot pull image" → Check ECR image exists
# - "Environment variable not found" → Check task definition

# 3. Fix the issue and rebuild/push image
./scripts/build-and-push.ps1

# 4. Force task replacement
aws ecs update-service \
  --cluster form-to-1milion \
  --service form-to-1milion-api-service \
  --force-new-deployment
```

## Database Issues

### ❌ Cannot connect to RDS

**Error:**
```
could not connect to server: Connection refused
```

**Solution:**

1. **Verify RDS is running:**
   ```bash
   aws rds describe-db-instances \
     --query "DBInstances[0].[DBInstanceIdentifier,DBInstanceStatus]"
   ```

2. **Check security group:**
   ```bash
   # Get security group
   $SG = (aws rds describe-db-instances \
     --query "DBInstances[0].VpcSecurityGroups[0].VpcSecurityGroupId" \
     --output text)
   
   # Check inbound rules (should allow port 5432)
   aws ec2 describe-security-groups --group-ids $SG
   ```

3. **Test from ECS task:**
   ```bash
   # SSH into bastion/instance
   # Install psql: apt-get install postgresql-client
   # Connect: psql -h <db-endpoint> -U postgres -d form_to_1milion_db
   ```

### ❌ Database creation failed

**Error:**
```
Error creating DB instance: InvalidParameterValue
```

**Solution:**
1. Check instance class is valid: `db.t3.micro`, `db.t3.small`, etc.
2. Check storage type: must be `gp2` for general purpose
3. Check password minimum requirements (8 characters, mixed case)
4. Check availability zones are valid for region

### ❌ RDS backup failed

**Error:**
```
Error: Backup operation failed
```

**Solution:**
1. Check database has enough free storage
2. Check backup window (3-4 AM UTC)
3. Manual backup:
   ```bash
   aws rds create-db-snapshot \
     --db-instance-identifier prod-postgres \
     --db-snapshot-identifier manual-snapshot
   ```

## Application Issues

### ❌ Database migrations not running

**Symptoms:**
- Tables not found in database
- "relation does not exist" errors

**Solution:**
1. Check application startup logs:
   ```bash
   aws logs tail /ecs/prod/api --follow
   ```

2. Verify migrations are being run in code
3. Manually run migrations:
   ```bash
   # From local machine
   go run cmd/api/main.go
   # or inside container
   docker exec <container-id> go run cmd/api/main.go
   ```

### ❌ Application can't send SQS messages

**Error:**
```
AccessDenied: User not authorized to perform sqs:SendMessage
```

**Solution:**
1. Check IAM role has SQS permissions:
   ```bash
   # Get role name
   $ROLE = (aws ecs describe-task-definition \
     --task-definition form-to-1milion-api \
     --query "taskDefinition.taskRoleArn" --output text | \
     cut -d'/' -f2)
   
   # Check policies
   aws iam list-role-policies --role-name $ROLE
   ```

2. Verify SQS policy includes queue ARN

### ❌ Worker not processing messages

**Symptoms:**
- Queue has messages but nothing happening
- Worker service running but no logs

**Solution:**
```bash
# 1. Check worker is running
aws ecs list-tasks --cluster form-to-1milion --service-name form-to-1milion-worker-service

# 2. Check logs
aws logs tail /ecs/prod/worker --follow

# 3. Check queue
$QUEUE_URL = (terraform output -raw sqs_queue_url)
aws sqs receive-message --queue-url $QUEUE_URL

# 4. Check message attributes
aws sqs get-queue-attributes --queue-url $QUEUE_URL --attribute-names All

# 5. Restart worker (force new deployment)
aws ecs update-service \
  --cluster form-to-1milion \
  --service form-to-1milion-worker-service \
  --force-new-deployment
```

## Networking Issues

### ❌ Cannot reach ALB

**Error:**
```
Failed to connect to <alb-dns>
```

**Solution:**

1. **Verify ALB is running:**
   ```bash
   aws elbv2 describe-load-balancers \
     --query "LoadBalancers[0].[LoadBalancerName,State.Code]"
   ```

2. **Check security group allows port 80:**
   ```bash
   $ALB_SG = (aws elbv2 describe-load-balancers \
     --query "LoadBalancers[0].SecurityGroups[0]" \
     --output text)
   
   aws ec2 describe-security-groups --group-ids $ALB_SG
   ```

3. **Check target group configuration:**
   ```bash
   aws elbv2 describe-target-groups
   ```

4. **Check listener:**
   ```bash
   aws elbv2 describe-listeners \
     --load-balancer-arn <alb-arn>
   ```

### ❌ VPC resources can't communicate

**Symptoms:**
- Tasks can't reach database
- Tasks can't reach SQS
- Workers can't reach tasks

**Solution:**

1. **Check security groups:**
   ```bash
   # ECS tasks security group should allow RDS traffic
   aws ec2 describe-security-groups \
     --filters "Name=tag:Name,Values=*ecs*"
   ```

2. **Check NAT gateway:**
   ```bash
   # NAT should be in public subnets
   aws ec2 describe-nat-gateways
   ```

3. **Check route tables:**
   ```bash
   aws ec2 describe-route-tables
   ```

## Performance Issues

### ❌ High response times

**Symptoms:**
- Slow API responses
- High latency

**Solution:**

1. **Check database performance:**
   ```bash
   aws cloudwatch get-metric-statistics \
     --namespace AWS/RDS \
     --metric-name CPUUtilization \
     --dimensions Name=DBInstanceIdentifier,Value=prod-postgres \
     --start-time $(date -u -d '1 hour ago' +%Y-%m-%dT%H:%M:%S) \
     --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
     --period 300 \
     --statistics Average
   ```

2. **Scale up resources:**
   - Increase `api_cpu` and `api_memory`
   - Increase instances: `api_desired_count`
   - Upgrade database: `db_instance_class`

3. **Add caching:**
   - ElastiCache (Redis/Memcached)
   - CloudFront CDN

### ❌ High CPU/Memory usage

**Solution:**

1. **Identify bottleneck:**
   ```bash
   aws logs tail /ecs/prod/api --follow | grep -i error
   ```

2. **Profile application:**
   - Check for infinite loops
   - Check for DB N+1 queries
   - Check for memory leaks

3. **Scale horizontally:**
   - Increase desired count
   - ECS auto-scaling

## Checking Health

### Quick health check script

```bash
#!/bin/bash

echo "=== AWS Health Check ==="

# 1. ECS Cluster
echo "ECS Cluster:"
aws ecs describe-clusters --clusters form-to-1milion \
  --query "clusters[0].[clusterName,status]"

# 2. Services
echo "ECS Services:"
aws ecs list-services --cluster form-to-1milion

# 3. RDS
echo "RDS Database:"
aws rds describe-db-instances \
  --query "DBInstances[0].[DBInstanceIdentifier,DBInstanceStatus]"

# 4. ALB
echo "Load Balancer:"
aws elbv2 describe-load-balancers \
  --query "LoadBalancers[0].[LoadBalancerName,State.Code]"

# 5. SQS
echo "SQS Queue:"
aws sqs get-queue-attributes \
  --queue-url $(terraform output -raw sqs_queue_url) \
  --attribute-names ApproximateNumberOfMessages

echo "=== End Health Check ==="
```

## Getting Help

1. Check CloudWatch Logs: `aws logs tail /ecs/prod/api --follow`
2. Review AWS Console: https://console.aws.amazon.com
3. Check Terraform plan: `terraform plan`
4. Search AWS documentation
5. Ask in AWS forums or Stack Overflow

---

**Last Updated:** 2024-04-01
