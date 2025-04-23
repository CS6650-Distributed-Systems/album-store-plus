// S3 bucket for album covers
resource "aws_s3_bucket" "album_covers" {
  bucket = var.album_covers_bucket_name

  tags = {
    Name = "${var.project_name}-album-covers"
  }
}

// Block public access to S3 bucket
resource "aws_s3_bucket_public_access_block" "album_covers_access_block" {
  bucket = aws_s3_bucket.album_covers.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

// Enable server-side encryption for the S3 bucket
resource "aws_s3_bucket_server_side_encryption_configuration" "album_covers_encryption" {
  bucket = aws_s3_bucket.album_covers.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

// Configure versioning for the S3 bucket
resource "aws_s3_bucket_versioning" "album_covers_versioning" {
  bucket = aws_s3_bucket.album_covers.id
  versioning_configuration {
    status = "Enabled"
  }
}

// Configure lifecycle rules for the S3 bucket
resource "aws_s3_bucket_lifecycle_configuration" "album_covers_lifecycle" {
  bucket = aws_s3_bucket.album_covers.id

  rule {
    id     = "cleanup-old-processed-images"
    status = "Enabled"

    filter {
      prefix = "albums/*/processed/"
    }

    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }
}

// Configure CORS for the S3 bucket
resource "aws_s3_bucket_cors_configuration" "album_covers_cors" {
  bucket = aws_s3_bucket.album_covers.id

  cors_rule {
    allowed_headers = ["*"]
    allowed_methods = ["GET", "PUT", "POST"]
    allowed_origins = ["*"] # In production, restrict to your domain
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}

// Create ECR repository for the main application
resource "aws_ecr_repository" "app_repository" {
  name                 = var.app_repository_name
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Name = "${var.project_name}-app-ecr"
  }
}

// Create ECR repository for the worker
resource "aws_ecr_repository" "worker_repository" {
  name                 = var.worker_repository_name
  image_tag_mutability = "MUTABLE"

  image_scanning_configuration {
    scan_on_push = true
  }

  tags = {
    Name = "${var.project_name}-worker-ecr"
  }
}

// Create a lifecycle policy for the app repository
resource "aws_ecr_lifecycle_policy" "app_lifecycle_policy" {
  repository = aws_ecr_repository.app_repository.name

  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Keep last 5 images"
      selection = {
        tagStatus     = "any"
        countType     = "imageCountMoreThan"
        countNumber   = 5
      }
      action = {
        type = "expire"
      }
    }]
  })
}

// Create a lifecycle policy for the worker repository
resource "aws_ecr_lifecycle_policy" "worker_lifecycle_policy" {
  repository = aws_ecr_repository.worker_repository.name

  policy = jsonencode({
    rules = [{
      rulePriority = 1
      description  = "Keep last 5 images"
      selection = {
        tagStatus     = "any"
        countType     = "imageCountMoreThan"
        countNumber   = 5
      }
      action = {
        type = "expire"
      }
    }]
  })
}
