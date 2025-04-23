// Create an SNS topic for album reviews
resource "aws_sns_topic" "album_reviews_topic" {
  name = "${var.project_name}-album-reviews"

  tags = {
    Name = "${var.project_name}-album-reviews-topic"
  }
}

// Create an SQS queue for processing album reviews
resource "aws_sqs_queue" "album_reviews_queue" {
  name                      = "${var.project_name}-album-reviews"
  delay_seconds             = 0
  max_message_size          = 262144
  message_retention_seconds = 345600 // 4 days
  receive_wait_time_seconds = 10
  visibility_timeout_seconds = 60

  // Enable SQS server-side encryption
  sqs_managed_sse_enabled = true

  // Configure dead-letter queue if provided
  redrive_policy = var.dead_letter_queue_arn != "" ? jsonencode({
    deadLetterTargetArn = var.dead_letter_queue_arn
    maxReceiveCount     = 5
  }) : null

  tags = {
    Name = "${var.project_name}-album-reviews-queue"
  }
}

// Create a policy for the SQS queue to allow SNS to send messages
resource "aws_sqs_queue_policy" "album_reviews_queue_policy" {
  queue_url = aws_sqs_queue.album_reviews_queue.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action   = "sqs:SendMessage"
        Resource = aws_sqs_queue.album_reviews_queue.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_sns_topic.album_reviews_topic.arn
          }
        }
      }
    ]
  })
}

// Subscribe the SQS queue to the SNS topic
resource "aws_sns_topic_subscription" "album_reviews_subscription" {
  topic_arn = aws_sns_topic.album_reviews_topic.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.album_reviews_queue.arn
}

// Optional: Create a dead-letter queue for failed messages
resource "aws_sqs_queue" "album_reviews_dlq" {
  count = var.create_dead_letter_queue ? 1 : 0

  name                      = "${var.project_name}-album-reviews-dlq"
  message_retention_seconds = 1209600 // 14 days

  // Enable SQS server-side encryption
  sqs_managed_sse_enabled = true

  tags = {
    Name = "${var.project_name}-album-reviews-dlq"
  }
}
