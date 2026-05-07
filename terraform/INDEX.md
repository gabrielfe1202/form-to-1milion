🚀 **Welcome to form-to-1milion AWS Infrastructure**

This directory contains complete Terraform configuration to deploy your Go application to AWS with SQS and PostgreSQL.

---

## 📚 Documentation

Start with these in order:

1. **[QUICKSTART.md](QUICKSTART.md)** ⭐ - START HERE (30 min deployment)
   - Prerequisites checklist
   - Step-by-step deployment
   - Verification commands

2. **[README.md](README.md)** - Complete Setup Guide
   - Detailed instructions
   - Configuration options
   - Troubleshooting basics

3. **[ARCHITECTURE.md](ARCHITECTURE.md)** - Infrastructure Design
   - Architecture diagram
   - Component details
   - Security design
   - Scaling considerations

4. **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** - Common Issues
   - Setup problems
   - Terraform errors
   - ECS issues
   - Database problems

---

## 🗂️ File Structure

```
terraform/
├── main.tf                      # Main configuration
├── variables.tf                 # Variable definitions
├── outputs.tf                   # Output values
├── terraform.tfvars.example     # Copy & customize this
├── .gitignore                   # Git configuration
│
├── modules/
│   ├── vpc/                     # Network infrastructure
│   ├── rds/                     # PostgreSQL database
│   ├── sqs/                     # Message queue
│   ├── ecr/                     # Docker registries
│   └── ecs/                     # Container orchestration
│
├── scripts/
│   ├── build-and-push.sh        # Build images (Linux/Mac)
│   └── build-and-push.ps1       # Build images (Windows)
│
├── Makefile                     # Helper commands
├── README.md                    # Full documentation
├── QUICKSTART.md                # Quick deployment
├── ARCHITECTURE.md              # Infrastructure design
├── TROUBLESHOOTING.md           # Issue solutions
└── INDEX.md                     # This file
```

---

## ⚡ Quick Start (TL;DR)

```bash
# 1. Store database password
aws ssm put-parameter --name "/prod/db/password" --value "your-password" --type "SecureString" --region us-east-2

# 2. Copy configuration
cp terraform.tfvars.example terraform.tfvars

# 3. Build and push Docker images (PowerShell on Windows)
.\scripts\build-and-push.ps1 -Environment prod -Region us-east-2

# OR on Mac/Linux
bash scripts/build-and-push.sh prod us-east-2

# 4. Deploy infrastructure
terraform init
terraform plan -out=tfplan
terraform apply tfplan

# 5. Get API URL
terraform output api_load_balancer_dns

# 6. Test it!
curl http://<api-url>
```

---

## 🔑 Key Information

### Deployment Checklist

- [ ] AWS Account with permissions
- [ ] AWS CLI configured: `aws --version`
- [ ] Terraform installed: `terraform --version`
- [ ] Docker running: `docker --version`
- [ ] Database password stored in SSM
- [ ] Docker images built and pushed to ECR
- [ ] `terraform.tfvars` created and customized
- [ ] Infrastructure deployed with `terraform apply`

### Resources Created

| Resource | Details |
|----------|---------|
| **VPC** | 10.0.0.0/16 across 3 availability zones |
| **Subnets** | 3 public + 3 private for high availability |
| **RDS** | PostgreSQL 15 (Multi-AZ, encrypted) |
| **SQS** | Queue named "minha-fila" |
| **ECS** | Fargate cluster with API + Worker services |
| **ALB** | Distributes traffic to API tasks |
| **ECR** | Docker image repositories |
| **Logs** | CloudWatch logs for monitoring |

### Estimated Monthly Cost

~$136/month on `us-east-2`

See [ARCHITECTURE.md](ARCHITECTURE.md#cost-optimization-tips) for cost optimization.

---

## 🚀 Deployment Process

### Phase 1: Preparation (5 min)
1. Store database password in SSM
2. Create `terraform.tfvars` from example

### Phase 2: Building (5-10 min)
3. Build Docker images with `build-and-push.ps1`
4. Push to ECR

### Phase 3: Infrastructure (10-15 min)
5. Run `terraform init`
6. Run `terraform plan`
7. Run `terraform apply`
8. RDS typically takes 5-10 minutes

### Phase 4: Verification (5 min)
9. Get ALB URL and test
10. Check CloudWatch logs

**Total Time: ~30 minutes**

---

## 📝 First-Time Commands

```powershell
# Navigate to terraform directory
cd terraform

# View plan without applying
terraform plan

# See what will be created (recommended first step)
terraform plan -out=tfplan

# Review the tfplan file
# Then apply when ready:
terraform apply tfplan

# View all outputs
terraform output

# Get specific output
terraform output api_load_balancer_dns

# View CloudWatch logs (real-time)
aws logs tail /ecs/prod/api --follow
aws logs tail /ecs/prod/worker --follow

# Check ECS service status
aws ecs describe-services --cluster form-to-1milion --services form-to-1milion-api-service

# Destroy when done (WARNING: deletes database)
terraform destroy
```

---

## ⚠️ Important Notes

### Before You Deploy

1. **Database Password**: Store in AWS SSM Parameter Store (NOT in tfvars)
   ```bash
   aws ssm put-parameter --name "/prod/db/password" --value "YourPassword123!" --type "SecureString"
   ```

2. **Docker Images**: Must be pushed to ECR before `terraform apply`
   ```bash
   .\scripts\build-and-push.ps1
   ```

3. **AWS Credentials**: Configure with `aws configure`

4. **Review Plan**: Always run `terraform plan` before `apply`

### Security Best Practices

- ✅ Database password in SSM (secure parameter store)
- ✅ TLS encryption for database connections
- ✅ Private subnets for applications
- ✅ Security groups restrict access
- ⚠️ Add HTTPS to ALB after deployment
- ⚠️ Enable WAF for production

### Monitoring

After deployment, monitor:

```bash
# Check application health
terraform output api_load_balancer_dns
curl <dns-name>

# View logs
aws logs tail /ecs/prod/api --follow

# Check services
aws ecs list-services --cluster form-to-1milion

# Monitor queue
aws sqs get-queue-attributes --queue-url <queue-url> --attribute-names ApproximateNumberOfMessages
```

---

## 🆘 Need Help?

1. **Setup Issues?** → Read [QUICKSTART.md](QUICKSTART.md)
2. **Something Broken?** → Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
3. **Understand Design?** → Review [ARCHITECTURE.md](ARCHITECTURE.md)
4. **Specific Configuration?** → See [README.md](README.md)

---

## 🔄 Common Operations

```bash
# Rebuild Docker images and redeploy
.\scripts\build-and-push.ps1
aws ecs update-service --cluster form-to-1milion --service form-to-1milion-api-service --force-new-deployment

# Scale services
# Edit terraform.tfvars, change api_desired_count or worker_desired_count
terraform apply

# View outputs again
terraform output

# Update environment variables
# Edit terraform/modules/ecs/main.tf
terraform apply

# Stop all services (keeps infrastructure)
aws ecs update-service --cluster form-to-1milion --service form-to-1milion-api-service --desired-count 0
aws ecs update-service --cluster form-to-1milion --service form-to-1milion-worker-service --desired-count 0

# Delete everything
terraform destroy  # WARNING: Deletes database!
```

---

## 📞 Support Resources

- Terraform Docs: https://www.terraform.io/docs
- AWS Docs: https://docs.aws.amazon.com
- ECS Best Practices: https://docs.aws.amazon.com/AmazonECS/latest/developerguide/
- AWS Support: https://console.aws.amazon.com/support

---

## ✅ Validation Checklist

After deployment, verify:

- [ ] Terraform apply completed without errors
- [ ] API service is running (check ECS)
- [ ] Worker service is running (check ECS)
- [ ] ALB is healthy
- [ ] Can reach API via `curl http://<alb-dns>`
- [ ] Database is running (check RDS)
- [ ] SQS queue exists
- [ ] CloudWatch logs showing service activity
- [ ] Messages can be sent via API
- [ ] Worker processes messages from queue

---

## 🎯 Next Steps (After Deployment)

1. **Add HTTPS**: Upload ACM certificate to ALB
2. **Custom Domain**: Point DNS to ALB
3. **Auto-scaling**: Set up CloudWatch alarms
4. **Backups**: Configure RDS backup strategy
5. **Monitoring**: Create dashboards
6. **CI/CD**: Automate image builds and deployments

---

**Last Updated**: April 1, 2026

**Ready to deploy?** → Start with [QUICKSTART.md](QUICKSTART.md) 🚀
