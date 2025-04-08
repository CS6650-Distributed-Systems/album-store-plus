variable "project_name" {
  description = "Name of the project, used as a prefix for resource names"
  type        = string
  default     = "album-store"
}

variable "environment" {
  description = "Environment name (development, staging, production)"
  type        = string
  default     = "development"
}

variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

variable "use_localstack" {
  description = "Whether to use LocalStack for local development"
  type        = bool
  default     = false
}

variable "localstack_endpoint" {
  description = "LocalStack endpoint URL"
  type        = string
  default     = "http://localhost:4566"
}

# VPC Configuration
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones to use"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b"]
}

# S3 Configuration
variable "s3_bucket_name" {
  description = "Name of the S3 bucket for album images"
  type        = string
  default     = "album-store-images"
}

# DynamoDB Configuration
variable "dynamodb_table_name" {
  description = "Name of the DynamoDB table for albums"
  type        = string
  default     = "albums"
}

variable "dynamodb_read_capacity" {
  description = "Read capacity for DynamoDB table"
  type        = number
  default     = 5
}

variable "dynamodb_write_capacity" {
  description = "Write capacity for DynamoDB table"
  type        = number
  default     = 5
}

# RDS Configuration
variable "rds_instance_class" {
  description = "Instance class for RDS"
  type        = string
  default     = "db.t3.micro"
}

variable "rds_allocated_storage" {
  description = "Allocated storage for RDS in GB"
  type        = number
  default     = 20
}

variable "rds_db_name" {
  description = "Database name for RDS"
  type        = string
  default     = "albumstore"
}

variable "rds_username" {
  description = "Username for RDS"
  type        = string
  default     = "albumuser"
  sensitive   = true
}

variable "rds_password" {
  description = "Password for RDS"
  type        = string
  default     = "albumpass"  # This is for demonstration only, use AWS Secrets Manager or similar in production
  sensitive   = true
}

# SNS Configuration
variable "sns_topic_name" {
  description = "Name of the SNS topic for album events"
  type        = string
  default     = "album-events"
}

# SQS Configuration
variable "sqs_queue_name" {
  description = "Name of the SQS queue for album events"
  type        = string
  default     = "album-events-queue"
}

# Lambda Configuration
variable "lambda_function_name" {
  description = "Name of the Lambda function for image processing"
  type        = string
  default     = "album-image-processor"
}