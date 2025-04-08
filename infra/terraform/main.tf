provider "aws" {
  region = var.aws_region

  # Use these settings for local development with LocalStack
  dynamic "endpoints" {
    for_each = var.use_localstack ? [1] : []
    content {
      apigateway     = var.localstack_endpoint
      cloudformation = var.localstack_endpoint
      cloudwatch     = var.localstack_endpoint
      dynamodb       = var.localstack_endpoint
      es             = var.localstack_endpoint
      firehose       = var.localstack_endpoint
      iam            = var.localstack_endpoint
      kinesis        = var.localstack_endpoint
      lambda         = var.localstack_endpoint
      rds            = var.localstack_endpoint
      s3             = var.localstack_endpoint
      secretsmanager = var.localstack_endpoint
      ses            = var.localstack_endpoint
      sns            = var.localstack_endpoint
      sqs            = var.localstack_endpoint
      ssm            = var.localstack_endpoint
    }
  }

  # Skip credentials validation and metadata API check for LocalStack
  skip_credentials_validation = var.use_localstack
  skip_metadata_api_check     = var.use_localstack
  skip_requesting_account_id  = var.use_localstack

  # S3 specific settings for LocalStack
  s3_use_path_style           = var.use_localstack
}

# VPC Configuration
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "${var.project_name}-vpc"
  }
}

# Public Subnets
resource "aws_subnet" "public" {
  count             = length(var.availability_zones)
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index)
  availability_zone = element(var.availability_zones, count.index)

  tags = {
    Name = "${var.project_name}-public-subnet-${count.index + 1}"
  }
}

# Private Subnets
resource "aws_subnet" "private" {
  count             = length(var.availability_zones)
  vpc_id            = aws_vpc.main.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index + length(var.availability_zones))
  availability_zone = element(var.availability_zones, count.index)

  tags = {
    Name = "${var.project_name}-private-subnet-${count.index + 1}"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name = "${var.project_name}-igw"
  }
}

# Route Table for Public Subnets
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name = "${var.project_name}-public-route-table"
  }
}

# Route Table Association for Public Subnets
resource "aws_route_table_association" "public" {
  count          = length(var.availability_zones)
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# S3 Bucket for Album Images
resource "aws_s3_bucket" "album_images" {
  bucket = var.s3_bucket_name
  force_destroy = true  # Allow deletion of non-empty bucket for demo purposes

  tags = {
    Name        = "${var.project_name}-images"
    Environment = var.environment
  }
}

# S3 Bucket Public Access Block (Restricting public access)
resource "aws_s3_bucket_public_access_block" "album_images" {
  bucket = aws_s3_bucket.album_images.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# S3 Bucket CORS Configuration
resource "aws_s3_bucket_cors_configuration" "album_images" {
  bucket = aws_s3_bucket.album_images.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST", "DELETE", "HEAD"]
    allowed_origins = ["*"]  # Restrict this in production
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}

# DynamoDB Table for Albums
resource "aws_dynamodb_table" "albums" {
  name         = var.dynamodb_table_name
  billing_mode = "PROVISIONED"
  hash_key     = "id"
  
  read_capacity  = var.dynamodb_read_capacity
  write_capacity = var.dynamodb_write_capacity

  attribute {
    name = "id"
    type = "S"
  }

  attribute {
    name = "artist"
    type = "S"
  }

  attribute {
    name = "year"
    type = "S"
  }

  global_secondary_index {
    name               = "ArtistIndex"
    hash_key           = "artist"
    projection_type    = "ALL"
    read_capacity      = var.dynamodb_read_capacity
    write_capacity     = var.dynamodb_write_capacity
  }

  global_secondary_index {
    name               = "YearIndex"
    hash_key           = "year"
    projection_type    = "ALL"
    read_capacity      = var.dynamodb_read_capacity
    write_capacity     = var.dynamodb_write_capacity
  }

  tags = {
    Name        = "${var.project_name}-albums"
    Environment = var.environment
  }
}

# RDS MySQL Instance
resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-db-subnet-group"
  subnet_ids = aws_subnet.private.*.id

  tags = {
    Name = "${var.project_name}-db-subnet-group"
  }
}

resource "aws_security_group" "rds" {
  name        = "${var.project_name}-rds-sg"
  description = "Security group for RDS MySQL instance"
  vpc_id      = aws_vpc.main.id

  ingress {
    from_port   = 3306
    to_port     = 3306
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.project_name}-rds-sg"
  }
}

resource "aws_db_instance" "main" {
  identifier             = "${var.project_name}-db"
  allocated_storage      = var.rds_allocated_storage
  storage_type           = "gp2"
  engine                 = "mysql"
  engine_version         = "8.0"
  instance_class         = var.rds_instance_class
  db_name                = var.rds_db_name
  username               = var.rds_username
  password               = var.rds_password
  parameter_group_name   = "default.mysql8.0"
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  skip_final_snapshot    = true
  publicly_accessible    = false
  multi_az               = var.environment == "production"

  tags = {
    Name        = "${var.project_name}-db"
    Environment = var.environment
  }
}

# SNS Topic for Event Bus
resource "aws_sns_topic" "album_events" {
  name = var.sns_topic_name

  tags = {
    Name        = "${var.project_name}-events"
    Environment = var.environment
  }
}

# SQS Queue for Event Processing
resource "aws_sqs_queue" "album_events" {
  name                      = var.sqs_queue_name
  delay_seconds             = 0
  max_message_size          = 262144
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10

  tags = {
    Name        = "${var.project_name}-events-queue"
    Environment = var.environment
  }
}

# SQS Queue Policy for SNS Subscription
resource "aws_sqs_queue_policy" "album_events" {
  queue_url = aws_sqs_queue.album_events.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action = "sqs:SendMessage"
        Resource = aws_sqs_queue.album_events.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_sns_topic.album_events.arn
          }
        }
      }
    ]
  })
}

# SNS Subscription for SQS
resource "aws_sns_topic_subscription" "album_events" {
  topic_arn = aws_sns_topic.album_events.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.album_events.arn
}

# IAM Role for Lambda
resource "aws_iam_role" "lambda_role" {
  name = "${var.project_name}-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      },
    ]
  })

  tags = {
    Name        = "${var.project_name}-lambda-role"
    Environment = var.environment
  }
}

# IAM Policy for Lambda
resource "aws_iam_policy" "lambda_policy" {
  name        = "${var.project_name}-lambda-policy"
  description = "Policy for Lambda function to access S3, SQS, and CloudWatch Logs"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Effect   = "Allow"
        Resource = "arn:aws:logs:*:*:*"
      },
      {
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Effect   = "Allow"
        Resource = [
          aws_s3_bucket.album_images.arn,
          "${aws_s3_bucket.album_images.arn}/*"
        ]
      },
      {
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes"
        ]
        Effect   = "Allow"
        Resource = aws_sqs_queue.album_events.arn
      }
    ]
  })
}

# Attach IAM Policy to Lambda Role
resource "aws_iam_role_policy_attachment" "lambda_policy_attachment" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = aws_iam_policy.lambda_policy.arn
}

# Lambda Function for Image Processing
resource "aws_lambda_function" "image_processor" {
  function_name    = var.lambda_function_name
  role             = aws_iam_role.lambda_role.arn
  handler          = "imageprocessor"
  runtime          = "go1.x"
  filename         = "../bin/imageprocessor.zip"
  source_code_hash = filebase64sha256("../bin/imageprocessor.zip")
  memory_size      = 256
  timeout          = 30

  environment {
    variables = {
      S3_BUCKET_NAME = aws_s3_bucket.album_images.bucket
    }
  }

  tags = {
    Name        = "${var.project_name}-image-processor"
    Environment = var.environment
  }
}

# Event Source Mapping for Lambda
resource "aws_lambda_event_source_mapping" "image_processor" {
  event_source_arn = aws_sqs_queue.album_events.arn
  function_name    = aws_lambda_function.image_processor.arn
  batch_size       = 10
}