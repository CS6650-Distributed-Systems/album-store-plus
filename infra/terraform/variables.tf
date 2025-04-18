// General
variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "album-store"
}

variable "environment" {
  description = "Environment (e.g., development, production)"
  type        = string
  default     = "development"
}

variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-west-2"
}

// Networking
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "List of availability zones to use"
  type        = list(string)
  default     = ["us-west-2a", "us-west-2b"]
}

// Application
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

// ECS Fargate
variable "fargate_cpu" {
  description = "Fargate instance CPU units (1 vCPU = 1024 CPU units)"
  type        = number
  default     = 1024
}

variable "worker_fargate_cpu" {
  description = "Fargate instance CPU units for worker"
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

# variable "container_image" {
#   description = "Docker image for the container"
#   type        = string
#   default     = "album-store-plus:latest"
# }

# variable "worker_container_image" {
#   description = "Docker image for the worker container"
#   type        = string
#   default     = "album-store-worker:latest"
# }

// Database - RDS
variable "db_instance_class" {
  description = "Instance class for the RDS instance"
  type        = string
  default     = "db.t4g.micro"
}

variable "db_allocated_storage" {
  description = "Allocated storage for the RDS instance in GB"
  type        = number
  default     = 20
}

variable "db_name" {
  description = "Name of the database"
  type        = string
  default     = "album_store"
}

variable "db_username" {
  description = "Username for the database"
  type        = string
  default     = "album_store_user"
}

variable "db_password" {
  description = "Password for the database"
  type        = string
  sensitive   = true
}

variable "db_multi_az" {
  description = "Whether to enable Multi-AZ for the RDS instance"
  type        = bool
  default     = false
}

// Database - DynamoDB
variable "dynamodb_table_name" {
  description = "Name of the DynamoDB table for album reviews"
  type        = string
  default     = "album_reviews"
}

// Storage - S3
variable "album_covers_bucket_name" {
  description = "Name of the S3 bucket for album covers"
  type        = string
  default     = "album-store-covers"
}

// Serverless - Lambda
variable "lambda_function_name" {
  description = "Name of the Lambda function"
  type        = string
  default     = "album-image-processor"
}

variable "lambda_handler" {
  description = "Handler for the Lambda function"
  type        = string
  default     = "index.handler"
}

variable "lambda_runtime" {
  description = "Runtime for the Lambda function"
  type        = string
  default     = "nodejs18.x"
}

variable "lambda_zip_file" {
  description = "Path to the Lambda function zip file"
  type        = string
  default     = "../infra/lambda/process_image/nodejs/function.zip"
}

variable "lambda_memory_size" {
  description = "Memory size for the Lambda function in MB"
  type        = number
  default     = 512
}

variable "lambda_timeout" {
  description = "Timeout for the Lambda function in seconds"
  type        = number
  default     = 30
}

variable "max_image_width" {
  description = "Maximum width for processed images"
  type        = number
  default     = 100
}

variable "max_image_height" {
  description = "Maximum height for processed images"
  type        = number
  default     = 100
}

variable "image_quality" {
  description = "Quality for processed JPEG images (1-100)"
  type        = number
  default     = 85
}
