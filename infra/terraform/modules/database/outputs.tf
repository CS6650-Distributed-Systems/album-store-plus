output "rds_endpoint" {
  description = "Endpoint of the RDS instance"
  value       = aws_db_instance.album_store_mysql.endpoint
}

output "rds_port" {
  description = "Port of the RDS instance"
  value       = aws_db_instance.album_store_mysql.port
}

output "rds_username" {
  description = "Username for the RDS instance"
  value       = aws_db_instance.album_store_mysql.username
}

output "rds_password" {
  description = "Password for the RDS instance"
  value       = aws_db_instance.album_store_mysql.password
}

output "rds_database_name" {
  description = "Name of the database"
  value       = aws_db_instance.album_store_mysql.db_name
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  value       = aws_dynamodb_table.album_reviews.name
}

output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table"
  value       = aws_dynamodb_table.album_reviews.arn
}
