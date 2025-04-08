output "vpc_id" {
  description = "ID of the VPC"
  value       = aws_vpc.main.id
}

output "public_subnet_ids" {
  description = "IDs of the public subnets"
  value       = aws_subnet.public.*.id
}

output "private_subnet_ids" {
  description = "IDs of the private subnets"
  value       = aws_subnet.private.*.id
}

output "s3_bucket_name" {
  description = "Name of the S3 bucket for album images"
  value       = aws_s3_bucket.album_images.bucket
}

output "s3_bucket_arn" {
  description = "ARN of the S3 bucket for album images"
  value       = aws_s3_bucket.album_images.arn
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table for albums"
  value       = aws_dynamodb_table.albums.name
}

output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table for albums"
  value       = aws_dynamodb_table.albums.arn
}

output "rds_endpoint" {
  description = "Endpoint of the RDS instance"
  value       = aws_db_instance.main.endpoint
}

output "rds_address" {
  description = "Address of the RDS instance"
  value       = aws_db_instance.main.address
}

output "rds_db_name" {
  description = "Database name of the RDS instance"
  value       = aws_db_instance.main.db_name
}

output "sns_topic_arn" {
  description = "ARN of the SNS topic for album events"
  value       = aws_sns_topic.album_events.arn
}

output "sqs_queue_url" {
  description = "URL of the SQS queue for album events"
  value       = aws_sqs_queue.album_events.id
}

output "sqs_queue_arn" {
  description = "ARN of the SQS queue for album events"
  value       = aws_sqs_queue.album_events.arn
}

output "lambda_function_name" {
  description = "Name of the Lambda function for image processing"
  value       = aws_lambda_function.image_processor.function_name
}

output "lambda_function_arn" {
  description = "ARN of the Lambda function for image processing"
  value       = aws_lambda_function.image_processor.arn
}