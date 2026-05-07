# Architecture Overview

This document describes the AWS infrastructure created by the Terraform configuration.

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                        Internet                             │
└────────────────────────┬────────────────────────────────────┘
                         │
                    ┌────▼────┐
                    │   ALB    │
                    │ (Port 80)│
                    └────┬────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
    ┌───▼──┐         ┌───▼──┐        ┌───▼──┐
    │ API  │         │ API  │        │ API  │
    │Task 1│         │Task 2│ ...    │Task N│
    └──────┘         └──────┘        └──────┘
        │                │                │
        └────────────────┼────────────────┘
                         │
        ┌────────────────┼────────────────┐
        │                │                │
    ┌───▼──┐         ┌───▼──┐        ┌───▼──┐
    │Worker│         │Worker│        │Worker│
    │Task 1│         │Task 2│ ...    │Task N│
    └──────┘         └──────┘        └──────┘
        │                │                │
        └────────────────┼────────────────┘
                         │
        ┌────────────────▼────────────────┐
        │      RDS PostgreSQL             │
        │  (Multi-AZ, Encrypted)          │
        └──────────────────────────────────┘
        
            ┌──────────────────────────┐
            │    SQS Queue             │
            │  (Task Messages)         │
            └──────────────────────────┘
        
            ┌──────────────────────────┐
            │  ECR Repositories        │
            │  (Docker Images)         │
            └──────────────────────────┘
```

## Components

### 1. VPC (Virtual Private Cloud)

**Configuration:**
- CIDR Block: `10.0.0.0/16`
- 3 Availability Zones (AZs) for High Availability
- Public Subnets: For load balancer and NAT gateways
- Private Subnets: For ECS tasks and RDS

**Subnets:**

| Type | AZ | CIDR | Purpose |
|------|-----|------|---------|
| Public | us-east-2a | 10.0.1.0/24 | ALB, NAT Gateway |
| Public | us-east-2b | 10.0.2.0/24 | ALB, NAT Gateway |
| Public | us-east-2c | 10.0.3.0/24 | ALB, NAT Gateway |
| Private | us-east-2a | 10.0.101.0/24 | API Tasks, Worker Tasks |
| Private | us-east-2b | 10.0.102.0/24 | API Tasks, Worker Tasks |
| Private | us-east-2c | 10.0.103.0/24 | RDS Database |

**Security:**
- Internet Gateway for public subnets
- NAT Gateways for private subnet internet access
- Security Groups for ALB, ECS tasks, and RDS

### 2. Application Load Balancer (ALB)

**Purpose:** Distribute traffic to API tasks

**Configuration:**
- Protocol: HTTP (port 80)
- Target Group: ECS services
- Health Checks: Every 30 seconds
- Type: Application Load Balancer

**Future Enhancements:**
- HTTPS/TLS with ACM certificate
- Custom domain name
- WAF (Web Application Firewall)

### 3. ECS Cluster (Elastic Container Service)

**Configuration:**
- Launch Type: Fargate (serverless containers)
- Capacity Providers: FARGATE + FARGATE_SPOT (cost optimization)

**Services:**

#### API Service

| Setting | Value |
|---------|-------|
| Task Definition | form-to-1milion-api |
| Desired Count | 2 (configurable) |
| CPU | 512 vCPU |
| Memory | 1024 MB |
| Container Port | 8080 |
| Health Check | HTTP root path |

**Environment Variables:**
```
DB_HOST=<rds-endpoint>
DB_PORT=5432
DB_NAME=form_to_1milion_db
DB_USER=postgres
DB_PASSWORD=<from SSM>
DB_SSLMODE=require
SQS_QUEUE_URL=<queue-url>
AWS_REGION=us-east-2
```

#### Worker Service

| Setting | Value |
|---------|-------|
| Task Definition | form-to-1milion-worker |
| Desired Count | 2 (configurable) |
| CPU | 256 vCPU |
| Memory | 512 MB |

**Environment Variables:** Same as API

**IAM Permissions:**
- Read/Write access to SQS queue
- Read access to database
- Write access to CloudWatch Logs

### 4. RDS PostgreSQL Database

**Configuration:**

| Setting | Value |
|---------|-------|
| Engine | PostgreSQL 15 |
| Instance Class | db.t3.small (configurable) |
| Storage | 100 GB SSD (configurable) |
| Multi-AZ | Enabled (high availability) |
| Backup Retention | 7 days |
| Encryption | Enabled at-rest |
| Backup Window | Daily 03:00-04:00 UTC |

**Database:**
- Name: `form_to_1milion_db`
- Master Username: `postgres`
- Password: Stored in AWS SSM Parameter Store

**Networking:**
- Private subnet only (no public access)
- Security group restricts to port 5432
- Only ECS tasks can connect

**Backups:**
- Automated daily backups (7-day retention)
- Final snapshot on destruction
- Can restore to any point in time (PITR)

### 5. SQS Queue

**Configuration:**

| Setting | Value |
|---------|-------|
| Queue Name | minha-fila |
| Message Retention | 4 days (345,600 seconds) |
| Visibility Timeout | 300 seconds (5 minutes) |
| Encryption | KMS (managed key) |

**Usage:**
- Producer: API adds tasks to queue
- Consumer: Worker processes tasks
- Auto-delete on successful processing
- Auto-requeue on failure (no manual delete)

### 6. ECR (Elastic Container Registry)

**Repositories:**

| Repository | Purpose | Lifecycle |
|------------|---------|-----------|
| `prod/form-to-1milion-api` | API image | Keep 10 latest |
| `prod/form-to-1milion-worker` | Worker image | Keep 10 latest |

**Image Scanning:** Vulnerability scanning enabled

**Tagging:**
- `latest`: Always points to latest build
- `YYYYMMDD-HHMMSS`: Timestamped for rollback

### 7. CloudWatch Logging

**Log Groups:**

| Log Group | Retention | Purpose |
|-----------|-----------|---------|
| `/ecs/prod/api` | 30 days | API container logs |
| `/ecs/prod/worker` | 30 days | Worker container logs |

**Access:**
```bash
# Real-time logs
aws logs tail /ecs/prod/api --follow
aws logs tail /ecs/prod/worker --follow
```

### 8. IAM Roles and Policies

#### ECS Task Execution Role
- Permissions to pull images from ECR
- Permissions to write logs to CloudWatch
- Permissions to retrieve secrets from SSM/Secrets Manager

#### ECS Task Role
- Permissions to send/receive messages from SQS
- Permissions to read database tables

## Data Flow

### API → SQS → Worker

```
1. User sends request to API
   └─ HTTP POST /user → ALB

2. ALB routes to API task
   └─ ECS runs container

3. API processes request
   └─ Creates task object

4. API sends task to SQS
   └─ `producer.EnqueueTask()`

5. Worker polls SQS
   └─ Long polling (20 seconds)

6. Worker receives task
   └─ SQS message visible for 5 minutes

7. Worker processes task
   └─ Queries/updates database

8. Worker deletes message
   └─ Task complete (fila-vazia)

9. Message removed from queue
   └─ SQS auto-requeue on failure
```

### Database Connection

```
ECS Task (API/Worker)
    ↓
RDS PostgreSQL
    ↓
SSL/TLS Encryption
    ↓
Database Instance
```

## High Availability Design

### Zone Distribution
- Resources spread across 3 availability zones
- Multi-AZ RDS for database high availability
- ALB distributes across zones

### Fault Tolerance
- Multiple API tasks (default 2, configurable)
- Multiple Worker tasks (default 2, configurable)
- NAT gateways in each AZ
- RDS failover replica in different AZ

### Load Balancing
- ALB health checks every 30 seconds
- Failed tasks replaced automatically
- Traffic routed to healthy tasks

## Security Architecture

### Network Security
- VPC isolates resources
- Private subnets for applications
- Security groups restrict traffic
- NACLs (implicit allow-all)

### Data Security
- RDS encryption at-rest
- SQS KMS encryption
- SSL for database connections
- Secrets stored in SSM Parameter Store

### Access Control
- IAM roles for EC2/ECS
- Least-privilege policies
- No hardcoded credentials
- Service-to-service authentication via IAM

### Monitoring & Logging
- CloudWatch logs for all containers
- ECS Container Insights enabled
- Log retention: 30 days

## Scaling Considerations

### Horizontal Scaling
- Change `api_desired_count` and `worker_desired_count` in terraform.tfvars
- Apply with `terraform apply`
- ALB automatically distributes to new tasks

### Vertical Scaling
- Change `api_cpu`, `api_memory` for larger tasks
- Requires task restart
- May increase costs

### Database Scaling
- Change `db_instance_class` for more powerful database
- Change `db_allocated_storage` for more disk space
- Can scale without downtime

### Cost Optimization
- Use FARGATE_SPOT for worker tasks (70% discount)
- Adjust desired counts based on traffic
- Monitor CloudWatch for actual usage

## Disaster Recovery

### RDS Backups
- Automated daily backups (7-day retention)
- Point-in-time recovery (PITR)
- Final snapshot before destruction
- Manual snapshots (create anytime)

### ECR Image Versioning
- Images tagged with timestamp
- Can roll back to previous version
- Lifecycle policy keeps 10 images

### IaC Backup
- Terraform state stored locally (or in S3 backend)
- Version control for all Terraform files
- Can recreate entire infrastructure

## Cost Optimization Tips

1. **Use FARGATE_SPOT for worker tasks**
   - 70% cheaper than FARGATE
   - Replace task if interrupted (2-3 min)
   - Good for batch/async work

2. **Right-size resources**
   - Monitor actual CPU/memory usage
   - Start small, scale up as needed
   - Use CloudWatch Container Insights

3. **Control desired counts**
   - Single task during low traffic
   - Scale up for high traffic periods
   - Use auto-scaling (future enhancement)

4. **Database instance class**
   - Start with db.t3.micro (free tier eligible)
   - db.t3.small for small workloads
   - db.t3.medium for medium workloads

5. **Data transfer**
   - Resources in same region = free
   - Cross-region = $0.02/GB
   - Internet egress = $0.09/GB

## Monitoring & Observability

### Key Metrics

**API Service:**
- Request count
- Response time
- Error rate
- CPU utilization
- Memory utilization

**Worker Service:**
- Tasks processed
- Processing time
- Error rate
- Queue depth
- CPU utilization

**RDS:**
- CPU utilization
- Database connections
- Query performance
- Storage usage

**SQS:**
- Messages sent/received
- Queue depth
- Oldest message age
- Dead letter queue

### Dashboards to Create

1. **Application Dashboard**
   - Request counts
   - Error rates
   - Response times

2. **Operational Dashboard**
   - CPU/Memory utilization
   - Database connections
   - Queue depth

3. **Business Dashboard**
   - Tasks processed
   - User count
   - System health

## Future Enhancements

- [ ] HTTPS with ACM certificate
- [ ] Custom domain with Route 53
- [ ] CloudFront CDN
- [ ] Auto-scaling policies
- [ ] WAF rules
- [ ] VPC Flow Logs
- [ ] X-Ray tracing
- [ ] GuardDuty threat detection
- [ ] Backup automation
- [ ] Disaster recovery plan

---

## Questions?

Refer to `README.md` for detailed setup instructions or `QUICKSTART.md` for quick deployment.
