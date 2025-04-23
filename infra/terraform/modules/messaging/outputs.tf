output "sns_topic_arn" {
  description = "ARN of the SNS topic"
  value       = aws_sns_topic.album_reviews_topic.arn
}

output "sqs_queue_url" {
  description = "URL of the SQS queue"
  value       = aws_sqs_queue.album_reviews_queue.id
}

output "sqs_queue_arn" {
  description = "ARN of the SQS queue"
  value       = aws_sqs_queue.album_reviews_queue.arn
}

output "sqs_queue_name" {
  description = "Name of the SQS queue"
  value       = aws_sqs_queue.album_reviews_queue.name
}

output "dead_letter_queue_url" {
  description = "URL of the dead-letter queue"
  value       = var.create_dead_letter_queue ? aws_sqs_queue.album_reviews_dlq[0].id : ""
}

output "dead_letter_queue_arn" {
  description = "ARN of the dead-letter queue"
  value       = var.create_dead_letter_queue ? aws_sqs_queue.album_reviews_dlq[0].arn : ""
}
