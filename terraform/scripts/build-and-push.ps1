# Build and push Docker images to ECR
# Usage: .\build-and-push.ps1 -Environment prod -Region us-east-2

param(
    [string]$Environment = "prod",
    [string]$Region = "us-east-2"
)

Write-Host "==========================================" -ForegroundColor Green
Write-Host "Building and pushing Docker images" -ForegroundColor Green
Write-Host "Environment: $Environment" -ForegroundColor Green
Write-Host "Region: $Region" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green

# Get AWS account ID
$AccountId = (aws sts get-caller-identity --query Account --output text --region $Region)
$EcrRegistry = "$AccountId.dkr.ecr.$Region.amazonaws.com"

Write-Host ""
Write-Host "AWS Account ID: $AccountId" -ForegroundColor Cyan
Write-Host "ECR Registry: $EcrRegistry" -ForegroundColor Cyan

# Change to project root directory
$ScriptPath = $PSCommandPath
$ScriptsDir = Split-Path -Parent $ScriptPath
$TerraformDir = Split-Path -Parent $ScriptsDir
$ProjectRoot = Split-Path -Parent $TerraformDir
Write-Host "Project root: $ProjectRoot" -ForegroundColor Cyan
Push-Location $ProjectRoot

# Authenticate Docker with ECR
Write-Host ""
Write-Host "Authenticating Docker with ECR..." -ForegroundColor Yellow
aws ecr get-login-password --region $Region | docker login --username AWS --password-stdin $EcrRegistry

# Build and push API image
Write-Host ""
Write-Host "Building API image..." -ForegroundColor Yellow
$ApiImageName = "$Environment/form-to-1milion-api"
$ApiFullImage = "$EcrRegistry/$ApiImageName"
$Timestamp = (Get-Date -Format "yyyyMMdd-HHmmss")

docker build `
  -f cmd/api/Dockerfile `
  -t "$ApiFullImage`:latest" `
  -t "$ApiFullImage`:$Timestamp" `
  .

Write-Host "Pushing API image..." -ForegroundColor Yellow
docker push "$ApiFullImage`:latest"
docker push "$ApiFullImage`:$Timestamp"

# Build and push Worker image
Write-Host ""
Write-Host "Building Worker image..." -ForegroundColor Yellow
$WorkerImageName = "$Environment/form-to-1milion-worker"
$WorkerFullImage = "$EcrRegistry/$WorkerImageName"

docker build `
  -f cmd/worker/Dockerfile `
  -t "$WorkerFullImage`:latest" `
  -t "$WorkerFullImage`:$Timestamp" `
  .

Write-Host "Pushing Worker image..." -ForegroundColor Yellow
docker push "$WorkerFullImage`:latest"
docker push "$WorkerFullImage`:$Timestamp"

Write-Host ""
Write-Host "==========================================" -ForegroundColor Green
Write-Host "Images built and pushed successfully!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host ""
Write-Host "API Image: $ApiFullImage`:latest" -ForegroundColor Cyan
Write-Host "Worker Image: $WorkerFullImage`:latest" -ForegroundColor Cyan
Write-Host ""

# Return to original directory
Pop-Location
