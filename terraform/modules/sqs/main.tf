resource "aws_sqs_queue" "main" {
  name                       = var.queue_name
  message_retention_seconds  = var.message_retention_seconds
  visibility_timeout_seconds = var.visibility_timeout_seconds
  
  # Encryption
  kms_master_key_id = "alias/aws/sqs"

  tags = {
    Name = "${var.environment}-${var.queue_name}"
  }
}
