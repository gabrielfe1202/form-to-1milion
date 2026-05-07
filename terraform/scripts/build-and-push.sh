#!/bin/bash

# Build and push Docker images to ECR
# Usage: ./build-and-push.sh [environment] [region]

set -e  # Exit on error

ENVIRONMENT=${1:-prod}
REGION=${2:-us-east-2}

echo "=========================================="
echo "Building and pushing Docker images"
echo "Environment: $ENVIRONMENT"
echo "Region: $REGION"
echo "=========================================="

# Get AWS account ID
AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text --region $REGION)
ECR_REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com"

echo ""
echo "AWS Account ID: $AWS_ACCOUNT_ID"
echo "ECR Registry: $ECR_REGISTRY"

# Authenticate Docker with ECR
echo ""
echo "Authenticating Docker with ECR..."
aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin $ECR_REGISTRY

# Build and push API image
echo ""
echo "Building API image..."
API_IMAGE_NAME="${ENVIRONMENT}/form-to-1milion-api"
API_FULL_IMAGE="${ECR_REGISTRY}/${API_IMAGE_NAME}"

docker build \
  -f cmd/api/Dockerfile \
  -t "${API_FULL_IMAGE}:latest" \
  -t "${API_FULL_IMAGE}:$(date +%Y%m%d-%H%M%S)" \
  .

echo "Pushing API image..."
docker push "${API_FULL_IMAGE}:latest"
docker push "${API_FULL_IMAGE}:$(date +%Y%m%d-%H%M%S)"

# Build and push Worker image
echo ""
echo "Building Worker image..."
WORKER_IMAGE_NAME="${ENVIRONMENT}/form-to-1milion-worker"
WORKER_FULL_IMAGE="${ECR_REGISTRY}/${WORKER_IMAGE_NAME}"

docker build \
  -f cmd/worker/Dockerfile \
  -t "${WORKER_FULL_IMAGE}:latest" \
  -t "${WORKER_FULL_IMAGE}:$(date +%Y%m%d-%H%M%S)" \
  .

echo "Pushing Worker image..."
docker push "${WORKER_FULL_IMAGE}:latest"
docker push "${WORKER_FULL_IMAGE}:$(date +%Y%m%d-%H%M%S)"

echo ""
echo "=========================================="
echo "✅ Images built and pushed successfully!"
echo "=========================================="
echo ""
echo "API Image: ${API_FULL_IMAGE}:latest"
echo "Worker Image: ${WORKER_FULL_IMAGE}:latest"
echo ""
