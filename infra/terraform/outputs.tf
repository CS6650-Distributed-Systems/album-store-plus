// Networking outputs
output "vpc_id" {
  description = "The ID of the VPC"
  value       = module.networking.vpc_id
}

output "alb_dns_name" {
  description = "The DNS name of the ALB"
  value       = module.networking.alb_dns_name
}

// Database outputs
output "rds_endpoint" {
  description = "Endpoint of the RDS instance"
  value       = module.database.rds_endpoint
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  value       = module.database.dynamodb_table_name
}

// Storage outputs
output "album_covers_bucket_name" {
  description = "Name of the S3 bucket for album covers"
  value       = module.storage.album_covers_bucket_name
}

output "app_repository_url" {
  description = "URL of the ECR repository for the main application"
  value       = module.storage.app_repository_url
}

output "worker_repository_url" {
  description = "URL of the ECR repository for the worker"
  value       = module.storage.worker_repository_url
}

// Serverless outputs
output "lambda_function_name" {
  description = "Name of the Lambda function"
  value       = module.serverless.lambda_function_name
}

// Messaging outputs
output "sns_topic_arn" {
  description = "ARN of the SNS topic"
  value       = module.messaging.sns_topic_arn
}

output "sqs_queue_url" {
  description = "URL of the SQS queue"
  value       = module.messaging.sqs_queue_url
}

// Compute outputs
output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = module.compute.ecs_cluster_name
}

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = module.compute.ecs_service_name
}

output "worker_service_name" {
  description = "Name of the worker ECS service"
  value       = module.compute.worker_service_name
}

// Application URL
output "application_url" {
  description = "URL of the application"
  value       = "http://${module.networking.alb_dns_name}"
}
