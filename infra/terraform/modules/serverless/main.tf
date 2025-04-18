// Reference the existing IAM role for the Lambda function
data "aws_iam_role" "lambda_role" {
  name = "RoleForLambdaModLabRole"
}

// Create a CloudWatch log group for Lambda logs
resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${var.lambda_function_name}"
  retention_in_days = 30

  tags = {
    Name = "${var.project_name}-lambda-logs"
  }
}

// Create the Lambda function
resource "aws_lambda_function" "image_processor" {
  function_name    = var.lambda_function_name
  role             = data.aws_iam_role.lambda_role.arn
  handler          = var.lambda_handler
  runtime          = var.lambda_runtime
  filename         = var.lambda_zip_file
  source_code_hash = filebase64sha256(var.lambda_zip_file)
  memory_size      = var.lambda_memory_size
  timeout          = var.lambda_timeout

  environment {
    variables = {
      S3_BUCKET = var.s3_bucket_name
      MAX_WIDTH = tostring(var.max_image_width)
      MAX_HEIGHT = tostring(var.max_image_height)
      QUALITY = tostring(var.image_quality)
    }
  }

  depends_on = [
    aws_cloudwatch_log_group.lambda_logs,
  ]

  tags = {
    Name = "${var.project_name}-image-processor"
  }
}
