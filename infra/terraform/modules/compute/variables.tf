variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "environment" {
  description = "Environment (e.g., development, production)"
  type        = string
  default     = "development"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
}

variable "app_port" {
  description = "Port on which the application runs"
  type        = number
  default     = 8080
}

variable "app_count" {
  description = "Number of containers to run"
  type        = number
  default     = 2
}

variable "worker_count" {
  description = "Number of worker containers to run"
  type        = number
  default     = 2
}

variable "fargate_cpu" {
  description = "Fargate instance CPU units (1 vCPU = 1024 CPU units)"
  type        = number
  default     = 1024
}

variable "worker_fargate_cpu" {
  description = "Fargate instance CPU units for worker (1 vCPU = 1024 CPU units)"
  type        = number
  default     = 512
}

variable "fargate_memory" {
  description = "Fargate instance memory in MiB"
  type        = number
  default     = 2048
}

variable "worker_fargate_memory" {
  description = "Fargate instance memory in MiB for worker"
  type        = number
  default     = 1024
}

variable "container_image" {
  description = "Docker image for the container"
  type        = string
}

variable "worker_container_image" {
  description = "Docker image for the worker container"
  type        = string
}

variable "ecs_security_group_id" {
  description = "ID of the security group for ECS tasks"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for ECS tasks"
  type        = list(string)
}

variable "alb_target_group_arn" {
  description = "ARN of the ALB target group"
  type        = string
}

variable "s3_bucket_name" {
  description = "Name of the S3 bucket for album covers"
  type        = string
}

variable "dynamodb_table_name" {
  description = "Name of the DynamoDB table for album reviews"
  type        = string
}

variable "sns_topic_arn" {
  description = "ARN of the SNS topic"
  type        = string
}

variable "sqs_queue_url" {
  description = "URL of the SQS queue"
  type        = string
}

variable "lambda_function_name" {
  description = "Name of the Lambda function for image processing"
  type        = string
}

variable "rds_endpoint" {
  description = "Endpoint of the RDS instance"
  type        = string
}

variable "rds_port" {
  description = "Port of the RDS instance"
  type        = string
}

variable "rds_username" {
  description = "Username for the RDS instance"
  type        = string
}

variable "rds_database_name" {
  description = "Name of the database"
  type        = string
}

variable "db_password_arn" {
  description = "ARN of the secret containing the database password"
  type        = string
}
