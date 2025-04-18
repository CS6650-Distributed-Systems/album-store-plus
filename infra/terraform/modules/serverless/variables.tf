variable "project_name" {
  description = "Name of the project"
  type        = string
}

variable "iam_role_name" {
  description = "Name of the IAM role to use for Lambda functions"
  type        = string
}

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
  default     = "nodejs18.x" // Using Node.js for better image processing performance
}

variable "lambda_zip_file" {
  description = "Path to the Lambda function zip file"
  type        = string
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

variable "s3_bucket_name" {
  description = "Name of the S3 bucket for album covers"
  type        = string
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
