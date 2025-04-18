// This file contains the variable values for your deployment

project_name = "album-store"
environment  = "development"
aws_region   = "us-west-2"

// Networking
vpc_cidr           = "10.0.0.0/16"
availability_zones = ["us-west-2a", "us-west-2b"]

// Application
app_port = 8080
app_count = 2

// ECS Fargate
fargate_cpu    = 1024
fargate_memory = 2048
// Worker
worker_count = 2
worker_fargate_cpu = 512
worker_fargate_memory = 1024

// Database - RDS
db_instance_class     = "db.t4g.micro"
db_allocated_storage  = 20
db_name               = "album_store"
db_username           = "album_store_user"
// WARNING: Never commit actual passwords to version control!
// Instead, use environment variables or a secure secret management tool
db_password           = "CS6650_GetRichTeam"
db_multi_az           = false

// Database - DynamoDB
dynamodb_table_name   = "album_reviews"

// Storage - S3
// S3 bucket names must be globally unique
album_covers_bucket_name = "album-store-covers-neu-cs6650"

// Serverless - Lambda
lambda_function_name = "album-image-processor"
lambda_handler       = "index.handler"
lambda_runtime       = "nodejs18.x"
lambda_zip_file      = "../lambda/process_image/nodejs/function.zip"
lambda_memory_size   = 512
lambda_timeout       = 30
max_image_width      = 100
max_image_height     = 100
image_quality        = 85
